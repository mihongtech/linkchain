package rpcserver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/linkchain/client/explorer/rpc/rpcjson"
	"github.com/linkchain/common/util/log"
)

const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the
	// RPC server is allowed to stay open without authenticating before it
	// is closed.
	rpcAuthTimeoutSeconds = 10
	listenPort            = ":8083"

	RPCMaxClients = 60
	RPCQuirks     = true
)

type Server struct {
	numClients int32

	config Config

	statusLines map[int]string
	statusLock  sync.RWMutex

	//quit channel
	requestProcessShutdown chan struct{}
}

// rpcserverConfig is a descriptor containing the RPC server configuration.
type Config struct {
	// StartupTime is the unix timestamp for when the server that is hosting
	// the RPC server started.
	StartupTime int64
}

// newRPCServer returns a new instance of the rpcServer struct.
func NewRPCServer(cfg *Config) (*Server, error) {
	rpc := Server{
		config:                 *cfg,
		statusLines:            make(map[int]string),
		requestProcessShutdown: make(chan struct{}),
	}

	return &rpc, nil
}

// Start is used by rpcserver.go to start the rpcserver listener.
func (s *Server) Start() {
	log.Info("Starting RPC rpcserver")
	rpcServeMux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    listenPort,
		Handler: rpcServeMux,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}

	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		r.Close = true

		// Limit the number of connections to max allowed.
		if s.limitConnections(w, r.RemoteAddr) {
			return
		}

		// Keep track of the number of connected clients.
		s.incrementClients()
		defer s.decrementClients()

		// Read and respond to the request.
		s.jsonRPCRead(w, r)
	})

	httpServer.ListenAndServe()
}

func (s *Server) Stop() bool {
	s.requestProcessShutdown <- struct{}{}
	return true
}

// jsonRPCRead handles reading and responding to RPC messages.
func (s *Server) jsonRPCRead(w http.ResponseWriter, r *http.Request) {
	// Read and close the JSON-RPC request body from the caller.
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	log.Info("rpcServer", "raw", string(body))
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON message: %v",
			errCode, err), errCode)
		return
	}

	// Unfortunately, the httpclient server doesn't provide the ability to
	// change the read deadline for the new connection and having one breaks
	// long polling.  However, not having a read deadline on the initial
	// connection would mean clients can connect and idle forever.  Thus,
	// hijack the connecton from the HTTP server, clear the read deadline,
	// and handle writing the response manually.
	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		log.Warn(errMsg)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		log.Warn("Failed to hijack HTTP connection: %v", err)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+err.Error(), errCode)
		return
	}
	defer conn.Close()
	defer buf.Flush()

	// Attempt to parse the raw body into a JSON-RPC request.
	var responseID interface{}
	var jsonErr error
	var result interface{}
	var request rpcjson.Request
	if err := json.Unmarshal(body, &request); err != nil {
		jsonErr = &rpcjson.RPCError{
			Code:    rpcjson.ErrRPCParse.Code,
			Message: "Failed to parse request: " + err.Error(),
		}
	}

	if jsonErr == nil {
		if request.ID == nil && !(RPCQuirks && request.Jsonrpc == "") {
			return
		}

		// The parse was at least successful enough to have an ID so
		// set it for the response.
		responseID = request.ID

		// Setup a close notifier.  Since the connection is hijacked,
		// the CloseNotifer on the ResponseWriter is not available.
		closeChan := make(chan struct{}, 1)
		go func() {
			_, err := conn.Read(make([]byte, 1))
			if err != nil {
				close(closeChan)
			}
		}()

		if jsonErr == nil {
			// Attempt to parse the JSON-RPC request into a known concrete
			// command.
			parsedCmd, err := parseCmd(&request)
			if err != nil {
				jsonErr = err
			} else {
				result, jsonErr = s.standardCmdResult(request.Method, parsedCmd, closeChan)
			}
		}
	}

	// Marshal the response.
	msg, err := createMarshalledReply(responseID, result, jsonErr)
	fmt.Println(string(msg))
	if err != nil {
		log.Error("Failed to marshal reply: %v", err)
		return
	}

	// Write the response.
	err = s.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
	if err != nil {
		log.Error("rpc", "rpcResponse", err)
		return
	}
	if _, err := buf.Write(msg); err != nil {
		log.Error("rpc", "Failed to write marshalled reply: %v", err)
	}

	// Terminate with newline to maintain compatibility
	if err := buf.WriteByte('\n'); err != nil {
		log.Error("rpc", "Failed to append terminating newline to reply: %v", err)
	}
}

// createMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert errors that are not of
// the type *btcjson.RPCError to the appropriate type as needed.
func createMarshalledReply(id, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *rpcjson.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*rpcjson.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = internalRPCError(replyErr.Error(), "")
		}
	}

	return rpcjson.MarshalResponse(id, result, jsonErr)
}

// standardCmdResult checks that a parsed command is a standard Bitcoin JSON-RPC
// command and runs the appropriate handler to reply to the command.  Any
// commands which are not recognized or not implemented will return an error
// suitable for use in replies.
func (s *Server) standardCmdResult(method string, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	handler, ok := handlerPool[method]
	if !ok {
		log.Error("ErrRPCMethodNotFound", method)
		return nil, rpcjson.ErrRPCMethodNotFound
	}

	return handler(s, cmd, closeChan)
}

// parseCmd parses a JSON-RPC request rpcobject into known concrete command.  The
// err field of the returned parsedRPCCmd struct will contain an RPC error that
// is suitable for use in replies if the command is invalid in some way such as
// an unregistered command or invalid parameters.
func parseCmd(request *rpcjson.Request) (interface{}, error) {
	if request.Params == nil {
		return nil, nil
	}

	rtp := cmdPool[request.Method]
	cmd := reflect.New(rtp.Elem()).Interface()
	err := json.Unmarshal(request.Params, cmd)

	return cmd, err
}

// writeHTTPResponseHeaders writes the necessary response headers prior to

// writeHTTPResponseHeaders writes the necessary response headers prior to
// writing an HTTP body given a request to use for protocol negotiation, headers
// to write, a status code, and a writer.
func (s *Server) writeHTTPResponseHeaders(req *http.Request, headers http.Header, code int, w io.Writer) error {
	_, err := io.WriteString(w, s.httpStatusLine(req, code))
	if err != nil {
		return err
	}

	err = headers.Write(w)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, "\r\n")
	return err
}

// incrementClients adds one to the number of connected RPC clients.  Note
// this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (s *Server) incrementClients() {
	atomic.AddInt32(&s.numClients, 1)
}

// decrementClients subtracts one from the number of connected RPC clients.
// Note this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (s *Server) decrementClients() {
	atomic.AddInt32(&s.numClients, -1)
}

// limitConnections responds with a 503 service unavailable and returns true if
// adding another client would exceed the maximum allow RPC clients.
//
// This function is safe for concurrent access.
func (s *Server) limitConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&s.numClients)+1) > RPCMaxClients {
		log.Info("Max RPC clients exceeded [%d] - "+
			"disconnecting client %s", RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate code set.  It also logs the error to the
// RPC server subsystem since internal errors really should not occur.  The
// context parameter is only used in the log message and may be empty if it's
// not needed.
func internalRPCError(errStr, context string) *rpcjson.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	log.Error(logStr)
	return rpcjson.NewRPCError(rpcjson.ErrRPCInternal.Code, errStr)
}

// httpStatusLine returns a response Status-Line (RFC 2616 Section 6.1)
// for the given request and response status code.  This function was lifted and
// adapted from the standard library HTTP server code since it's not exported.
func (s *Server) httpStatusLine(req *http.Request, code int) string {
	// Fast path:
	key := code
	proto11 := req.ProtoAtLeast(1, 1)
	if !proto11 {
		key = -key
	}
	s.statusLock.RLock()
	line, ok := s.statusLines[key]
	s.statusLock.RUnlock()
	if ok {
		return line
	}

	// Slow path:
	proto := "HTTP/1.0"
	if proto11 {
		proto = "HTTP/1.1"
	}
	codeStr := strconv.Itoa(code)
	text := http.StatusText(code)
	if text != "" {
		line = proto + " " + codeStr + " " + text + "\r\n"
		s.statusLock.Lock()
		s.statusLines[key] = line
		s.statusLock.Unlock()
	} else {
		text = "status code " + codeStr
		line = proto + " " + codeStr + " " + text + "\r\n"
	}

	return line
}

// RequestedProcessShutdown returns a channel that is sent to when an authorized
// RPC client requests the process to shutdown.  If the request can not be read
// immediately, it is dropped.
func (s *Server) RequestedProcessShutdown() <-chan struct{} {
	return s.requestProcessShutdown
}
