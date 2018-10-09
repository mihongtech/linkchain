package serialize

import "github.com/golang/protobuf/proto"

type SerializeStream interface {
	proto.Message
}

type ISerialize interface {
	//Serialize/Deserialize
	Serialize() SerializeStream
	Deserialize(s SerializeStream)

	//
	String() string
}
