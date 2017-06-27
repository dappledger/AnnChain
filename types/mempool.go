package types

type IMempool interface {
	Lock()
	Unlock()
	Update(height int64, txs []Tx)
}
