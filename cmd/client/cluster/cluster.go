package cluster

type ValidatorInfo struct {
	RPC     string
	Address string
}

type Cluster interface {
	Up() error
	Down() error
	GetValidatorInfo(index int) (*ValidatorInfo, error)
	StartValidator(index int) error
	StopValidator(index int) error
}
