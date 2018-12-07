package datamanager

import (
	"math/big"

	"github.com/dappledger/AnnChain/genesis/chain/database"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

// AddLedgerHeaderData insert a ledger header record
func (m *DataManager) AddLedgerHeaderData(data *types.LedgerHeaderData) (uint64, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	fields := []database.Feild{
		database.Feild{Name: "height", Value: data.Height.String()},
		database.Feild{Name: "hash", Value: data.Hash.Hex()},
		database.Feild{Name: "prevhash", Value: data.PrevHash.Hex()},
		database.Feild{Name: "transactioncount", Value: data.TransactionCount},
		database.Feild{Name: "closedat", Value: data.ClosedAt},
		database.Feild{Name: "totalcoins", Value: data.TotalCoins.String()},
		database.Feild{Name: "basefee", Value: data.BaseFee.String()},
		database.Feild{Name: "maxtxsetsize", Value: data.MaxTxSetSize},
	}

	sqlRes, err := m.qdb.Insert(database.TableLedgerheaders, fields)
	if err != nil {
		return 0, err
	}

	id, err := sqlRes.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// QueryLedgerHeaderData query a specified ledger header record
func (m *DataManager) QueryLedgerHeaderData(seq *big.Int) (*types.LedgerHeaderQueryData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "height", Value: seq.String()},
	}

	var result []database.LedgerHeader
	err := m.qdb.SelectRows(database.TableLedgerheaders, where, nil, nil, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	r := result[0]
	lhd := &types.LedgerHeaderQueryData{
		Height:           ethcmn.Big(r.Height),
		Hash:             r.Hash,
		PrevHash:         r.PrevHash,
		TransactionCount: r.TransactionCount,
		ClosedAt:         r.ClosedAt,
		TotalCoins:       ethcmn.Big(r.TotalCoins),
		BaseFee:          ethcmn.Big(r.BaseFee),
		MaxTxSetSize:     r.MaxTxSetSize,
	}

	return lhd, nil
}

// QueryAllLedgerHeaderData query ledger header records using paging param
func (m *DataManager) QueryAllLedgerHeaderData(cursor, limit uint64, order string) ([]types.LedgerHeaderQueryData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "1", Value: 1},
	}
	orderT, err := database.MakeOrder(order, "ledgerid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("ledgerid", cursor, limit)

	var result []database.LedgerHeader
	err = m.qdb.SelectRows(database.TableLedgerheaders, where, orderT, paging, &result)
	if err != nil {
		return nil, err
	}

	var res []types.LedgerHeaderQueryData
	for _, r := range result {
		lhd := types.LedgerHeaderQueryData{
			Height:           ethcmn.Big(r.Height),
			Hash:             r.Hash,
			PrevHash:         r.PrevHash,
			TransactionCount: r.TransactionCount,
			ClosedAt:         r.ClosedAt,
			TotalCoins:       ethcmn.Big(r.TotalCoins),
			BaseFee:          ethcmn.Big(r.BaseFee),
			MaxTxSetSize:     r.MaxTxSetSize,
		}
		res = append(res, lhd)
	}

	return res, nil
}
