package annchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	at "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/go-sdk/annchain-go-sdk/rpc"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/crypto"
)

func GenerateKey() (privekey, address string) {

	privkey, err := crypto.GenerateKey()
	if err != nil {
		return "", ""
	}

	privekey = ethcmn.Bytes2Hex(crypto.FromECDSA(privkey))

	address = crypto.PubkeyToAddress(privkey.PublicKey).Hex()

	return
}

type AnnChainClient struct {
	rpcClient *rpc.ClientJSONRPC
}

func NewAnnChainClient(address string) *AnnChainClient {
	return &AnnChainClient{
		rpcClient: rpc.NewClientJSONRPC(address),
	}
}

func (c *AnnChainClient) ExcuteContract(nonce uint64, privKey, basefee, memo, from, to string, contractParam ContractParam) (string, at.CodeType, error) {

	privateKey := crypto.ToECDSA(ethcmn.Hex2Bytes(privKey))

	signBytes, txhash, code, err := c.signAndEncodeTx(NewExecuteContractTx(nonce, basefee, memo, from, to, contractParam), privateKey)
	if err != nil {
		return "", code, err
	}
	_, code, err = c.rpcClient.Call("execute_contract", []interface{}{signBytes}, nil)

	return txhash, code, err
}

func (c *AnnChainClient) CreateContract(nonce uint64, privKey, basefee, memo, from string, contractParam ContractParam) (string, at.CodeType, error) {

	privateKey := crypto.ToECDSA(ethcmn.Hex2Bytes(privKey))

	signBytes, txhash, code, err := c.signAndEncodeTx(NewCreateContractTx(nonce, basefee, memo, from, contractParam), privateKey)
	if err != nil {
		return "", code, err
	}

	_, code, err = c.rpcClient.Call("create_contract", []interface{}{signBytes}, nil)

	return txhash, code, err
}

func (c *AnnChainClient) ManageData(nonce uint64, privKey, basefee, memo, from string, datas []ManageDataParam) (string, at.CodeType, error) {

	privateKey := crypto.ToECDSA(ethcmn.Hex2Bytes(privKey))

	signBytes, txhash, code, err := c.signAndEncodeTx(NewManageDataTx(nonce, basefee, memo, from, datas), privateKey)
	if err != nil {
		return "", code, err
	}

	_, code, err = c.rpcClient.Call("manage_data", []interface{}{signBytes}, nil)

	return txhash, code, err
}

func (c *AnnChainClient) Payment(nonce uint64, privKey, basefee, memo, from, to, amount string) (string, at.CodeType, error) {

	privateKey := crypto.ToECDSA(ethcmn.Hex2Bytes(privKey))

	signBytes, txhash, code, err := c.signAndEncodeTx(NewPaymentTx(nonce, basefee, memo, from, to, amount), privateKey)
	if err != nil {
		return "", code, err
	}

	_, code, err = c.rpcClient.Call("payment", []interface{}{signBytes}, nil)

	return txhash, code, err
}

func (c *AnnChainClient) CreateAccount(nonce uint64, privKey, basefee, memo, from, to, startBalance string) (string, at.CodeType, error) {

	privateKey := crypto.ToECDSA(ethcmn.Hex2Bytes(privKey))

	signBytes, txhash, code, err := c.signAndEncodeTx(NewCreateAccountTx(nonce, basefee, memo, from, to, startBalance), privateKey)
	if err != nil {
		return "", code, err
	}
	_, code, err = c.rpcClient.Call("create_account", []interface{}{signBytes}, nil)

	return txhash, code, err
}

func (c *AnnChainClient) QueryNonce(address string) (uint64, at.CodeType, error) {

	result, code, err := c.rpcClient.Call("query_nonce", []interface{}{address}, nil)
	if err != nil {
		return 0, code, err
	}
	nonce, _ := strconv.ParseUint(string(result), 10, 64)
	return nonce, at.CodeType_OK, nil
}

func (c *AnnChainClient) QueryAccount(address string) (QueryAccountResult, at.CodeType, error) {

	var query QueryAccountResult

	_, code, err := c.rpcClient.Call("query_account", []interface{}{address}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryLeders(order string, limit, cursor uint64) ([]QueryLedgerResult, at.CodeType, error) {

	var query []QueryLedgerResult

	_, code, err := c.rpcClient.Call("query_ledgers", []interface{}{order, limit, cursor}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryLeder(height *big.Int) ([]QueryLedgerResult, at.CodeType, error) {

	var query []QueryLedgerResult

	_, code, err := c.rpcClient.Call("query_ledger", []interface{}{height}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryPayments(order string, limit, cursor uint64) ([]QueryPaymentResult, at.CodeType, error) {

	var query []QueryPaymentResult

	_, code, err := c.rpcClient.Call("query_payments", []interface{}{order, limit, cursor}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryAccountPayments(address, order string, limit, cursor uint64) ([]QueryPaymentResult, at.CodeType, error) {

	var query []QueryPaymentResult

	_, code, err := c.rpcClient.Call("query_account_payments", []interface{}{address, order, limit, cursor}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryPayment(txhash string) ([]QueryPaymentResult, at.CodeType, error) {

	var query []QueryPaymentResult

	_, code, err := c.rpcClient.Call("query_payment", []interface{}{txhash}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryTransactions(order string, limit, cursor uint64) ([]QueryTransactionResult, at.CodeType, error) {

	var query []QueryTransactionResult

	_, code, err := c.rpcClient.Call("query_transactions", []interface{}{order, limit, cursor}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryTransaction(txhash string) (interface{}, at.CodeType, error) {

	var queryType []QueryTransactionOpTypeResult

	jsonObj, code, err := c.rpcClient.Call("query_transaction", []interface{}{txhash}, nil)
	if code != at.CodeType_OK {
		return nil, code, err
	}

	if err = json.Unmarshal(jsonObj, &queryType); err != nil {
		return nil, at.CodeType_DecodingError, err
	}

	if len(queryType) <= 0 {
		return nil, at.CodeType_DecodingError, fmt.Errorf("txhash %v not exist", txhash)
	}

	switch queryType[0].OpType {
	case CREATE_ACCOUNT:
		var query []QueryCreateAccountTransactionResult
		if err = json.Unmarshal(jsonObj, &query); err != nil {
			return nil, at.CodeType_DecodingError, err
		}
		return query, at.CodeType_OK, nil
	case PAYMENT:
		var query []QueryPaymentTransactionResult
		if err = json.Unmarshal(jsonObj, &query); err != nil {
			return nil, at.CodeType_DecodingError, err
		}
		return query, at.CodeType_OK, nil
	case CREATE_CONTRACT:
		var query []QueryCreateContractTransactionResult
		if err = json.Unmarshal(jsonObj, &query); err != nil {
			return nil, at.CodeType_DecodingError, err
		}
		return query, at.CodeType_OK, nil
	case EXECUTE_CONTRACT:
		var query []QueryExecuteContractTransactionResult
		if err = json.Unmarshal(jsonObj, &query); err != nil {
			return nil, at.CodeType_DecodingError, err
		}
		return query, at.CodeType_OK, nil
	case MANAGE_DATA:
		var query []QueryManageDataTransactionResult
		if err = json.Unmarshal(jsonObj, &query); err != nil {
			return nil, at.CodeType_DecodingError, err
		}
		return query, at.CodeType_OK, nil
	}
	return nil, at.CodeType_DecodingError, fmt.Errorf("txhash %v not exist", txhash)
}

func (c *AnnChainClient) QueryAccountTransactions(address, order string, limit, cursor uint64) ([]QueryTransactionResult, at.CodeType, error) {

	var query []QueryTransactionResult

	_, code, err := c.rpcClient.Call("query_account_transactions", []interface{}{address, order, limit, cursor}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryLedgerTransactions(height uint64, order string, limit, cursor uint64) ([]QueryTransactionResult, at.CodeType, error) {

	var query []QueryTransactionResult

	_, code, err := c.rpcClient.Call("query_ledger_transactions", []interface{}{height, order, limit, cursor}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryContract(privKey, from, to, funcName, abis string, params []interface{}) (interface{}, at.CodeType, error) {

	var payLoad string

	privateKey := crypto.ToECDSA(ethcmn.Hex2Bytes(privKey))

	contractParam, err := NewQueryContractParam(funcName, abis, params)
	if err != nil {
		return nil, at.CodeType_EncodingError, err
	}

	signBytes, _, code, err := c.signAndEncodeTx(NewQueryContractTx(from, to, contractParam), privateKey)
	if err != nil {
		return nil, code, err
	}

	if _, code, err = c.rpcClient.Call("query_contract", []interface{}{signBytes}, &payLoad); err != nil {
		return nil, code, err
	}

	bPayLoad, err := hex.DecodeString(payLoad)
	if err != nil {
		return nil, at.CodeType_EncodingError, err
	}
	result, err := unpackResultToArray(funcName, abis, bPayLoad)
	if err != nil {
		return nil, at.CodeType_DecodingError, err
	}

	return result, at.CodeType_OK, nil
}

func (c *AnnChainClient) QueryContractExist(contractAddress string) (QueryContractExistResult, at.CodeType, error) {

	var query QueryContractExistResult

	_, code, err := c.rpcClient.Call("query_contract_exist", []interface{}{contractAddress}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryReceipt(txhash string) (QueryReceiptResult, at.CodeType, error) {

	var query QueryReceiptResult

	_, code, err := c.rpcClient.Call("query_receipt", []interface{}{txhash}, &query)

	return query, code, err

}

func (c *AnnChainClient) QueryAccountManageDatas(address, order string, limit, cursor uint64) ([]QueryManageDataResult, at.CodeType, error) {

	var query []QueryManageDataResult

	_, code, err := c.rpcClient.Call("query_account_managedatas", []interface{}{address, order, limit, cursor}, &query)

	return query, code, err
}

func (c *AnnChainClient) QueryAccountManageData(address, name string) (map[string]string, at.CodeType, error) {

	query := make(map[string]string)

	_, code, err := c.rpcClient.Call("query_account_managedata", []interface{}{address, name}, &query)

	return query, code, err
}
