package annchain

const (
	CREATE_ACCOUNT   = "create_account"
	PAYMENT          = "payment"
	MANAGE_DATA      = "manage_data"
	CREATE_CONTRACT  = "create_contract"
	EXECUTE_CONTRACT = "execute_contract"
	QUERY_CONTRACT   = "query_contract"
)

type CreateAccountParam struct {
	StartBalance string `json:"starting_balance"`
}

type PaymentParam struct {
	Amount string `json:"amount"`
}

type ManageDataParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ContractParam struct {
	PayLoad  string `json:"payload"`
	GasPrice string `json:"gas_price"`
	GasLimit string `json:"gas_limit"`
	Amount   string `json:"amount"`
}
