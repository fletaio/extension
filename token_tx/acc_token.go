package token_tx

import (
	"io"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/data"
)

func init() {
	data.RegisterAccount("fleta.TokenAccount", func(t account.Type) account.Account {
		return &TokenAccount{
			Base: account.Base{
				Type_: t,
			},
		}
	}, func(loader data.Loader, a account.Account, signers []common.PublicHash) error {
		acc := a.(*TokenAccount)
		if len(signers) != 1 {
			return ErrInvalidSignerCount
		}
		signer := signers[0]
		if !acc.KeyHash.Equal(signer) {
			return ErrInvalidAccountSigner
		}
		return nil
	})
}

// TokenAccount is a fleta.SingleAccount
// It is used as a basic account
type TokenAccount struct {
	account.Base
	TokenCoord common.Coordinate
	KeyHash    common.PublicHash
}

// Clone returns the clonend value of it
func (acc *TokenAccount) Clone() account.Account {
	return &TokenAccount{
		Base: account.Base{
			Address_: acc.Address_,
			Type_:    acc.Type_,
		},
		TokenCoord: *acc.TokenCoord.Clone(),
		KeyHash:    acc.KeyHash.Clone(),
	}
}

// WriteTo is a serialization function
func (acc *TokenAccount) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := acc.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := acc.TokenCoord.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := acc.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (acc *TokenAccount) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := acc.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := acc.TokenCoord.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := acc.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}
