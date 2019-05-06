package token_tx

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/fletaio/core/amount"
	"github.com/fletaio/extension/account_tx"

	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.TokenIssue", func(t transaction.Type) transaction.Transaction {
		return &TokenIssue{
			Base: account_tx.Base{
				Base: transaction.Base{
					Type_: t,
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*TokenIssue)
		if !transaction.IsMainChain(loader.ChainCoord()) {
			return ErrNotMainChain
		}
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}

		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*TokenIssue)
		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		fromAcc, err := ctx.Account(tx.From())
		if err != nil {
			return nil, err
		}
		if err := fromAcc.SubBalance(Fee); err != nil {
			return nil, err
		}
		if err := fromAcc.SubBalance(tx.Amount); err != nil {
			return nil, err
		}

		ctx.Commit(sn)
		return nil, nil
	})
}

// TokenIssue is a fleta.TokenIssue
// It is used to make a single account
type TokenIssue struct {
	account_tx.Base
	TokenAddress common.Address
	Height       uint32
	Amount       *amount.Amount
	Tag          []byte
}

// Hash returns the hash value of it
func (tx *TokenIssue) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *TokenIssue) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.TokenAddress.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, tx.Height); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.Amount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteBytes(w, tx.Tag); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *TokenIssue) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.TokenAddress.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Height = v
	}
	if n, err := tx.Amount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if bs, n, err := util.ReadBytes(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Tag = bs
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *TokenIssue) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(tx.Type_); err != nil {
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
	buffer.WriteString(`"token_address":`)
	if bs, err := tx.TokenAddress.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"height":`)
	if bs, err := json.Marshal(tx.Height); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"tag":`)
	if bs, err := json.Marshal(tx.Tag); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
