package token_tx

import (
	"errors"
)

// account_tx errors
var (
	ErrInvalidSequence             = errors.New("invalid sequence")
	ErrInsuffcientBalance          = errors.New("insufficient balance")
	ErrExistAddress                = errors.New("exist address")
	ErrInvalidTransactionSignature = errors.New("invalid transaction signature")
	ErrInvalidMultiSigKeyHashCount = errors.New("invalid multisig key hash count")
	ErrNotMainChain                = errors.New("not main chain")
	ErrDustAmount                  = errors.New("dust amount")
	ErrInvalidSignerCount          = errors.New("invalid signer count")
	ErrInvalidAccountSigner        = errors.New("invalid account signer")
	ErrLockedAccount               = errors.New("locked account")
	ErrFromTypeMustTokenAccount    = errors.New("only TokenAccount can initialize the chain")
)
