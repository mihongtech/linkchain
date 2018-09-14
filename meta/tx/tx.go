package tx
import (
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/common/math"
)


type ITxPeer interface{
}


type ITx interface {

	GetTxID() meta.DataID

	//tx content
	SetFrom(from ITxPeer)
	SetTo(to ITxPeer)
	SetAmount(meta.IAmount)

	GetFrom() ITxPeer
	GetTo() ITxPeer
	GetAmount() meta.IAmount

	ChangeFromTo() ITx

	//signature
	Sign()(math.ISignature, error)
	GetSignature()(math.ISignature)
	Verify()(error)

	//serialize
	serialize.ISerialize
}
