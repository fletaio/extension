package account_tx

import (
	"io"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/transaction"
)

// Base is the parts of account model based transaction functions that are not changed by derived one
type Base struct {
	transaction.Base
	Seq_  uint64
	From_ common.Address //MAXLEN : 255
}

// IsUTXO returns false
func (tx *Base) IsUTXO() bool {
	return false
}

// From returns the creator of the transaction
func (tx *Base) From() common.Address {
	return tx.From_
}

// Seq returns the sequence of the transaction
func (tx *Base) Seq() uint64 {
	return tx.Seq_
}

// WriteTo is a serialization function
func (tx *Base) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, tx.Seq_); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.From_.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
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
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Seq_ = v
	}
	if n, err := tx.From_.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}
