package account_tx

import (
	"bytes"
	"encoding/json"
	"io"

	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/extension/account_def"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.CreateAccount", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &CreateAccount{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*CreateAccount)
		if !transaction.IsMainChain(loader.ChainCoord()) {
			return ErrNotMainChain
		}
		if len(tx.Name) < 8 || len(tx.Name) > 16 {
			return ErrInvalidAccountName
		}

		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}

		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}

		if is, err := loader.IsExistAccountName(tx.Name); err != nil {
			return err
		} else if is {
			return ErrExistAccountName
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*CreateAccount)
		if len(tx.Name) < 8 || len(tx.Name) > 16 {
			return nil, ErrInvalidAccountName
		}

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		chainCoord := ctx.ChainCoord()
		fromBalance, err := ctx.AccountBalance(tx.From())
		if err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(chainCoord, Fee); err != nil {
			return nil, err
		}

		addr := common.NewAddress(coord, chainCoord, 0)
		if is, err := ctx.IsExistAccount(addr); err != nil {
			return nil, err
		} else if is {
			return nil, ErrExistAddress
		} else if isn, err := ctx.IsExistAccountName(tx.Name); err != nil {
			return nil, err
		} else if isn {
			return nil, ErrExistAccountName
		} else {
			a, err := ctx.Accounter().NewByTypeName("fleta.SingleAccount")
			if err != nil {
				return nil, err
			}
			acc := a.(*account_def.SingleAccount)
			acc.Address_ = addr
			acc.Name_ = tx.Name
			acc.KeyHash = tx.KeyHash
			ctx.CreateAccount(acc)
		}
		ctx.Commit(sn)
		return nil, nil
	})
}

// CreateAccount is a fleta.CreateAccount
// It is used to make a single account
type CreateAccount struct {
	Base
	Name    string
	KeyHash common.PublicHash
}

// Hash returns the hash value of it
func (tx *CreateAccount) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *CreateAccount) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, tx.Name); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *CreateAccount) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Name = v
	}
	if n, err := tx.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *CreateAccount) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"chain_coord":`)
	if bs, err := tx.ChainCoord_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(tx.Type_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"seq":`)
	if bs, err := json.Marshal(tx.Seq_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"from":`)
	if bs, err := tx.From_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"name":`)
	if bs, err := json.Marshal(tx.Name); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
