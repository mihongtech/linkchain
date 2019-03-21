package normal

import (
	"encoding/hex"
	"testing"

	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/helper"
	"github.com/linkchain/insurance"
	"github.com/linkchain/protobuf"
	"github.com/linkchain/unittest"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/config"
	"github.com/linkchain/normal/validator"
	"github.com/linkchain/storage/state"
)

var (
	normalTx = struct {
		TxId    string
		Height  uint32
		Miner   string
		HexData string
	}{
		TxId:    "6fbb001ecc61e742870ca2f6b0394f323eeb69c11733ba853eac2492c5edfdd7",
		Height:  3,
		Miner:   "025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0",
		HexData: "080110011a4f0a4d0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012260a220a201084ed40d78b8d0d5f2a0f31e4c9ac9e02489225cb58f1727947c473b22106ec100122540a280a230a21033b2b402318ce82946e2356220c7a875691bf3c44fb4843004a2210bc24178c4a12010a0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012011e2a490a473045022100832493c754da7345e732cd0a697fc953b4da2e1f2a9e726738ce1afc588b6d8a022013c26994b793a9e072854c8126249149c7af55ce455f408d88b0acdfb54f5840",
	}
	coinBaseTx = struct {
		TxId    string
		Height  uint32
		Miner   string
		HexData string
	}{
		TxId:    "cb39560bd6ee984796d84b8d9918980552a8b1e29ae864d7c6b5dca2e19b3511",
		Height:  1,
		Miner:   "02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50",
		HexData: "08ffffffff0f10001a00222a0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012013232080000000100000000",
	}
	createInsuranceTx = struct {
		TxId    string
		Height  uint32
		Miner   string
		HexData string
	}{
		TxId:    "ec0621b273c4477972f158cb259248029eacc9e4310f2a5f0d8d8bd740ed8410",
		Height:  2,
		Miner:   "03de3b38a7f61312003c61ab8bee55ba6c6aa94464dc7e5a91f4ff11bf1c60dc59",
		HexData: "080110021a4f0a4d0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012260a220a2011359be1a2dcb5c6d764e89ae2b1a852059818998d4bd8964798eed60b5639cb100022540a280a230a2103e1ea8de600f857dab5a3c3261f978d60e4280e0c012e113bd6d8681db485852c12010a0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf501201282a480a46304402202be2782918b2cf4501473d08d5bdc00abb43586cd03d33a929c7bc3e924af56602204925c9ee230425f6dec73875512bf33e871f948110d989624530958deeface8d3227080a12230a2103df398d9da7d3b4d8e7e11a4d8114685465a68ff059a2e3e762532172b3790711",
	}
)

func convertTestTransaction(data struct {
	TxId    string
	Height  uint32
	Miner   string
	HexData string
}) *meta.Transaction {
	buffer, _ := hex.DecodeString(data.HexData)
	tx := &protobuf.Transaction{}

	proto.Unmarshal(buffer, tx)

	newTx := meta.Transaction{}
	newTx.Deserialize(tx)

	return &newTx
}

func convertTestAccountId(data struct {
	TxId    string
	Height  uint32
	Miner   string
	HexData string
}) *meta.AccountID {
	buffer, _ := hex.DecodeString(data.Miner)
	return &meta.AccountID{ID: buffer}
}

func getTestInputData(testTx struct {
	TxId    string
	Height  uint32
	Miner   string
	HexData string
}) *insurance.InputData {
	inputData := validator.CreateInputData()
	db, _ := lcdb.NewMemDatabase()
	inputData.StateDB, _ = state.New(math.Hash{}, db)
	inputData.Height = testTx.Height
	inputData.BlockSigner = *convertTestAccountId(testTx)

	return inputData
}

func TestStateProcessor_processTxState(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)

	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)

	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("ec0621b273c4477972f158cb259248029eacc9e4310f2a5f0d8d8bd740ed8410")
	ticket := meta.NewTicket(*txid, uint32(1))
	u := meta.NewUTXO(ticket, uint32(2), uint32(2), *meta.NewAmount(40))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err, amount := v.ProcessTxState(convertTestTransaction(normalTx), inputData)
	unittest.NotError(t, err)
	unittest.Equal(t, amount.GetInt64(), int64(40))
}

func TestValidator_processTxState_CoinBase(t *testing.T) {
	//prepareData
	inputData := getTestInputData(coinBaseTx)
	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)

	err, amount := v.ProcessTxState(convertTestTransaction(coinBaseTx), inputData)
	unittest.NotError(t, err)
	unittest.Equal(t, amount.GetInt64(), int64(0))
}

func TestValidator_processTxState_Other(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)

	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("cb39560bd6ee984796d84b8d9918980552a8b1e29ae864d7c6b5dca2e19b3511")
	ticket := meta.NewTicket(*txid, uint32(0))
	u := meta.NewUTXO(ticket, uint32(1), uint32(1), *meta.NewAmount(50))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err, amount := v.ProcessTxState(convertTestTransaction(createInsuranceTx), inputData)
	unittest.NotError(t, err)
	unittest.Equal(t, amount.GetInt64(), int64(50))
}

func TestValidator_processTxState_ErrorTxType(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)

	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("ec0621b273c4477972f158cb259248029eacc9e4310f2a5f0d8d8bd740ed8410")
	ticket := meta.NewTicket(*txid, uint32(1))
	u := meta.NewUTXO(ticket, uint32(2), uint32(2), *meta.NewAmount(40))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	tx := convertTestTransaction(normalTx)
	tx.Type = config.TxTypeCount + 1
	err, _ := v.ProcessTxState(tx, inputData)
	unittest.Error(t, err)
}

func TestValidator_processTxState_ErrorUnCoinbase(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)

	tx := convertTestTransaction(normalTx)
	err, _ := v.ProcessTxState(tx, inputData)
	unittest.Error(t, err)
}

func TestValidator_processSpecialTx(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)

	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("cb39560bd6ee984796d84b8d9918980552a8b1e29ae864d7c6b5dca2e19b3511")
	ticket := meta.NewTicket(*txid, uint32(0))
	u := meta.NewUTXO(ticket, uint32(1), uint32(1), *meta.NewAmount(50))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err, _ := v.processSpecialTx(convertTestTransaction(createInsuranceTx), inputData)
	unittest.NotError(t, err)
}

func TestValidator_processSpecialTx_ErrorSpecial(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)

	buff, _ := hex.DecodeString("03e1ea8de600f857dab5a3c3261f978d60e4280e0c012e113bd6d8681db485852c")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("cb39560bd6ee984796d84b8d9918980552a8b1e29ae864d7c6b5dca2e19b3511")
	ticket := meta.NewTicket(*txid, uint32(0))
	u := meta.NewUTXO(ticket, uint32(1), uint32(1), *meta.NewAmount(50))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err, _ := v.processSpecialTx(convertTestTransaction(createInsuranceTx), inputData)
	unittest.Error(t, err)
}

func Test_processTxTo(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)

	err := processTxTo(convertTestTransaction(createInsuranceTx), inputData)
	unittest.NotError(t, err)
}

func Test_processTxFrom(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	v := NewStateProcessor()
	v.AddSpecialProcessor(insurance.ProcessTxState)
	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("cb39560bd6ee984796d84b8d9918980552a8b1e29ae864d7c6b5dca2e19b3511")
	ticket := meta.NewTicket(*txid, uint32(0))
	u := meta.NewUTXO(ticket, uint32(1), uint32(1), *meta.NewAmount(50))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err, amount := processTxFrom(convertTestTransaction(createInsuranceTx), inputData)
	unittest.NotError(t, err)
	unittest.Equal(t, amount.GetInt64(), int64(50))
}

func Test_processTxFrom_ErrorNoFrom(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)

	err, _ := processTxFrom(convertTestTransaction(createInsuranceTx), inputData)
	unittest.Error(t, err)
}

func Test_processTxFrom_ErrorFCTicket(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("ec0621b273c4477972f158cb259248029eacc9e4310f2a5f0d8d8bd740ed8410")
	ticket := meta.NewTicket(*txid, uint32(0))
	u := meta.NewUTXO(ticket, uint32(1), uint32(1), *meta.NewAmount(50))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err, _ := processTxFrom(convertTestTransaction(createInsuranceTx), inputData)
	unittest.Error(t, err)
}
