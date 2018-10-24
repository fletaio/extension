package account_def

import (
	"io"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"
)

func init() {
	data.RegisterAccount("fleta.MultiSigAccount", func(t account.Type) account.Account {
		return &MultiSigAccount{
			Base: account.Base{
				Type_:       t,
				BalanceHash: map[uint64]*amount.Amount{},
			},
		}
	}, func(loader data.Loader, a account.Account, signers []common.PublicHash) error {
		acc := a.(*MultiSigAccount)
		if len(signers) <= 1 || len(signers) >= 255 {
			return ErrInvalidSignerCount
		}
		signerHash := map[common.PublicHash]bool{}
		for _, signer := range signers {
			signerHash[signer] = true
		}
		matchCount := 0
		for _, addr := range acc.KeyHashes {
			if signerHash[addr] {
				matchCount++
			}
		}
		if matchCount != int(acc.Required) {
			return ErrInvalidAccountSigner
		}
		return nil
	})
}

// MultiSigAccount TODO
type MultiSigAccount struct {
	account.Base
	Required  uint8
	KeyHashes []common.PublicHash
}

// Clone TODO
func (acc *MultiSigAccount) Clone() account.Account {
	balanceHash := map[uint64]*amount.Amount{}
	for k, v := range acc.BalanceHash {
		balanceHash[k] = v.Clone()
	}
	keyHashes := make([]common.PublicHash, 0, len(acc.KeyHashes))
	for _, v := range acc.KeyHashes {
		keyHashes = append(keyHashes, v.Clone())
	}
	return &MultiSigAccount{
		Base: account.Base{
			Address_:    acc.Address_,
			Type_:       acc.Type_,
			BalanceHash: balanceHash,
		},
		Required:  acc.Required,
		KeyHashes: keyHashes,
	}
}

// WriteTo TODO
func (acc *MultiSigAccount) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := acc.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, acc.Required); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(len(acc.KeyHashes))); err != nil {
		return wrote, err
	} else {
		wrote += n
		for _, v := range acc.KeyHashes {
			if n, err := v.WriteTo(w); err != nil {
				return wrote, err
			} else {
				wrote += n
			}
		}
	}
	return wrote, nil
}

// ReadFrom TODO
func (acc *MultiSigAccount) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := acc.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		acc.Required = v
	}
	if Len, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		acc.KeyHashes = make([]common.PublicHash, 0, Len)
		for i := 0; i < int(Len); i++ {
			var pubhash common.PublicHash
			if n, err := pubhash.ReadFrom(r); err != nil {
				return read, err
			} else {
				read += n
				acc.KeyHashes = append(acc.KeyHashes, pubhash)
			}
		}
	}
	return read, nil
}
