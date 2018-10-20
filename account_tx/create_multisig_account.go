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
	transactor.RegisterHandler("fleta.CreateMultiSigAccount", func(t transaction.Type) transaction.Transaction {
		return &CreateMultiSigAccount{
			Base: Base{
				Base: transaction.Base{
					ChainCoord_: &common.Coordinate{},
					Type_:       t,
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*CreateMultiSigAccount)
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
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) error {
		tx := t.(*CreateMultiSigAccount)
		if !ctx.IsMainChain() {
			return ErrNotMainChain
		}

		if len(tx.KeyHashes) <= 1 {
			return ErrInvalidMultiSigKeyHashCount
		}
		keyHashMap := map[common.PublicHash]bool{}
		for _, v := range tx.KeyHashes {
			keyHashMap[v] = true
		}
		if len(keyHashMap) != len(tx.KeyHashes) {
			return ErrInvalidMultiSigKeyHashCount
		}

		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		if tx.Seq() != ctx.Seq(tx.From())+1 {
			return ErrInvalidSequence
		}
		ctx.AddSeq(tx.From())

		fromAcc, err := ctx.Account(tx.From())
		if err != nil {
			return err
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
			a, err := act.NewByTypeName("fleta.MultiSigAccount")
			if err != nil {
				return err
			}
			acc := a.(*account_def.MultiSigAccount)
			acc.Address_ = addr
			acc.KeyHashes = tx.KeyHashes
			ctx.CreateAccount(acc)
		}
		fromAcc.SetBalance(chainCoord, balance)
		ctx.Commit(sn)
		return nil
	})
}

// CreateMultiSigAccount TODO
type CreateMultiSigAccount struct {
	Base
	KeyHashes []common.PublicHash
}

// Hash TODO
func (tx *CreateMultiSigAccount) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	if _, err := tx.WriteTo(&buffer); err != nil {
		panic(err)
	}
	return hash.DoubleHash(buffer.Bytes())
}

// WriteTo TODO
func (tx *CreateMultiSigAccount) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(len(tx.KeyHashes))); err != nil {
		return wrote, err
	} else {
		wrote += n
		for _, v := range tx.KeyHashes {
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
func (tx *CreateMultiSigAccount) ReadFrom(r io.Reader) (int64, error) {
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
		tx.KeyHashes = make([]common.PublicHash, 0, Len)
		for i := 0; i < int(Len); i++ {
			var pubhash common.PublicHash
			if n, err := pubhash.ReadFrom(r); err != nil {
				return read, err
			} else {
				read += n
				tx.KeyHashes = append(tx.KeyHashes, pubhash)
			}
		}
	}
	return read, nil
}
