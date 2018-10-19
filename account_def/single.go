package account_def

import (
	"io"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/accounter"
	"git.fleta.io/fleta/core/amount"
)

func init() {
	accounter.RegisterHandler("fleta.SingleAccount", func(t account.Type) account.Account {
		return &SingleAccount{
			Base: account.Base{
				Type_:       t,
				BalanceHash: map[uint64]*amount.Amount{},
			},
		}
	}, func(a account.Account, signers []common.PublicHash) error {
		acc := a.(*SingleAccount)
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

// SingleAccount TODO
type SingleAccount struct {
	account.Base
	KeyHash common.PublicHash
}

// Clone TODO
func (acc *SingleAccount) Clone() account.Account {
	balanceHash := map[uint64]*amount.Amount{}
	for k, v := range acc.BalanceHash {
		balanceHash[k] = v.Clone()
	}
	return &SingleAccount{
		Base: account.Base{
			Address_:    acc.Address_,
			Type_:       acc.Type_,
			Seq_:        acc.Seq_,
			BalanceHash: balanceHash,
		},
		KeyHash: acc.KeyHash.Clone(),
	}
}

// WriteTo TODO
func (acc *SingleAccount) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := acc.Base.WriteTo(w); err != nil {
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

// ReadFrom TODO
func (acc *SingleAccount) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := acc.Base.ReadFrom(r); err != nil {
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
