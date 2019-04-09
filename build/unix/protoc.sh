cd ../..
protoc --go_out=. protobuf/block.proto
protoc --go_out=. protobuf/transaction.proto
protoc --go_out=. protobuf/protobufmsg.proto
protoc --go_out=. protobuf/trie.proto
protoc --go_out=. protobuf/account.proto
protoc --go_out=. protobuf/contract.proto
