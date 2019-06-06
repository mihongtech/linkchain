package serialize

import "github.com/golang/protobuf/proto"

type SerializeStream interface {
	proto.Message
}

type ISerialize interface {
	//Serialize/Deserialize
	Serialize() SerializeStream //TODO Serialize() need return SerializeStream,error
	Deserialize(s SerializeStream) error

	//
	String() string
}

type Codec interface {
	ISerialize
	Encode() ([]byte, error)
	Decode([]byte) error
}
