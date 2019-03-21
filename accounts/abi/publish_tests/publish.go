package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"math/big"
	"os"
	"time"

	"github.com/linkchain/accounts/abi/bind"
	"github.com/linkchain/accounts/abi/publish_tests/test"
	"github.com/linkchain/client/lcclient"
	"github.com/linkchain/common/btcec"
	_ "github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
)

var DeployContractTimeout time.Duration = 60

func DeployContracts(auth *bind.TransactOpts, conn *lcclient.Client, initialSupply *big.Int) (meta.AccountID, error) {
	coin := meta.FromCoin{Id: auth.From}
	auth.FromCoin = &coin
	auth.Value = big.NewInt(3)
	auth.GasPrice = big.NewInt(1)
	auth.GasLimit = 100000000
	address, tx, _, err := test.DeployMyToken(auth, conn, initialSupply)
	if err != nil {
		log.Error("Failed to deploy new token contract", "err", err)
		return meta.AccountID{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DeployContractTimeout*time.Second)
	defer cancel()
	addressAfterMined, err := bind.WaitDeployed(ctx, conn, tx)
	if err != nil {
		log.Error("failed to deploy contact when mining", "err", err)
		return meta.AccountID{}, errors.New("Create contract wait timeout")
	}
	if bytes.Compare(address.CloneBytes(), addressAfterMined.CloneBytes()) != 0 {
		log.Error("compare address failed", "after mined address", addressAfterMined.String(), "address", address.String())
		return meta.AccountID{}, errors.New("Mind address error")
	}
	return address, nil
}

func main() {
	var (
		privKey   = flag.String("privkey", "", "account to deploy contract")
		verbosity = flag.Int("verbosity", int(log.LvlInfo), "log verbosity (0-9)")
	)
	flag.Parse()

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(*verbosity))
	log.Root().SetHandler(glogger)

	privkeyBuff, err := hex.DecodeString(*privKey)
	if err != nil {
		return
	}
	privkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeyBuff)
	auth := bind.NewKeyedTransactor(privkey)
	conn := lcclient.NewClient("")

	id, err := DeployContracts(auth, conn, big.NewInt(300))
	if err != nil {
		log.Error("deploy contract failed", "err", err)
		return
	}

	//	trans, err := test.NewMyTokenTransactor(id, conn)
	//	if err != nil {
	//		log.Error("get NewMyTokenTransactor failed", "err", err)
	//		return
	//	}
	//
	//	recID, _ := meta.HexToAccountID("56c5636befbe7cc23f5157c9278fca4e09109ffc")
	//	auth.Value = big.NewInt(0)
	//	tx, err := trans.Transfer(auth, recID, big.NewInt(500))
	//	log.Info("transct contract is", "tx", tx, "id", id)
	//
	//	time.Sleep(30 * time.Second)

	log.Info("deploy contract is", "id", id)
	caller, err := test.NewMyTokenCaller(id, conn)
	if err != nil {
		log.Error("get caller failed", "err", err)
		return
	}

	from := meta.NewAccountId(privkey.PubKey())

	value, err := caller.BalanceOf(nil, *from)
	if err != nil {
		log.Error("get balance of failed", "err", err)
		return
	}
	log.Info("deploy contract value is", "value", value)
}
