package account_tx

import (
	"bytes"
	"io"

	"git.fleta.io/fleta/core/accounter"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/transactor"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	transactor.RegisterHandler("fleta.Transfer", func(t transaction.Type) transaction.Transaction {
		return &Transfer{
			Base: transaction.Base{
				ChainCoord_: &common.Coordinate{},
				Type_:       t,
			},
			TokenCoord: &common.Coordinate{},
			Amount:     amount.NewCoinAmount(0, 0),
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*Transfer)
		fromAcc, err := loader.Account(tx.From)
		if err != nil {
			return err
		}
		if tx.Seq <= fromAcc.Seq() {
			return ErrInvalidSequence
		}
		if tx.Amount.Less(amount.COIN.DivC(10)) {
			return ErrDustAmount
		}

		act, err := accounter.ByCoord(loader.ChainCoord())
		if err != nil {
			return err
		}
		if err := act.Validate(fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) error {
		tx := t.(*Transfer)

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		fromAcc, err := ctx.Account(tx.From)
		if err != nil {
			return err
		}
		if tx.Seq != fromAcc.Seq()+1 {
			return ErrInvalidSequence
		}

		fromBalance := fromAcc.Balance(tx.TokenCoord)
		if fromBalance.Less(Fee) {
			return ErrInsuffcientBalance
		}
		fromBalance = fromBalance.Sub(Fee)

		if fromBalance.Less(tx.Amount) {
			return ErrInsuffcientBalance
		}
		fromBalance = fromBalance.Sub(tx.Amount)

		toAcc, err := ctx.Account(tx.To)
		if err != nil {
			return err
		}
		toBalance := toAcc.Balance(tx.TokenCoord)
		toBalance = toBalance.Add(tx.Amount)
		toAcc.SetBalance(tx.TokenCoord, toBalance)

		fromAcc.SetBalance(tx.TokenCoord, fromBalance)
		ctx.AddSeq(fromAcc)
		ctx.Commit(sn)
		return nil
	})
}

// Transfer TODO
type Transfer struct {
	transaction.Base
	Seq        uint64
	From       common.Address //MAXLEN : 255
	TokenCoord *common.Coordinate
	Amount     *amount.Amount
	To         common.Address
}

// IsUTXO TODO
func (tx *Transfer) IsUTXO() bool {
	return false
}

// Bucket TODO
func (tx *Transfer) Bucket() common.Address {
	return tx.From
}

// BucketOrder TODO
func (tx *Transfer) BucketOrder() uint64 {
	return tx.Seq
}

// Hash TODO
func (tx *Transfer) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo TODO
func (tx *Transfer) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, tx.Seq); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.From.WriteTo(w); err != nil {
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
	return wrote, nil
}

// ReadFrom TODO
func (tx *Transfer) ReadFrom(r io.Reader) (int64, error) {
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
		tx.Seq = v
	}
	if n, err := tx.From.ReadFrom(r); err != nil {
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
	return read, nil
}
