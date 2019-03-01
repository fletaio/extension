package account_def

import (
	"bytes"
	"encoding/json"
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
				Type_:    t,
				Balance_: amount.NewCoinAmount(0, 0),
			},
		}
	}, func(loader data.Loader, a account.Account, signers []common.PublicHash) error {
		acc := a.(*MultiSigAccount)
		if len(signers) <= 1 || len(signers) >= 255 {
			return ErrInvalidSignerCount
		}
		signerMap := map[common.PublicHash]bool{}
		for _, signer := range signers {
			signerMap[signer] = true
		}
		matchCount := 0
		for _, addr := range acc.KeyHashes {
			if signerMap[addr] {
				matchCount++
			}
		}
		if matchCount != int(acc.Required) {
			return ErrInvalidAccountSigner
		}
		return nil
	})
}

// MultiSigAccount is a fleta.MultiSigAccount
// It is used to sign transaction using multiple keys
type MultiSigAccount struct {
	account.Base
	Required  uint8
	KeyHashes []common.PublicHash
}

// Clone returns the clonend value of it
func (acc *MultiSigAccount) Clone() account.Account {
	keyHashes := make([]common.PublicHash, 0, len(acc.KeyHashes))
	for _, v := range acc.KeyHashes {
		keyHashes = append(keyHashes, v.Clone())
	}
	return &MultiSigAccount{
		Base: account.Base{
			Address_: acc.Address_,
			Type_:    acc.Type_,
		},
		Required:  acc.Required,
		KeyHashes: keyHashes,
	}
}

// WriteTo is a serialization function
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

// ReadFrom is a deserialization function
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

// MarshalJSON is a marshaler function
func (acc *MultiSigAccount) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"address":`)
	if bs, err := acc.Address_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(acc.Type_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"required":`)
	if bs, err := json.Marshal(acc.Required); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hashes":`)
	buffer.WriteString(`[`)
	for i, pubhash := range acc.KeyHashes {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := pubhash.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
