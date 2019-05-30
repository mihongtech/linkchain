package normal

import (
	"encoding/hex"
	"testing"

	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/insurance"
	"github.com/mihongtech/linkchain/offchain"
	"github.com/mihongtech/linkchain/protobuf"
	"github.com/mihongtech/linkchain/storage/state"
	"github.com/mihongtech/linkchain/unittest"

	"github.com/golang/protobuf/proto"
	"github.com/mihongtech/linkchain/common"
	"github.com/mihongtech/linkchain/helper"
	"github.com/mihongtech/linkchain/node/consensus/poa"
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

func TestValidator_CheckTx_Normal(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))

	err := v.CheckTx(convertTestTransaction(normalTx))
	unittest.NotError(t, err)
}

func TestValidator_CheckTx_CoinBase(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))

	err := v.CheckTx(convertTestTransaction(coinBaseTx))
	unittest.NotError(t, err)
}

func TestValidator_CheckTx_Other(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))

	err := v.CheckTx(convertTestTransaction(createInsuranceTx))
	unittest.NotError(t, err)
}

func TestValidator_CheckTx_Error_CommonCheck(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))

	tx := convertTestTransaction(createInsuranceTx)
	tx.Type = config.TxTypeCount + 1
	err := v.CheckTx(tx)
	unittest.Error(t, err)
}

func TestValidator_checkTxSpecial(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddChecker(insurance.CheckTx)

	err := v.checkTxSpecial(convertTestTransaction(createInsuranceTx))
	unittest.NotError(t, err)
}

func TestValidator_checkTxSpecial_Error_UnCoinBase(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddChecker(insurance.CheckTx)

	tx := convertTestTransaction(createInsuranceTx)
	tx.Sign = tx.Sign[:0]
	err := v.checkTxSpecial(tx)
	unittest.Error(t, err)
}

func TestValidator_checkTxSpecial_Error_Special(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddChecker(insurance.CheckTx)

	tx := convertTestTransaction(createInsuranceTx)
	tx.Data = tx.Data[:0]
	err := v.checkTxSpecial(tx)
	unittest.Error(t, err)
}

func Test_commonCheck(t *testing.T) {
	tx := convertTestTransaction(createInsuranceTx)
	err := commonCheck(tx)
	unittest.NotError(t, err)
}

func Test_commonCheck_Error_ToCount(t *testing.T) {
	tx := convertTestTransaction(createInsuranceTx)
	tx.To.Coins = tx.To.Coins[:0]

	err := commonCheck(tx)
	unittest.Error(t, err)
}

func Test_commonCheck_Error_ToValue(t *testing.T) {
	tx := convertTestTransaction(createInsuranceTx)
	tx.To.Coins[0].Value = *meta.NewAmount(0)

	err := commonCheck(tx)
	unittest.Error(t, err)
}

func Test_commonCheck_Error_Type(t *testing.T) {
	tx := convertTestTransaction(createInsuranceTx)
	tx.Type = config.TxTypeCount + 1
	err := commonCheck(tx)
	unittest.Error(t, err)
}

func Test_checkCoinBaseTx(t *testing.T) {
	tx := convertTestTransaction(coinBaseTx)

	err := checkCoinBaseTx(tx)
	unittest.NotError(t, err)
}

func Test_checkCoinBaseTx_Error_ToCount(t *testing.T) {
	tx := convertTestTransaction(coinBaseTx)
	tx.To.Coins = append(tx.To.Coins, tx.To.Coins[0])

	err := checkCoinBaseTx(tx)
	unittest.Error(t, err)
}

func Test_checkCoinBaseTx_Error_From(t *testing.T) {
	tx := convertTestTransaction(coinBaseTx)
	temp := convertTestTransaction(normalTx)
	tx.AddFromCoin(temp.From.Coins[0])
	tx.AddSignature(&temp.Sign[0])

	err := checkCoinBaseTx(tx)
	unittest.Error(t, err)
}

func Test_checkNormalTx(t *testing.T) {
	tx := convertTestTransaction(normalTx)

	err := checkNormalTx(tx)
	unittest.NotError(t, err)
}

func Test_checkNormalTx_Error_UnCoinBase(t *testing.T) {
	tx := convertTestTransaction(coinBaseTx)

	err := checkNormalTx(tx)
	unittest.Error(t, err)
}

func Test_checkUnCoinBaseTx(t *testing.T) {
	tx := convertTestTransaction(normalTx)

	err := checkUnCoinBaseTx(tx)
	unittest.NotError(t, err)
}

func Test_checkUnCoinBaseTx_Error_FromCount(t *testing.T) {
	tx := convertTestTransaction(coinBaseTx)

	err := checkUnCoinBaseTx(tx)
	unittest.Error(t, err)
}

func Test_checkUnCoinBaseTx_Error_SameAccountId(t *testing.T) {
	tx := convertTestTransaction(normalTx)
	tx.From.Coins = append(tx.From.Coins, tx.From.Coins[0])

	err := checkUnCoinBaseTx(tx)
	unittest.Error(t, err)
}

func Test_checkUnCoinBaseTx_Error_SameTicket(t *testing.T) {
	tx := convertTestTransaction(normalTx)
	errFc := tx.From.Coins[0]
	errFc.Id = tx.To.Coins[0].Id
	tx.AddFromCoin(errFc)

	err := checkUnCoinBaseTx(tx)
	unittest.Error(t, err)
}

func getTestInputData(testTx struct {
	TxId    string
	Height  uint32
	Miner   string
	HexData string
}) *insurance.InputData {
	inputData := CreateInputData()
	db, _ := lcdb.NewMemDatabase()
	inputData.StateDB, _ = state.New(math.Hash{}, db)
	inputData.Header.Height = testTx.Height
	inputData.BlockSigner = *convertTestAccountId(testTx)
	inputData.Offchain = offchain.NewOffChainState(nil)

	return inputData
}

func TestValidator_VerifyTx_Normal(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddVerifier(insurance.VerifyTx)

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

	err := v.VerifyTx(convertTestTransaction(normalTx), inputData)
	unittest.NotError(t, err)
}

func TestValidator_VerifyTx_Other(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddVerifier(insurance.VerifyTx)

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

	err := v.VerifyTx(convertTestTransaction(createInsuranceTx), inputData)
	unittest.NotError(t, err)
}

func TestValidator_VerifyTx_Coinbase(t *testing.T) {
	//prepareData
	inputData := getTestInputData(coinBaseTx)
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddVerifier(insurance.VerifyTx)

	err := v.VerifyTx(convertTestTransaction(coinBaseTx), inputData)
	unittest.NotError(t, err)
}

func TestValidator_verifyTxSpecial(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddVerifier(insurance.VerifyTx)

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

	err := v.verifyTxSpecial(convertTestTransaction(createInsuranceTx), inputData)
	unittest.NotError(t, err)
}

func TestValidator_verifyTxSpecial_Error_UnCoinBase(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddVerifier(insurance.VerifyTx)

	err := v.verifyTxSpecial(convertTestTransaction(createInsuranceTx), inputData)
	unittest.Error(t, err)
}

func TestValidator_verifyTxSpecial_Error_TxSpecial(t *testing.T) {
	//prepareData
	inputData := getTestInputData(createInsuranceTx)
	db, _ := lcdb.NewMemDatabase()
	v := NewValidator(poa.NewPoa(&config.ChainConfig{}, db))
	v.AddVerifier(insurance.VerifyTx)

	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("cb39560bd6ee984796d84b8d9918980552a8b1e29ae864d7c6b5dca2e19b3511")
	ticket := meta.NewTicket(*txid, uint32(0))
	u := meta.NewUTXO(ticket, uint32(1), uint32(1), *meta.NewAmount(50))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)

	buff1, _ := hex.DecodeString("03e1ea8de600f857dab5a3c3261f978d60e4280e0c012e113bd6d8681db485852c")
	accountId1 := meta.AccountID{ID: buff1}
	account1 := helper.CreateTemplateAccount(accountId1)
	account1.UTXOs = append(account1.UTXOs, *u)

	inputData.StateDB.SetObject(inputData.StateDB.NewObject(meta.GetAccountHash(accountId1), *account1))
	inputData.StateDB.Commit()

	err := v.verifyTxSpecial(convertTestTransaction(createInsuranceTx), inputData)
	unittest.Error(t, err)
}

func Test_verifyNormalTx(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
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

	err := verifyNormalTx(convertTestTransaction(normalTx), inputData)
	unittest.NotError(t, err)
}

func Test_verifyNormalTx_ErrorUnCoinBase(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("ec0621b273c4477972f158cb259248029eacc9e4310f2a5f0d8d8bd740ed8410")
	ticket := meta.NewTicket(*txid, uint32(1))
	u := meta.NewUTXO(ticket, uint32(2), uint32(2), *meta.NewAmount(39))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	tx := convertTestTransaction(normalTx)
	err := verifyNormalTx(tx, inputData)
	unittest.Error(t, err)
}

func Test_verifyNormalTx_ErrorFrom(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)

	tx := convertTestTransaction(normalTx)
	err := verifyNormalTx(tx, inputData)
	unittest.Error(t, err)
}

func Test_verifyNormalTx_ErrorAccountType(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	account.AccountType = config.InsuranceAccount
	txid, _ := math.NewHashFromStr("ec0621b273c4477972f158cb259248029eacc9e4310f2a5f0d8d8bd740ed8410")
	ticket := meta.NewTicket(*txid, uint32(1))
	u := meta.NewUTXO(ticket, uint32(2), uint32(2), *meta.NewAmount(40))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err := verifyNormalTx(convertTestTransaction(normalTx), inputData)
	unittest.Error(t, err)
}

func Test_verifyNormalTx_ErrorEffectHeight(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("ec0621b273c4477972f158cb259248029eacc9e4310f2a5f0d8d8bd740ed8410")
	ticket := meta.NewTicket(*txid, uint32(1))
	u := meta.NewUTXO(ticket, uint32(2), uint32(10), *meta.NewAmount(40))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	err := verifyNormalTx(convertTestTransaction(normalTx), inputData)
	unittest.Error(t, err)
}

func Test_verifyCoinBaseTx(t *testing.T) {
	//prepareData
	inputData := getTestInputData(coinBaseTx)

	err := verifyCoinBaseTx(convertTestTransaction(coinBaseTx), inputData)
	unittest.NotError(t, err)
}

func Test_verifyCoinBaseTx_ErrorHeight(t *testing.T) {
	//prepareData
	inputData := getTestInputData(coinBaseTx)
	tx := convertTestTransaction(coinBaseTx)
	tx.Data = common.UInt32ToBytes(9)

	err := verifyCoinBaseTx(tx, inputData)
	unittest.Error(t, err)
}

func Test_verifyCoinBaseTx_ErrorToAccountType(t *testing.T) {
	//prepareData
	inputData := getTestInputData(coinBaseTx)
	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	account.AccountType = config.InsuranceAccount
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()
	tx := convertTestTransaction(coinBaseTx)

	err := verifyCoinBaseTx(tx, inputData)
	unittest.Error(t, err)
}

func Test_verifyUnCoinBaseTx(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
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
	err := verifyUnCoinBaseTx(tx, inputData)
	unittest.NotError(t, err)
}

func Test_verifyUnCoinBaseTx_ErrorFromAccount(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)

	tx := convertTestTransaction(normalTx)
	err := verifyUnCoinBaseTx(tx, inputData)
	unittest.Error(t, err)
}

func Test_verifyUnCoinBaseTx_ErrorFromCoin(t *testing.T) {
	//prepareData
	inputData := getTestInputData(normalTx)
	buff, _ := hex.DecodeString("02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50")
	accountId := meta.AccountID{ID: buff}
	account := helper.CreateTemplateAccount(accountId)
	txid, _ := math.NewHashFromStr("cb39560bd6ee984796d84b8d9918980552a8b1e29ae864d7c6b5dca2e19b3511")
	ticket := meta.NewTicket(*txid, uint32(1))
	u := meta.NewUTXO(ticket, uint32(2), uint32(2), *meta.NewAmount(40))
	account.UTXOs = append(account.UTXOs, *u)
	accountObj := inputData.StateDB.NewObject(meta.GetAccountHash(accountId), *account)
	inputData.StateDB.SetObject(accountObj)
	inputData.StateDB.Commit()

	tx := convertTestTransaction(normalTx)
	err := verifyUnCoinBaseTx(tx, inputData)
	unittest.Error(t, err)
}
