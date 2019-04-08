cd ../..
go build -o lcd
(cd client;go build -o lccli)
cp ./lcd $GOPATH/bin
cp ./client/lccli $GOPATH/bin
