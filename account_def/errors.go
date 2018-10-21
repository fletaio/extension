package account_def

import (
	"errors"
)

// account_def errors
var (
	ErrInvalidSignerCount   = errors.New("invalid signer count")
	ErrInvalidAccountSigner = errors.New("invalid account signer")
	ErrLockedAccount        = errors.New("locked account")
)
