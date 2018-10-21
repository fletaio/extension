package utxo_tx

import (
	"errors"
)

// utxo_tx errors
var (
	ErrInvalidTxInCount            = errors.New("invalid txin count")
	ErrInvalidTxOutCount           = errors.New("invalid txout count")
	ErrInvalidTransactionSignature = errors.New("invalid transaction signature")
	ErrInvalidOutputAmount         = errors.New("invalid output amount")
	ErrInvalidSignerCount          = errors.New("invalid signer count")
	ErrNotMainChain                = errors.New("not main chain")
	ErrDustAmount                  = errors.New("dust amount")
	ErrExistAddress                = errors.New("exist address")
)
