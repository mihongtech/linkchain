package tx
import (
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/common/math"
	"github.com/linkchain/meta/account"
)


type ITxPeer interface{
	GetID() account.IAccountID
}


type ITx interface {

	GetTxID() meta.DataID

	//tx content
	SetFrom(from ITxPeer)
	SetTo(to ITxPeer)
	SetAmount(meta.IAmount)
	SetNounce(nounce uint32)

	GetFrom() ITxPeer
	GetTo() ITxPeer
	GetAmount() meta.IAmount
	GetNounce() uint32

	ChangeFromTo() ITx

	//signature
	Sign()(math.ISignature, error)
	SetSignature(code []byte)
	GetSignature()(math.ISignature)
	Verify()(error)

	//serialize
	serialize.ISerialize
}
