package types

type StateDataItfc interface {
	Key() string
	Bytes() ([]byte, error)
	Copy() StateDataItfc
	OnCommit() error
}

type FromBytesFunc func(string, []byte) (StateDataItfc, error)

type StateItfc interface {
}
