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
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	transactor.RegisterHandler("fleta.CreateAccount", func(t transaction.Type) transaction.Transaction {
		return &CreateAccount{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: &common.Coordinate{},
					Type_:       t,
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*CreateAccount)
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
		}

		fromAcc, err := loader.Account(tx.From())
		if err != nil {
			return err
		}

		act, err := accounter.ByCoord(loader.ChainCoord())
		if err != nil {
			return err
		}
		if err := act.Validate(fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*CreateAccount)
		if !ctx.IsMainChain() {
			return nil, ErrNotMainChain
		}

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		fromAcc, err := ctx.Account(tx.From())
		if err != nil {
			return nil, err
		}

		chainCoord := ctx.ChainCoord()
		balance := fromAcc.Balance(chainCoord)
		if balance.Less(Fee) {
			return nil, ErrInsuffcientBalance
		}
		balance = balance.Sub(Fee)
		fromAcc.SetBalance(chainCoord, balance)

		addr := common.NewAddress(coord, chainCoord, 0)
		if is, err := ctx.IsExistAccount(addr); err != nil {
			return nil, err
		} else if is {
			return nil, ErrExistAddress
		} else {
			act, err := accounter.ByCoord(ctx.ChainCoord())
			if err != nil {
				return nil, err
			}
			a, err := act.NewByTypeName("fleta.SingleAccount")
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

// CreateAccount TODO
type CreateAccount struct {
	Base
	KeyHash common.PublicHash
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
	if n, err := tx.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}
