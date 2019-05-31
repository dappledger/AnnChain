package commons

import "errors"

var (
	// ErrEmptyAddress means transaction sender is empty
	ErrEmptyAddress = errors.New("address is empty")
	// ErrEmptySender means transaction sender is empty
	ErrEmptySender = errors.New("sender is empty")
	// ErrEmptyPassphrase errors
	ErrEmptyPassphrase = errors.New("passphrase is empty")
	// ErrIncorrectPassphrase errors
	ErrIncorrectPassphrase = errors.New("passphrase is incorrect")

	ErrEmptyTemplateName    = errors.New("template's name is empty")
	ErrEmptyTemplateVersion = errors.New("template's version is empty")
	ErrEmptyTemplateSource  = errors.New("template's source is empty")
	ErrEmptyTemplateCode    = errors.New("template's code is empty")
	ErrEmptyTemplateABI     = errors.New("template's abi is empty")
	ErrDuplicatedTemplate   = errors.New("template exists already")
	ErrInvalidTemplateID    = errors.New("template's id is invalid")
	ErrNoSuchTemplate       = errors.New("no such template")

	// ErrEmptyContractCode errors
	ErrEmptyContractCode = errors.New("contract's code is empty")
	// ErrEmptyContractABI errors
	ErrEmptyContractABI = errors.New("contract's ABI is empty")
	// ErrEmptyContractAddress errors
	ErrEmptyContractAddress = errors.New("contract's address is empty")
	// ErrEmptyContractCreator errors
	ErrEmptyContractCreator = errors.New("contract's creator is empty")

	// ErrContractCreationTimeout errors
	ErrContractCreationTimeout = errors.New("contract creating timeout")
	// ErrContractCreationEmptyCode errors
	ErrContractCreationEmptyCode = errors.New("Fail to deploy contract, please check your input")
	// ErrInvalidContractTemplateID errors
	ErrInvalidContractTemplateID = errors.New("contract's templateID is invalid")
	// ErrInvalidContractID errors
	ErrInvalidContractID = errors.New("contract's ID is invalid")

	// ErrEmptyMethod errors
	ErrEmptyMethod = errors.New("method is empty")
	// ErrNoSuchMethod errors
	ErrNoSuchMethod = errors.New("no such method")
	// ErrNoSuchParam errors
	ErrNoSuchParam = errors.New("no such parameter")
	// ErrUnmatchedParams errors
	ErrUnmatchedParams = errors.New("number of params is unmatched")
	// ErrMethodNotConstant errors
	ErrMethodNotConstant = errors.New("method is not constant")
	// ErrInvalidOutputsNumber errors
	ErrInvalidOutputsNumber = errors.New("only the constant function that returns exactly one result could be read")
	// ErrUnsupportedOutputType errors
	ErrUnsupportedOutputType = errors.New("unsupported solidity output type")

	// ErrEmptyJSONRPCID errors
	ErrEmptyJSONRPCID = errors.New("id of JSONRPC request cannot be null")
	// ErrEmptyJSONRPCMethod errors
	ErrEmptyJSONRPCMethod = errors.New("method of JSONRPC request cannot be null")

	ErrNoSuchTransaction = errors.New("no such transaction")

	ErrNoSuchReceipt            = errors.New("no such transaction receipt")
	ErrNonPositiveTransferValue = errors.New("value must be positive")
	ErrNegativeValue            = errors.New("value cannot be negative")

	ErrEmptyResult = errors.New("result is empty")

	// RingSignature related
	ErrInsufficientRingSignMember = errors.New("member of ring signature is insufficient (require more than 1)")
	ErrMembersExcludeSelf         = errors.New("sender does not be contained in members")
	ErrMembersNotFound            = errors.New("cannot found all members in database")
	ErrEmptySignature             = errors.New("signature is empty")
	ErrUnmatchedSignLenAndMembers = errors.New("length of signature and members is unmatched")
	ErrEmptyRingSignTransaction   = errors.New("transaction is empty")

	ErrNoSuchAuditor = errors.New("no such auditor")

	// EmptyAddressHex represents a empty address in hex
	EmptyAddressHex = "0x0000000000000000000000000000000000000000"
)
