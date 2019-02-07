package account_tx

import (
	"io"

	"git.fleta.io/fleta/core/amount"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.Transfer", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &Transfer{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
			},
			TokenCoord: coord.Clone(),
			Amount:     amount.NewCoinAmount(0, 0),
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*Transfer)
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}
		if tx.Amount.Less(amount.COIN.DivC(10)) {
			return ErrDustAmount
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
		tx := t.(*Transfer)

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		fromBalance, err := ctx.AccountBalance(tx.From())
		if err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(tx.TokenCoord, Fee); err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(tx.TokenCoord, tx.Amount); err != nil {
			return nil, err
		}

		toBalance, err := ctx.AccountBalance(tx.To)
		if err != nil {
			return nil, err
		}
		toBalance.AddBalance(tx.TokenCoord, tx.Amount)
		ctx.Commit(sn)
		return nil, nil
	})
}

// Transfer is a fleta.Transfer
// It is used to transfer coins between accounts
type Transfer struct {
	Base
	TokenCoord *common.Coordinate
	Amount     *amount.Amount
	To         common.Address
	Tag        []byte
}

// Hash returns the hash value of it
func (tx *Transfer) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *Transfer) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.TokenCoord.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
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
func (tx *Transfer) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.TokenCoord.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
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
