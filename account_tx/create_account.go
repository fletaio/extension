package account_tx

import (
	"bytes"
	"io"

	"git.fleta.io/fleta/core/accounter"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/transactor"
	"git.fleta.io/fleta/extension/account_def"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	transactor.RegisterHandler("fleta.CreateAccount", func(t transaction.Type) transaction.Transaction {
		return &CreateAccount{
			Base: transaction.Base{
				ChainCoord_: &common.Coordinate{},
				Type_:       t,
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*CreateAccount)
		fromAcc, err := loader.Account(tx.From)
		if err != nil {
			return err
		}
		if tx.Seq <= fromAcc.Seq() {
			return ErrInvalidSequence
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
		tx := t.(*CreateAccount)
		if !ctx.IsMainChain() {
			return ErrNotMainChain
		}

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		fromAcc, err := ctx.Account(tx.From)
		if err != nil {
			return err
		}
		if tx.Seq != fromAcc.Seq()+1 {
			return ErrInvalidSequence
		}

		chainCoord := ctx.ChainCoord()
		balance := fromAcc.Balance(chainCoord)
		if balance.Less(Fee) {
			return ErrInsuffcientBalance
		}
		balance = balance.Sub(Fee)

		addr := common.NewAddress(coord, chainCoord, 0)
		if is, err := ctx.IsExistAccount(addr); err != nil {
			return err
		} else if is {
			return ErrExistAddress
		} else {
			act, err := accounter.ByCoord(ctx.ChainCoord())
			if err != nil {
				return err
			}
			a, err := act.NewByTypeName("fleta.SingleAccount")
			if err != nil {
				return err
			}
			acc := a.(*account_def.SingleAccount)
			acc.Address_ = addr
			acc.KeyHash = tx.KeyHash
			ctx.CreateAccount(acc)
		}
		fromAcc.SetBalance(chainCoord, balance)
		ctx.AddSeq(fromAcc)
		ctx.Commit(sn)
		return nil
	})
}

// CreateAccount TODO
type CreateAccount struct {
	transaction.Base
	Seq     uint64
	From    common.Address //MAXLEN : 255
	KeyHash common.PublicHash
}

// IsUTXO TODO
func (tx *CreateAccount) IsUTXO() bool {
	return false
}

// Bucket TODO
func (tx *CreateAccount) Bucket() common.Address {
	return tx.From
}

// BucketOrder TODO
func (tx *CreateAccount) BucketOrder() uint64 {
	return tx.Seq
}

// Hash TODO
func (tx *CreateAccount) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo TODO
func (tx *CreateAccount) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := tx.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom TODO
func (tx *CreateAccount) ReadFrom(r io.Reader) (int64, error) {
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
	if n, err := tx.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}
