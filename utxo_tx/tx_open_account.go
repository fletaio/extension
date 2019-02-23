package utxo_tx

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
	data.RegisterTransaction("fleta.OpenAccount", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &OpenAccount{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
				Vin: []*transaction.TxIn{},
			},
			Vout: []*transaction.TxOut{},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*OpenAccount)
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
		tx := t.(*OpenAccount)

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
		addr := common.NewAddress(coord, chainCoord, 0)
		if is, err := ctx.IsExistAccount(addr); err != nil {
			return nil, err
		} else if is {
			return nil, ErrExistAddress
		} else {
			a, err := ctx.Accounter().NewByTypeName("fleta.SingleAccount")
			if err != nil {
				return nil, err
			}
			acc := a.(*account_def.SingleAccount)
			acc.Address_ = addr
			acc.KeyHash = tx.KeyHash
			ctx.CreateAccount(acc)
		}
		ctx.Commit(sn)
		return nil, nil
	})
}

// OpenAccount is a fleta.OpenAccount
// It is used to create signle account using UTXOs
type OpenAccount struct {
	Base
	Vout    []*transaction.TxOut
	KeyHash common.PublicHash
}

// Hash returns the hash value of it
func (tx *OpenAccount) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *OpenAccount) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := tx.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *OpenAccount) ReadFrom(r io.Reader) (int64, error) {
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
	if n, err := tx.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *OpenAccount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
