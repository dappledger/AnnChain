package annchain

import (
	"math/big"
)

type QueryManageDataResult struct {
	Value    string `json:"value"`
	Category string `json:"category"`
}

type QueryAccountResult struct {
	Address string                           `json:"address"`
	Balance string                           `json:"balance"`
	Data    map[string]QueryManageDataResult `json:"data"`
}

type QueryLedgerResult struct {
	Height       *big.Int `json:"height"`
	Hash         string   `json:"hash"`
	PrevHash     string   `json:"prev_hash"`
	ClosedAt     string   `json:"closed_at"`
	TotalCoins   *big.Int `json:"total_coins"`
	BaseFee      *big.Int `json:"base_fee"`
	MaxTxSetSize uint64   `json:"max_tx_set_size"`
	TransCount   uint64   `json:"transaction_count"`
}

type QueryPaymentResult struct {
	Amount   string `json:"amount"`
	CreateAt uint64 `json:"created_at"`
	From     string `json:"from"`
	To       string `json:"to"`
	Hash     string `json:"hash"`
	OpType   string `json:"optype"`
}

type QueryTransactionOpTypeResult struct {
	OpType string `json:"optype"`
}

type QueryTransactionResult struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Hash     string   `json:"hash"`
	OpType   string   `json:"optype"`
	BaseFee  *big.Int `json:"basefee"`
	Height   *big.Int `json:"height"`
	Memo     string   `json:"memo"`
	CreateAt uint64   `json:"created_at"`
}

type QueryExecuteContractTransactionResult struct {
	QueryTransactionResult
	Nonce    uint64 `json:"nonce"`
	GasUsed  string `json:"gas_used"`
	GasPrice string `json:"gas_price"`
	GasLimit string `json:"gas_limit"`
	Amount   string `json:"amount"`
	PayLoad  string `json:"payload"`
}

type QueryCreateContractTransactionResult struct {
	QueryTransactionResult
	Nonce    uint64 `json:"nonce"`
	GasUsed  string `json:"gas_used"`
	GasPrice string `json:"gas_price"`
	GasLimit string `json:"gas_limit"`
	Amount   string `json:"amount"`
}

type QueryCreateAccountTransactionResult struct {
	QueryTransactionResult
	Nonce        uint64   `json:"nonce"`
	StartBalance *big.Int `json:"starting_balance"`
}

type QueryPaymentTransactionResult struct {
	QueryTransactionResult
	Nonce  uint64 `json:"nonce"`
	Amount string `json:"amount"`
}

type QueryManageDataTransactionResult struct {
	QueryTransactionResult
	Nonce   uint64            `json:"nonce"`
	KeyPair []ManageDataParam `json:"keypair"`
}

type QueryReceiptResult struct {
	From            string   `json:"from"`
	Hash            string   `json:"hash"`
	OpType          string   `json:"optype"`
	ContractAddress string   `json:"contract_address"`
	GasUsed         uint64   `json:"gas_used"`
	GasPrice        string   `json:"gas_price"`
	GasLimit        string   `json:"gas_limit"`
	Height          *big.Int `json:"height"`
	TxReceiptStatus bool     `json:"tx_receipt_status"`
	Msg             string   `json:"msg"`
	Result          string   `json:"result"`
	Logs            string   `json:"logs"`
	Payload         string   `json:"payload"`
	Nonce           uint64   `json:"nonce"`
}

type QueryContractExistResult struct {
	ByteCode string `json:"byte_code"`
	CodeHash string `json:"code_hash"`
	IsExist  bool   `json:"is_exist"`
}
