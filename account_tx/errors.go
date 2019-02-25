package account_tx

import (
	"errors"
)

// account_tx errors
var (
	ErrInvalidSequence             = errors.New("invalid sequence")
	ErrInsuffcientBalance          = errors.New("insufficient balance")
	ErrExistAddress                = errors.New("exist address")
	ErrExistAccountName            = errors.New("exist account name")
	ErrInvalidAccountName          = errors.New("invalid account name")
	ErrInvalidTransactionSignature = errors.New("invalid transaction signature")
	ErrInvalidMultiSigKeyHashCount = errors.New("invalid multisig key hash count")
	ErrNotMainChain                = errors.New("not main chain")
	ErrDustAmount                  = errors.New("dust amount")
)
