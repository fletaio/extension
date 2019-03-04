package token_tx

import (
	"io"
	"log"

	"github.com/fletaio/extension/account_tx"

	"github.com/fletaio/core/amount"

	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.TokenCreation", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &TokenCreation{
			Base: account_tx.Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*TokenCreation)
		if !transaction.IsMainChain(loader.ChainCoord()) {
			return ErrNotMainChain
		}
		if tx.Seq() <= loader.Seq(tx.From()) {
			return ErrInvalidSequence
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
		tx := t.(*TokenCreation)
		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return nil, ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		chainCoord := ctx.ChainCoord()
		fromBalance, err := ctx.AccountBalance(tx.From())
		if err != nil {
			return nil, err
		}
		if err := fromBalance.SubBalance(chainCoord, Fee); err != nil {
			return nil, err
		}

		addr := common.NewAddress(coord, chainCoord, 0)
		if is, err := ctx.IsExistAccount(addr); err != nil {
			return nil, err
		} else if is {
			return nil, ErrExistAddress
		}

		a, err := ctx.Accounter().NewByTypeName("fleta.TokenAccount")
		if err != nil {
			return nil, err
		}
		acc := a.(*TokenAccount)
		acc.Address_ = addr
		log.Println("fleta.TokenAccount ", addr.String())
		acc.TokenCoord = *coord.Clone()
		acc.KeyHash = tx.TokenPublicHash
		err = ctx.CreateAccount(acc)
		if err != nil {
			return nil, err
		}

		ctx.Commit(sn)
		return nil, nil
	})
}

// TokenCreation is a fleta.TokenCreation
// It is used to make a single account
type TokenCreation struct {
	account_tx.Base
	TokenPublicHash common.PublicHash
}

// Hash returns the hash value of it
func (tx *TokenCreation) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *TokenCreation) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.TokenPublicHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *TokenCreation) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.TokenPublicHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}
