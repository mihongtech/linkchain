package wallet

import (
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/meta/account"
)

type Wallet struct {
	accounts map[btcec.PublicKey]account.IAccount
}

type Address struct {
	account.IAccount

}