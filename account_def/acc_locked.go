package account_def

import (
	"bytes"
	"encoding/json"
	"io"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/data"
)

func init() {
	data.RegisterAccount("fleta.LockedAccount", func(t account.Type) account.Account {
		return &LockedAccount{
			Base: account.Base{
				Type_: t,
			},
		}
	}, func(loader data.Loader, a account.Account, signers []common.PublicHash) error {
		acc := a.(*LockedAccount)
		if acc.UnlockHeight > loader.TargetHeight() {
			return ErrLockedAccount
		}
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

// LockedAccount is a fleta.LockedAccount
// It is used to prevent transactions from the locked account until the unlock height
type LockedAccount struct {
	account.Base
	UnlockHeight uint32
	KeyHash      common.PublicHash
}

// Clone returns the clonend value of it
func (acc *LockedAccount) Clone() account.Account {
	return &LockedAccount{
		Base: account.Base{
			Address_: acc.Address_,
			Type_:    acc.Type_,
		},
		KeyHash: acc.KeyHash.Clone(),
	}
}

// WriteTo is a serialization function
func (acc *LockedAccount) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := acc.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, acc.UnlockHeight); err != nil {
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
func (acc *LockedAccount) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := acc.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		acc.UnlockHeight = v
	}
	if n, err := acc.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (acc *LockedAccount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"unlock_height":`)
	if bs, err := json.Marshal(acc.UnlockHeight); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := acc.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
