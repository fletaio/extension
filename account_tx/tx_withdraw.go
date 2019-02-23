package account_tx

import (
	"bytes"
	"encoding/json"
	"io"

	"git.fleta.io/fleta/core/amount"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.Withdraw", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &Withdraw{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
			},
			Vout: []*transaction.TxOut{},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*Withdraw)
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}

		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}
		for _, vout := range tx.Vout {
			if vout.Amount.Less(amount.COIN.DivC(10)) {
				return ErrDustAmount
			}
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*Withdraw)

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		outsum := Fee.Clone()
		for n, vout := range tx.Vout {
			outsum = outsum.Add(vout.Amount)
			if err := ctx.CreateUTXO(transaction.MarshalID(coord.Height, coord.Index, uint16(n)), vout); err != nil {
				return nil, err
			}
		}

		chainCoord := ctx.ChainCoord()
		fromBalance, err := ctx.AccountBalance(tx.From())
		if err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(chainCoord, Fee); err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(chainCoord, outsum); err != nil {
			return nil, err
		}
		ctx.Commit(sn)
		return nil, nil
	})
}

// Withdraw is a fleta.Withdraw
// It is used to make UTXO from the account
type Withdraw struct {
	Base
	Vout []*transaction.TxOut
}

// Hash returns the hash value of it
func (tx *Withdraw) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *Withdraw) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(len(tx.Vout))); err != nil {
		return wrote, err
	} else {
		wrote += n
		for _, v := range tx.Vout {
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
func (tx *Withdraw) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if Len, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Vout = make([]*transaction.TxOut, 0, Len)
		for i := 0; i < int(Len); i++ {
			vout := transaction.NewTxOut()
			if n, err := vout.ReadFrom(r); err != nil {
				return read, err
			} else {
				read += n
				tx.Vout = append(tx.Vout, vout)
			}
		}
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *Withdraw) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"vout":`)
	buffer.WriteString(`[`)
	for i, vout := range tx.Vout {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := vout.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
