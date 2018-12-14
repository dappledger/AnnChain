package types

type CreateAccount struct {
	StartingBalance string `json:"starting_balance"`
}

type Payment struct {
	Amount string `json:"amount"`
}

type KeyPair struct {
	Value    string `json:"value"`
	Category string `json:"category"`
}
type ManageData struct {
	Mparis map[string]KeyPair
}

type CreateContract struct {
	Payload  string `json:"payload"`
	Price    string `json:"gas_price"`
	GasLimit string `json:"gas_limit"`
	Amount   string `json:"amount"`
}

type ExcuteContract struct {
	Payload  string `json:"payload"`
	Price    string `json:"gas_price"`
	GasLimit string `json:"gas_limit"`
	Amount   string `json:"amount"`
}

type QueryContract struct {
	Payload string `json:"payload"`
}
