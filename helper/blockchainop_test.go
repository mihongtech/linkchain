package helper

import (
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"testing"
)

func TestSortTransaction(t *testing.T) {
	tx := CreateTempleteTx(config.DefaultTransactionVersion, config.NormalTx)

	a1 := meta.BytesToAccountID([]byte("01"))
	a2 := meta.BytesToAccountID([]byte("02"))
	a3 := meta.BytesToAccountID([]byte("03"))
	tx.SetTo(a2, *meta.NewAmount(1))
	tx.SetTo(a1, *meta.NewAmount(1))
	tx.SetTo(a3, *meta.NewAmount(1))
	tx.SetTo(a3, *meta.NewAmount(2))

	hash1, _ := math.NewHash([]byte("11"))
	hash2, _ := math.NewHash([]byte("22"))
	hash3, _ := math.NewHash([]byte("23"))
	ticket1 := meta.NewTicket(*hash1, 0)
	ticket2 := meta.NewTicket(*hash2, 0)
	ticket3 := meta.NewTicket(*hash3, 0)
	ticket4 := meta.NewTicket(*hash3, 1)
	ticket5 := meta.NewTicket(*hash3, 2)

	from1 := *meta.NewFromCoin(a1, make([]meta.Ticket, 0))
	from1.AddTicket(ticket2)
	from1.AddTicket(ticket1)
	from1.AddTicket(ticket4)
	from1.AddTicket(ticket3)
	from1.AddTicket(ticket5)

	from2 := *meta.NewFromCoin(a2, make([]meta.Ticket, 0))

	from3 := *meta.NewFromCoin(a3, make([]meta.Ticket, 0))

	tx.AddFromCoin(from2)
	tx.AddFromCoin(from1)
	tx.AddFromCoin(from3)

	t.Log("before sort", tx.String())

	SortTransaction(tx)

	t.Log("after sort", tx.String())

}
