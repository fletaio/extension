package utxo_tx

import (
	"io"

	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/transaction"
)

// Base is the parts of UTXO model based transaction functions that are not changed by derived one
type Base struct {
	transaction.Base
	Vin []*transaction.TxIn
}

// IsUTXO returns true
func (tx *Base) IsUTXO() bool {
	return true
}

// VinIDs returns ids of the vin
func (tx *Base) VinIDs() []uint64 {
	vins := make([]uint64, 0, len(tx.Vin))
	for _, vin := range tx.Vin {
		vins = append(vins, vin.ID())
	}
	return vins
}

// WriteTo is a serialization function
func (tx *Base) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(len(tx.Vin))); err != nil {
		return wrote, err
	} else {
		wrote += n
		for _, v := range tx.Vin {
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
func (tx *Base) ReadFrom(r io.Reader) (int64, error) {
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
		tx.Vin = make([]*transaction.TxIn, 0, Len)
		for i := 0; i < int(Len); i++ {
			vin := &transaction.TxIn{}
			if n, err := vin.ReadFrom(r); err != nil {
				return read, err
			} else {
				read += n
				tx.Vin = append(tx.Vin, vin)
			}
		}
	}
	return read, nil
}
