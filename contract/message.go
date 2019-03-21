package contract

import (
	"math/big"

	"github.com/linkchain/core/meta"
)

type Message struct {
	to   *meta.AccountID
	from meta.AccountID

	amount   *big.Int
	gasLimit uint64
	gasPrice *big.Int
	data     []byte
}

func NewMessage(from meta.AccountID, to *meta.AccountID, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
	return Message{
		from:     from,
		to:       to,
		amount:   amount,
		gasLimit: gasLimit,
		gasPrice: gasPrice,
		data:     data,
	}
}

func (m Message) From() meta.AccountID { return m.from }
func (m Message) To() *meta.AccountID  { return m.to }
func (m Message) GasPrice() *big.Int   { return m.gasPrice }
func (m Message) Value() *big.Int      { return m.amount }
func (m Message) Gas() uint64          { return m.gasLimit }
func (m Message) Data() []byte         { return m.data }
