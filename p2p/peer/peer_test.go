package peer

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/mihongtech/linkchain/p2p/discover"
	"github.com/mihongtech/linkchain/p2p/message"
	"github.com/mihongtech/linkchain/p2p/peer_error"
	"github.com/mihongtech/linkchain/p2p/proto/example"
	"github.com/mihongtech/linkchain/p2p/transport"
)

var discard = Protocol{
	Name:   "discard",
	Length: 1,
	Run: func(p *Peer, rw message.MsgReadWriter) error {
		for {
			msg, err := rw.ReadMsg()
			if err != nil {
				return err
			}
			fmt.Printf("discarding %d\n", msg.Code)
			if err = msg.Discard(); err != nil {
				return err
			}
		}
	},
}

type testTransport struct {
	id       discover.NodeID
	trans    transport.Transport
	closeErr error
}

func newTestTransport(id discover.NodeID, fd net.Conn) transport.Transport {
	wrapped := transport.NewTestPbfmsg(fd)
	return &testTransport{id: id, trans: wrapped}
}

func (c *testTransport) Close(err error) {
	c.trans.Close(err)
	c.closeErr = err
}

func (c *testTransport) DoProtoHandshake(our *message.ProtoHandshake) (*message.ProtoHandshake, error) {
	return &message.ProtoHandshake{ID: c.id, Name: "test"}, nil
}

func (c *testTransport) ReadMsg() (message.Msg, error) {
	return c.trans.ReadMsg()
}

func (c *testTransport) WriteMsg(msg message.Msg) error {
	return c.trans.WriteMsg(msg)
}

func testPeer(protos []Protocol) (func(), *Conn, *Peer, <-chan error) {
	fd1, fd2 := net.Pipe()
	c1 := &Conn{FD: fd1, Transport: newTestTransport(randomID(), fd1)}
	c2 := &Conn{FD: fd2, Transport: newTestTransport(randomID(), fd2)}
	for _, p := range protos {
		c1.Caps = append(c1.Caps, p.Cap())
		c2.Caps = append(c2.Caps, p.Cap())
	}

	peer := NewPeer(c1, protos)
	errc := make(chan error, 1)
	go func() {
		_, err := peer.Run()
		errc <- err
	}()

	closer := func() { c2.Close(errors.New("close func called")) }
	return closer, c2, peer, errc
}

func TestPeerProtoReadMsg(t *testing.T) {
	test1 := uint32(1)
	test3 := uint32(2)
	test2 := uint32(3)
	testData1 := &example.TestUint{U: &test1}
	testData2 := &example.TestUint{U: &test2}
	testData3 := &example.TestUint{U: &test3}
	proto := Protocol{
		Name:   "a",
		Length: 5,
		Run: func(peer *Peer, rw message.MsgReadWriter) error {
			if err := message.ExpectMsg(rw, 2, testData1); err != nil {
				t.Error(err)
			}
			if err := message.ExpectMsg(rw, 3, testData2); err != nil {
				t.Error(err)
			}
			if err := message.ExpectMsg(rw, 4, testData3); err != nil {
				t.Error(err)
			}
			return nil
		},
	}

	closer, rw, _, errc := testPeer([]Protocol{proto})
	defer closer()

	message.Send(rw, BaseProtocolLength+2, testData1)
	message.Send(rw, BaseProtocolLength+3, testData2)
	message.Send(rw, BaseProtocolLength+4, testData3)

	select {
	case err := <-errc:
		if err != peer_error.ErrProtocolReturned {
			t.Errorf("peer returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("receive timeout")
	}
}

func TestPeerProtoEncodeMsg(t *testing.T) {
	s1 := "foo"
	s2 := "bar"
	testData := &example.TestStringArray{S1: &s1, S2: &s2}
	proto := Protocol{
		Name:   "a",
		Length: 2,
		Run: func(peer *Peer, rw message.MsgReadWriter) error {
			if err := message.SendItems(rw, 2, nil); err == nil {
				t.Error("expected error for out-of-range msg code, got nil")
			}

			if err := message.SendItems(rw, 1, testData); err != nil {
				t.Errorf("write error: %v", err)
			}
			return nil
		},
	}
	closer, rw, _, _ := testPeer([]Protocol{proto})
	defer closer()

	if err := message.ExpectMsg(rw, 17, testData); err != nil {
		t.Error(err)
	}
}

func TestPeerPing(t *testing.T) {
	closer, rw, _, _ := testPeer(nil)
	defer closer()
	if err := message.SendItems(rw, message.PingMsg, nil); err != nil {
		t.Fatal(err)
	}
	if err := message.ExpectMsg(rw, message.PongMsg, nil); err != nil {
		t.Error(err)
	}
}

func TestPeerDisconnect(t *testing.T) {
	closer, rw, _, disc := testPeer(nil)
	defer closer()
	if err := message.SendItems(rw, message.DiscMsg, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case reason := <-disc:
		if reason != peer_error.DiscRequested {
			t.Errorf("run returned wrong reason: got %v, want %v", reason, peer_error.DiscRequested)
		}
	case <-time.After(3000 * time.Millisecond):
		t.Error("peer did not return")
	}
}

// This test is supposed to verify that Peer can reliably handle
// multiple causes of disconnection occurring at the same time.
func TestPeerDisconnectRace(t *testing.T) {
	maybe := func() bool { return rand.Intn(1) == 1 }

	for i := 0; i < 1000; i++ {
		protoclose := make(chan error)
		protodisc := make(chan peer_error.DiscReason)
		closer, rw, p, disc := testPeer([]Protocol{
			{
				Name:   "closereq",
				Run:    func(p *Peer, rw message.MsgReadWriter) error { return <-protoclose },
				Length: 1,
			},
			{
				Name:   "disconnect",
				Run:    func(p *Peer, rw message.MsgReadWriter) error { p.Disconnect(<-protodisc); return nil },
				Length: 1,
			},
		})

		// Simulate incoming messages.
		go message.SendItems(rw, BaseProtocolLength+1, nil)
		go message.SendItems(rw, BaseProtocolLength+2, nil)
		// Close the network connection.
		go closer()
		// Make protocol "closereq" return.
		protoclose <- errors.New("protocol closed")
		// Make protocol "disconnect" call peer.Disconnect
		protodisc <- peer_error.DiscAlreadyConnected
		// In some cases, simulate something else calling peer.Disconnect.
		if maybe() {
			go p.Disconnect(peer_error.DiscInvalidIdentity)
		}
		// In some cases, simulate remote requesting a disconnect.
		if maybe() {
			go message.SendItems(rw, message.DiscMsg, nil)
		}

		select {
		case <-disc:
		case <-time.After(2 * time.Second):
			// Peer.run should return quickly. If it doesn't the Peer
			// goroutines are probably deadlocked. Call panic in order to
			// show the stacks.
			panic("Peer.run took to long to return.")
		}
	}
}

func TestNewPeer(t *testing.T) {
	name := "nodename"
	caps := []message.Cap{{"foo", 2}, {"bar", 3}}
	id := randomID()
	p := NewTestPeer(id, name, caps)
	if p.ID() != id {
		t.Errorf("ID mismatch: got %v, expected %v", p.ID(), id)
	}
	if p.Name() != name {
		t.Errorf("Name mismatch: got %v, expected %v", p.Name(), name)
	}
	if !reflect.DeepEqual(p.Caps(), caps) {
		t.Errorf("Caps mismatch: got %v, expected %v", p.Caps(), caps)
	}

	p.Disconnect(peer_error.DiscAlreadyConnected) // Should not hang
}

func TestMatchProtocols(t *testing.T) {
	tests := []struct {
		Remote []message.Cap
		Local  []Protocol
		Match  map[string]protoRW
	}{
		{
			// No remote capabilities
			Local: []Protocol{{Name: "a"}},
		},
		{
			// No local protocols
			Remote: []message.Cap{{Name: "a"}},
		},
		{
			// No mutual protocols
			Remote: []message.Cap{{Name: "a"}},
			Local:  []Protocol{{Name: "b"}},
		},
		{
			// Some matches, some differences
			Remote: []message.Cap{{Name: "local"}, {Name: "match1"}, {Name: "match2"}},
			Local:  []Protocol{{Name: "match1"}, {Name: "match2"}, {Name: "remote"}},
			Match:  map[string]protoRW{"match1": {Protocol: Protocol{Name: "match1"}}, "match2": {Protocol: Protocol{Name: "match2"}}},
		},
		{
			// Various alphabetical ordering
			Remote: []message.Cap{{Name: "aa"}, {Name: "ab"}, {Name: "bb"}, {Name: "ba"}},
			Local:  []Protocol{{Name: "ba"}, {Name: "bb"}, {Name: "ab"}, {Name: "aa"}},
			Match:  map[string]protoRW{"aa": {Protocol: Protocol{Name: "aa"}}, "ab": {Protocol: Protocol{Name: "ab"}}, "ba": {Protocol: Protocol{Name: "ba"}}, "bb": {Protocol: Protocol{Name: "bb"}}},
		},
		{
			// No mutual versions
			Remote: []message.Cap{{Version: 1}},
			Local:  []Protocol{{Version: 2}},
		},
		{
			// Multiple versions, single common
			Remote: []message.Cap{{Version: 1}, {Version: 2}},
			Local:  []Protocol{{Version: 2}, {Version: 3}},
			Match:  map[string]protoRW{"": {Protocol: Protocol{Version: 2}}},
		},
		{
			// Multiple versions, multiple common
			Remote: []message.Cap{{Version: 1}, {Version: 2}, {Version: 3}, {Version: 4}},
			Local:  []Protocol{{Version: 2}, {Version: 3}},
			Match:  map[string]protoRW{"": {Protocol: Protocol{Version: 3}}},
		},
		{
			// Various version orderings
			Remote: []message.Cap{{Version: 4}, {Version: 1}, {Version: 3}, {Version: 2}},
			Local:  []Protocol{{Version: 2}, {Version: 3}, {Version: 1}},
			Match:  map[string]protoRW{"": {Protocol: Protocol{Version: 3}}},
		},
		{
			// Versions overriding sub-protocol lengths
			Remote: []message.Cap{{Version: 1}, {Version: 2}, {Version: 3}, {Name: "a"}},
			Local:  []Protocol{{Version: 1, Length: 1}, {Version: 2, Length: 2}, {Version: 3, Length: 3}, {Name: "a"}},
			Match:  map[string]protoRW{"": {Protocol: Protocol{Version: 3}}, "a": {Protocol: Protocol{Name: "a"}, offset: 3}},
		},
	}

	for i, tt := range tests {
		result := matchProtocols(tt.Local, tt.Remote, nil)
		if len(result) != len(tt.Match) {
			t.Errorf("test %d: negotiation mismatch: have %v, want %v", i, len(result), len(tt.Match))
			continue
		}
		// Make sure all negotiated protocols are needed and correct
		for name, proto := range result {
			match, ok := tt.Match[name]
			if !ok {
				t.Errorf("test %d, proto '%s': negotiated but shouldn't have", i, name)
				continue
			}
			if proto.Name != match.Name {
				t.Errorf("test %d, proto '%s': name mismatch: have %v, want %v", i, name, proto.Name, match.Name)
			}
			if proto.Version != match.Version {
				t.Errorf("test %d, proto '%s': version mismatch: have %v, want %v", i, name, proto.Version, match.Version)
			}
			if proto.offset-BaseProtocolLength != match.offset {
				t.Errorf("test %d, proto '%s': offset mismatch: have %v, want %v", i, name, proto.offset-BaseProtocolLength, match.offset)
			}
		}
		// Make sure no protocols missed negotiation
		for name := range tt.Match {
			if _, ok := result[name]; !ok {
				t.Errorf("test %d, proto '%s': not negotiated, should have", i, name)
				continue
			}
		}
	}
}

func randomID() (id discover.NodeID) {
	for i := range id {
		id[i] = byte(rand.Intn(255))
	}
	return id
}
