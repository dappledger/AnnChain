package types

type IMempool interface {
	Lock()
	Unlock()
	Update(height int, txs []Tx)
}
