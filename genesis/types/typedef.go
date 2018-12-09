package types

import (
	"fmt"
	"math/big"
	"time"

	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/ann-module/lib/go-merkle"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

type OP_NAME string
type TYPE_OP int

var (
	OP_S_CREATEACCOUNT   OP_NAME = "create_account"
	OP_S_PAYMENT         OP_NAME = "payment"
	OP_S_MANAGEDATA      OP_NAME = "manage_data"
	OP_S_CREATECONTRACT  OP_NAME = "create_contract"
	OP_S_EXECUTECONTRACT OP_NAME = "execute_contract"
)

func (o OP_NAME) OpStr() string {
	switch o {
	case OP_S_CREATEACCOUNT:
		return "create_account"
	case OP_S_PAYMENT:
		return "payment"
	case OP_S_MANAGEDATA:
		return "manage_data"
	case OP_S_CREATECONTRACT:
		return "create_contract"
	case OP_S_EXECUTECONTRACT:
		return "execute_contract"
	default:
		return ""
	}
}
func (o OP_NAME) OpInt() TYPE_OP {
	switch o {
	case OP_S_CREATEACCOUNT:
		return 1
	case OP_S_PAYMENT:
		return 2
	case OP_S_MANAGEDATA:
		return 3
	case OP_S_CREATEACCOUNT:
		return 4
	case OP_S_EXECUTECONTRACT:
		return 5
	default:
		return 0
	}
}

type TYPE_EFFECT int

const (
	EffectTypeAccountCreated TYPE_EFFECT = iota
	EffectTypeAccountCredited
	EffectTypeAccountDebited
	EffectTypePayment
	EffectTypeManageData
	EffectTypeCreateContract
	EffectTypeExecuteContract
)

func (e TYPE_EFFECT) String() string {
	switch e {
	case EffectTypeAccountCreated:
		return "create_account"
	case EffectTypeAccountCredited:
		return "account_credite"
	case EffectTypeAccountDebited:
		return "account_debite"
	case EffectTypePayment:
		return "payment"
	case EffectTypeManageData:
		return "manage_data"
	case EffectTypeCreateContract:
		return "create_contract"
	case EffectTypeExecuteContract:
		return "excute_contract"
	default:
		return ""
	}
}

type ERRORTYPE types.CodeType

const (
	CREATE_ACCOUNT_INVALID          ERRORTYPE = -100 //目标帐号无效
	CREATE_ACCOUNT_UNDEFUNDED       ERRORTYPE = -101 //源账户余额不足，不能初始账户
	CREATE_ACCOUNT_LOW_RESERVE      ERRORTYPE = -102 //保证金不足
	CREATE_ACCOUNT_ALREADY_EXIST    ERRORTYPE = -103 //已存在
	CREATE_ACCOUNT_BAD_SOURCE       ERRORTYPE = -104
	CREATE_ACCOUNT_BAD_STARTBALANCE ERRORTYPE = -105

	PAYMENT_MALFORMED          ERRORTYPE = -200
	PAYMENT_UNDERUNDED         ERRORTYPE = -201
	PAYMENT_SRC_NO_TRUST       ERRORTYPE = -202
	PAYMENT_SRC_NOT_AUTHORIZED ERRORTYPE = -203
	PAYMENT_NO_DESTINATION     ERRORTYPE = -204
	PAYMENT_FROM_TO_EQUAL      ERRORTYPE = -205
	PAYMENT_NOT_AUTHORIZED     ERRORTYPE = -206
	PAYMENT_LINE_FULL          ERRORTYPE = -207
	PAYMENT_NO_ISSUER          ERRORTYPE = -208
	PAYMENT_AMOUNT_LOW         ERRORTYPE = -209

	MANAGE_SOURCE_INVALID          ERRORTYPE = -300
	MANAGE_DATA_NAME_NOT_FOUND     ERRORTYPE = -301
	MANAGE_DATA_LOW_RESERVE        ERRORTYPE = -302
	MANAGE_DATA_INVALID_NAME       ERRORTYPE = -303
	MANAGE_DATA_NAME_ALREADY_EXIST ERRORTYPE = -304
	MANAGE_DATA_DB_SAVE            ERRORTYPE = -305

	CONTRACT_SOURCE_INVALID    ERRORTYPE = -400
	CONTRACT_LIMIT_INVALID     ERRORTYPE = -401
	CONTRACT_PRICE_INVALID     ERRORTYPE = -402
	CONTRACT_AMOUNT_INVALID    ERRORTYPE = -403
	CONTRACT_ADDRESS_INVALID   ERRORTYPE = -404
	CONTRACT_ADDRESS_NOT_EXIST ERRORTYPE = -405

	INFLATION_NOT_TIME ERRORTYPE = -1
)

func (e ERRORTYPE) String() string {
	switch e {
	case CREATE_ACCOUNT_INVALID:
		return "invalid target account"
	case CREATE_ACCOUNT_LOW_RESERVE:
		return "starting balance not enough"
	case CREATE_ACCOUNT_BAD_SOURCE:
		return "source address not init_account"
	case CREATE_ACCOUNT_ALREADY_EXIST:
		return "target address is already exist"
	case PAYMENT_NO_DESTINATION:
		return "destination account not exist"
	case PAYMENT_FROM_TO_EQUAL:
		return "from can not be to"
	case PAYMENT_AMOUNT_LOW:
		return "amount can not be zero"
	case PAYMENT_UNDERUNDED:
		return "not enough balance"
	case CONTRACT_SOURCE_INVALID:
		return "source address invalid"
	case CONTRACT_LIMIT_INVALID:
		return "gas limit invalid"
	case CONTRACT_PRICE_INVALID:
		return "gas price invalid"
	case CONTRACT_AMOUNT_INVALID:
		return "amount invalid"
	case CONTRACT_ADDRESS_INVALID:
		return "contract address invalid"
	case CONTRACT_ADDRESS_NOT_EXIST:
		return "contract address not exist"
	case MANAGE_SOURCE_INVALID:
		return "manage data source invalid"
	case MANAGE_DATA_NAME_NOT_FOUND:
		return "manage data key not found"
	case MANAGE_DATA_DB_SAVE:
		return "manage data db save error"
	default:
		return ""
	}

}

var BIG_INT0 = big.NewInt(0)
var BIG_MINBLC = big.NewInt(0)
var ZERO_ADDRESS = ethcmn.Address{}
var ZERO_HASH = ethcmn.Hash{}

type API_QUERY_TYPE byte

const (
	API_QUERY_NONCE API_QUERY_TYPE = iota + 1
	API_QUERY_BALANCE
	API_QUERY_RECEIPT
	API_QUERY_ACCOUNT
	API_QUERY_PAYMENT
	API_QUERY_TX
	API_LIST_TXS
	API_QUERY_ACTION
	API_QUERY_ACTION_AC
	API_QUERY_ACTION_TX
	API_QUERY_EFFECTS
	API_QUERY_PAYMENT_TX
	API_QUERY_ALL_BIGDATA
	API_QUERY_SINGLE_BIGDATA
	API_QUERY_OFFER_AC
	API_QUERY_ORDER_BOOK
	API_QUERY_TRADES
	API_QUERY_TRADES_AC
	API_QUERY_ALL_LEDGER
	API_QUERY_SEQ_LEDGER
	API_QUERY_CONTRACT
	API_QUERY_CONTRACT_EXIST
	API_QUERY_MANAGEDATA
	API_QUERY_SINGLE_MANAGEDATA
)

func (at API_QUERY_TYPE) AppendBytes(bt []byte) []byte {
	return append([]byte{byte(at)}, bt...)
}

// AppHeader ledger header for app
type AppHeader struct {
	// Config cfg.Config

	StateRoot ethcmn.Hash // fill after statedb commit

	Height   *big.Int  // refresh by new block
	ClosedAt time.Time // refresh by new block

	BaseFee      *big.Int // dynamic get according to block height
	MaxTxSetSize uint64   // dynamic get according to block height

	TxCount uint64 // fill when ready to save

	PrevHash  ethcmn.LedgerHash // global, just used to calculate header-hash
	TotalCoin *big.Int          // global, need be saved in lastblock
	Feepool   *big.Int          // global, need be saved in lastblock

}

func (h *AppHeader) String() string {
	return fmt.Sprintf("prevhash:%v,stateRoot:%v,height:%v,txCount:%v,BaseFee:%v,totalCoin:%v,feePool:%v,MaxTxSetSize:%v,ClosedAt:%v",
		h.PrevHash.Hex(),
		h.StateRoot.Hex(),
		h.Height,
		h.TxCount,
		h.BaseFee,
		h.TotalCoin,
		h.Feepool,
		h.MaxTxSetSize,
		h.ClosedAt)
}

// Hash hash
func (h *AppHeader) Hash() []byte {

	m := map[string]interface{}{
		"PrevHash":  h.PrevHash,
		"StateRoot": h.StateRoot,
		"Height":    h.Height,
		"TxCount":   h.TxCount,
		"BaseFee":   h.BaseFee,
		"TotalCoin": h.TotalCoin,
		"Feepool":   h.Feepool,
	}
	return merkle.SimpleHashFromMap(m)
}

// GetLedgerHeaderData create LedgerHeaderData using appheader info
func (h *AppHeader) GetLedgerHeaderData() *LedgerHeaderData {

	return &LedgerHeaderData{
		Height:           h.Height,
		Hash:             ethcmn.BytesToLedgerHash(h.Hash()),
		PrevHash:         h.PrevHash,
		TransactionCount: h.TxCount,
		ClosedAt:         h.ClosedAt,
		TotalCoins:       h.TotalCoin,
		BaseFee:          h.BaseFee,
		MaxTxSetSize:     h.MaxTxSetSize,
	}
}

type QueryBase struct {
	Order  string
	Limit  uint64
	Cursor uint64

	Begin uint64
	End   uint64
}

type AccountQuery struct {
	Account ethcmn.Address
}

type ContractQuery struct {
	ContractAddr ethcmn.Address
}

type ReceiptQuery struct {
	TxHash ethcmn.Hash
}

type ActionsQuery struct {
	QueryBase
	Account ethcmn.Address
	TxHash  ethcmn.Hash
	Typei   uint64
}

type EffectsQuery struct {
	QueryBase
	Account ethcmn.Address
	TxHash  ethcmn.Hash
	Typei   uint64
}

type TxQuery struct {
	QueryBase
	TxHash  ethcmn.Hash
	Account ethcmn.Address
	Height  string
}

type ManageDataQuery struct {
	QueryBase
	Account ethcmn.Address
	Name    string
	IsPub   string
}

type LedgerQuery struct {
	QueryBase
	Sequence uint64
}

type SingleData struct {
	Account ethcmn.Address
	Hash    string
}

type SingleManageData struct {
	Account ethcmn.Address
	Keys    string
}
type RspConvert struct {
	IsSuccess  bool    `json:"isSuccess"`
	ResultInfo *Result `json:"result"`
}
type Result struct {
	Privkey string `json:"privkey"`
	Address string `json:"address"`
}

const (
	OfferTypeActive            = 1
	OfferTypePassive           = 2
	INFLATION_START_TIME       = 1404172800 // 1-jul-2014
	INFLATION_FREQUENCY        = uint64(time.Hour) * 24 * 7
	INFLATION_RATE_TRILLIONTHS = 190721000
	INFLATION_NUM_WINNERS      = 2000
	INFLATION_WIN_MIN_PERCENT  = 500000000
	FIRST_MONDAY               = uint64(time.Hour)*24*4 + 1 // 1970-01-05 00:00:01

	// million : 1 000 000
	// billion : 1 000 000 000
	// trillion : 1 000 000 000 000
	TRILLION = 1000000000000
)

var BIG_BILLION = big.NewInt(1000000000)
var ORI_COINS = big.NewInt(100000000000000000)
