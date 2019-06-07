package token_tx

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/fletaio/extension/account_tx"

	"github.com/fletaio/core/amount"

	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/transaction"
)

func init() {
	data.RegisterTransaction("fleta.ChainInitialization", func(t transaction.Type) transaction.Transaction {
		return &ChainInitialization{
			Base: account_tx.Base{
				Base: transaction.Base{
					Type_: t,
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*ChainInitialization)
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

		acc := loader.Accounter()
		Name, err := acc.NameByType(fromAcc.Type())
		if err != nil {
			return err
		}

		if Name != "fleta.TokenAccount" {
			return ErrFromTypeMustTokenAccount
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*ChainInitialization)
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
		if err := fromAcc.SubBalance(Fee); err != nil {
			return nil, err
		}

		ctx.Commit(sn)
		return nil, nil
	})
}

// ChainInitialization is a fleta.ChainInitialization
// It is used to make a single account
type ChainInitialization struct {
	account_tx.Base
	TokenCreationInformation
}

// Hash returns the hash value of it
func (tx *ChainInitialization) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *ChainInitialization) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.TokenCreationInformation.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *ChainInitialization) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.TokenCreationInformation.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}

	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *ChainInitialization) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(tx.Type_); err != nil {
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
	buffer.WriteString(`"seq":`)
	if bs, err := json.Marshal(tx.Seq_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"from":`)
	if bs, err := tx.From_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"genesis_context_hash":`)
	if bs, err := tx.GenesisContextHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"observer_infos":`)
	if bs, err := json.Marshal(tx.ObserverInfos); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

// TokenCreationInformation is a information of token creation
type TokenCreationInformation struct {
	GenesisContextHash hash.Hash256
	ObserverInfos      []ObserverInfo
}

// ObserverInfo is a information of observer
type ObserverInfo struct {
	Hash string
	URL  string
}

// WriteTo is a serialization function
func (ti *TokenCreationInformation) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := ti.GenesisContextHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	if n, err := util.WriteUint8(w, byte(len(ti.ObserverInfos))); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	for _, v := range ti.ObserverInfos {
		if n, err := util.WriteString(w, v.Hash); err != nil {
			return wrote, err
		} else {
			wrote += n
		}
		if n, err := util.WriteString(w, v.URL); err != nil {
			return wrote, err
		} else {
			wrote += n
		}
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (ti *TokenCreationInformation) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := ti.GenesisContextHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}

	var hlen int
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		hlen = int(v)
	}

	ti.ObserverInfos = make([]ObserverInfo, hlen)
	for i := 0; i < hlen; i++ {
		hash, n, err := util.ReadString(r)
		if err != nil {
			return read, err
		} else {
			read += n
		}
		url, n, err := util.ReadString(r)
		if err != nil {
			return read, err
		} else {
			read += n
		}
		ti.ObserverInfos[i] = ObserverInfo{
			Hash: hash,
			URL:  url,
		}
	}

	return read, nil
}

// Equal returns a == b
func (ti *TokenCreationInformation) Equal(b *TokenCreationInformation) bool {
	if ti.GenesisContextHash.String() != b.GenesisContextHash.String() {
		return false
	}
	if len(ti.ObserverInfos) != len(b.ObserverInfos) {
		return false
	}

	for i, v := range ti.ObserverInfos {
		bv := b.ObserverInfos[i]
		if bv.Hash != v.Hash {
			return false
		}
		if bv.URL != v.URL {
			return false
		}
	}

	return true
}
