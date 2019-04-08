go build -o lcd.exe
cd client
go build -o lccli.exe
cd ..
copy lcd.exe $GOPATH/bin
copy client/lccli.exe $GOPATH/bin
