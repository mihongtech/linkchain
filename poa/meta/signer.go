package meta

import (
	"encoding/hex"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/protobuf"
)

const (
	firstPubMiner   = "025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0"
	fristPrivMiner  = "55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b"
	secondPubMiner  = "02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50"
	secondPrivMiner = "7a9c6f2b865c98c9fe174869de5818f4c62bc845441c08269487cdba6688f6b1"
	thirdPubMiner   = "03de3b38a7f61312003c61ab8bee55ba6c6aa94464dc7e5a91f4ff11bf1c60dc59"
	thirdPrivMiner  = "6647e717248720f1b50f3f1f765b731783205f2de2fedc9e447438966af7df85"
)

var PubSigners = []string{firstPubMiner, secondPubMiner, thirdPubMiner}

var PrivSigner = []string{fristPrivMiner, secondPrivMiner, thirdPrivMiner}

type Signer TransactionPeer

func NewSigner(id AccountID, extra []byte) *Signer {
	return &Signer{AccountID: id, Extra: extra}
}

func CreateSignerIdByPubKey(pubKey string) (*Signer, error) {
	id, err := CreateAccountIdByPubKey(pubKey)
	if err != nil {
		return nil, err
	}
	return NewSigner(*id, nil), nil
}

func CreateSignerIdByPrivKey(privKey string) (*Signer, error) {
	id, err := CreateAccountIdByPrivKey(privKey)
	if err != nil {
		return nil, err
	}
	return NewSigner(*id, nil), nil
}

func (s *Signer) Sign(signerPriv string, signHash math.Hash) error {
	privbuff, err := hex.DecodeString(signerPriv)
	if err != nil {
		return nil
	}
	priv, pub := btcec.PrivKeyFromBytes(btcec.S256(), privbuff)
	if !pub.IsEqual(&s.AccountID.ID) {
		return errors.New("Signer Sign() pubkey is error")
	}
	signature, err := priv.Sign(signHash.CloneBytes())
	if err != nil {
		log.Error("Signer", "Sign()", err)
		return err
	}
	s.Extra = signature.Serialize()
	return nil
}

func (s *Signer) Verify(signHash math.Hash) error {
	signature, err := btcec.ParseSignature(s.Extra, btcec.S256())
	if err != nil {
		log.Error("Signer", "VerifySign", err)
		return err
	}

	verified := signature.Verify(signHash.CloneBytes(), &s.AccountID.ID)
	if !verified {
		return errors.New("Signer VerifySign failed: Error Sign")
	}
	return nil
}

func (s *Signer) IsEqual(signer Signer) bool {
	return s.AccountID.ID.IsEqual(&signer.AccountID.ID)
}

func (s *Signer) Decode() ([]byte, error) {
	buffer, err := proto.Marshal(s.Serialize())
	if err != nil {
		return buffer, err
	}
	return buffer, nil
}

func (s *Signer) Encode(buff []byte) error {
	data := &protobuf.TransactionPeer{}
	err := proto.Unmarshal(buff, data)
	if err != nil {
		return err
	}
	s.Deserialize(data)
	return nil
}

//Serialize/Deserialize
func (txpeer *Signer) Serialize() serialize.SerializeStream {
	accountID := txpeer.AccountID.Serialize().(*protobuf.AccountID)
	peer := protobuf.TransactionPeer{
		AccountID: accountID,
		Extra:     proto.NewBuffer(txpeer.Extra).Bytes(),
	}
	return &peer
}

func (txpeer *Signer) Deserialize(s serialize.SerializeStream) {
	data := *s.(*protobuf.TransactionPeer)
	txpeer.AccountID.Deserialize(data.AccountID)
	txpeer.Extra = data.Extra
}

func (txpeer *Signer) GetID() account.IAccountID {
	return &txpeer.AccountID
}
