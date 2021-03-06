package utxo_tx

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/fletaio/core/amount"

	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.Deposit", func(t transaction.Type) transaction.Transaction {
		return &Deposit{
			Base: Base{
				Base: transaction.Base{
					Type_: t,
				},
				Vin: []*transaction.TxIn{},
			},
			Vout: []*transaction.TxOut{},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*Deposit)
		if len(tx.Vin) == 0 {
			return ErrInvalidTxInCount
		}
		if len(signers) > 1 {
			return ErrInvalidSignerCount
		}
		if tx.Amount.Less(amount.COIN.DivC(10)) {
			return ErrDustAmount
		}

		for _, vin := range tx.Vin {
			if utxo, err := loader.UTXO(vin.ID()); err != nil {
				return err
			} else {
				if !utxo.PublicHash.Equal(signers[0]) {
					return ErrInvalidTransactionSignature
				}
			}
		}

		for _, vout := range tx.Vout {
			if vout.Amount.Less(amount.COIN.DivC(10)) {
				return ErrDustAmount
			}
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*Deposit)

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		insum := amount.NewCoinAmount(0, 0)
		for _, vin := range tx.Vin {
			if utxo, err := ctx.UTXO(vin.ID()); err != nil {
				return nil, err
			} else {
				insum = insum.Add(utxo.Amount)
				if err := ctx.DeleteUTXO(vin.ID()); err != nil {
					return nil, err
				}
			}
		}

		outsum := Fee.Clone()
		outsum = outsum.Add(tx.Amount)
		for n, vout := range tx.Vout {
			if vout.Amount.Less(amount.COIN.DivC(10)) {
				return nil, ErrDustAmount
			}
			outsum = outsum.Add(vout.Amount)
			if err := ctx.CreateUTXO(transaction.MarshalID(coord.Height, coord.Index, uint16(n)), vout); err != nil {
				return nil, err
			}
		}

		if !insum.Equal(outsum) {
			return nil, ErrInvalidOutputAmount
		}

		toAcc, err := ctx.Account(tx.To)
		if err != nil {
			return nil, err
		}
		toAcc.AddBalance(tx.Amount)

		ctx.Commit(sn)
		return nil, nil
	})
}

// Deposit is a fleta.Deposit
// It is used to add balance to the account from UTXOs
type Deposit struct {
	Base
	Vout   []*transaction.TxOut
	Amount *amount.Amount
	To     common.Address
	Tag    []byte
}

// Hash returns the hash value of it
func (tx *Deposit) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *Deposit) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := tx.Amount.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.To.WriteTo(w); err != nil {
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
func (tx *Deposit) ReadFrom(r io.Reader) (int64, error) {
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
	if n, err := tx.Amount.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.To.ReadFrom(r); err != nil {
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
func (tx *Deposit) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"vin":`)
	buffer.WriteString(`[`)
	for i, vin := range tx.Vin {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(vin.ID()); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
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
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"to":`)
	if bs, err := tx.To.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"tag":`)
	if len(tx.Tag) == 0 {
		buffer.WriteString(`null`)
	} else {
		buffer.WriteString(`"`)
		buffer.WriteString(hex.EncodeToString(tx.Tag))
		buffer.WriteString(`"`)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
