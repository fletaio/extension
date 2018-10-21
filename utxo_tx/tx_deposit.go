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
	transactor.RegisterHandler("fleta.Deposit", func(t transaction.Type) transaction.Transaction {
		return &Deposit{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: &common.Coordinate{},
					Type_:       t,
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
			outsum = outsum.Add(vout.Amount)
			if err := ctx.CreateUTXO(transaction.MarshalID(coord.Height, coord.Index, uint16(n)), vout); err != nil {
				return nil, err
			}
		}

		if !insum.Equal(outsum) {
			return nil, ErrInvalidOutputAmount
		}

		chainCoord := ctx.ChainCoord()
		toAcc, err := ctx.Account(tx.To)
		if err != nil {
			return nil, err
		}
		toBalance := toAcc.Balance(chainCoord)
		toBalance = toBalance.Add(tx.Amount)
		toAcc.SetBalance(chainCoord, toBalance)

		ctx.Commit(sn)
		return nil, nil
	})
}

// Deposit TODO
type Deposit struct {
	Base
	Vout   []*transaction.TxOut
	Amount *amount.Amount
	To     common.Address
	Tag    []byte
}

// Hash TODO
func (tx *Deposit) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo TODO
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
	if n, err := util.WriteBytes8(w, tx.Tag); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom TODO
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
			vout := &transaction.TxOut{}
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
	if bs, n, err := util.ReadBytes8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Tag = bs
	}
	return read, nil
}
