package utxo_tx

import (
	"bytes"
	"io"

	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/transactor"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	transactor.RegisterHandler("fleta.Assign", func(t transaction.Type) transaction.Transaction {
		return &Assign{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: &common.Coordinate{},
					Type_:       t,
				},
				Vin: []*transaction.TxIn{},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*Assign)
		if len(tx.Vin) == 0 {
			return ErrInvalidTxInCount
		}
		if len(signers) > 1 {
			return ErrInvalidSignerCount
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
		tx := t.(*Assign)

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		insum := Fee
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

		outsum := amount.NewCoinAmount(0, 0)
		for n, vout := range tx.Vout {
			outsum = outsum.Add(vout.Amount)
			if err := ctx.CreateUTXO(transaction.MarshalID(coord.Height, coord.Index, uint16(n)), vout); err != nil {
				return nil, err
			}
		}

		if insum.Less(outsum) {
			return nil, ErrInvalidOutputAmount
		}

		ctx.Commit(sn)
		return nil, nil
	})
}

// Assign TODO
type Assign struct {
	Base
	Vout []*transaction.TxOut
}

// Hash TODO
func (tx *Assign) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo TODO
func (tx *Assign) WriteTo(w io.Writer) (int64, error) {
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

// ReadFrom TODO
func (tx *Assign) ReadFrom(r io.Reader) (int64, error) {
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
			vout := &transaction.TxOut{}
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
