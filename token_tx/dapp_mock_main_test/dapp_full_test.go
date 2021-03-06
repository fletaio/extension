package dappchaintest

import (
	"encoding/binary"
	"log"
	"os"
	"testing"
	"time"

	"github.com/fletaio/common"
	"github.com/fletaio/core/amount"
	"github.com/fletaio/core/block"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/formulator"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/core/key"
	"github.com/fletaio/core/message_def"
	"github.com/fletaio/core/transaction"
	"github.com/fletaio/extension/account_tx"
	"github.com/fletaio/extension/token_tx/dapp_mock_main_test/address"
	"github.com/fletaio/extension/token_tx/dapp_mock_main_test/dappchain"
	"github.com/fletaio/extension/token_tx/dapp_mock_main_test/mainchain"

	"github.com/fletaio/extension/token_tx"
	_ "github.com/fletaio/extension/utxo_tx"
)

func Test_dapp_chain(t *testing.T) {
	os.RemoveAll("./dappchain/formulator")
	os.RemoveAll("./dappchain/observer")
	os.RemoveAll("./mainchain/formulator")
	os.RemoveAll("./mainchain/observer")

	t.Run("dapp_test", func(t *testing.T) {
		mainkn, frlist := mainchain.RunMainChain()
		mainkn.AddEventHandler(&DappStarterEventHandler{
			mainkn:          mainkn,
			accountAddr:     address.ADDR.MainAccount.Addr,
			accountSigner:   address.ADDR.MainAccount.Signer,
			TokenPublicHash: address.ADDR.MainTokenAccount.Hash,
		})

		for _, fr := range frlist {
			go func(fr *formulator.Formulator) {
				fr.Run()
			}(fr)
		}

		time.Sleep(time.Second * 5)

		{ // send TokenCreation
			cc, err := mainkn.Loader().Transactor().NewByTypeName("fleta.TokenCreation")
			if err != nil {
				panic(err)
			}
			t := cc.(*token_tx.TokenCreation)
			t.From_ = address.ADDR.MainAccount.Addr
			t.TokenName = "testName"
			t.TokenPublicHash = common.MustParsePublicHash(address.ADDR.MainTokenAccount.Hash)
			t.Seq_ = mainkn.Loader().Seq(address.ADDR.MainAccount.Addr) + 1

			sig0, _ := address.ADDR.MainAccount.Signer.Sign(t.Hash())
			sigs0 := []common.Signature{sig0}

			mainkn.AddTransaction(t, sigs0)
			// end CreateContract
		}

		select {}
	})
}

type DappStarterEventHandler struct {
	mainkn                   *kernel.Kernel
	dappkn                   *kernel.Kernel
	TokenCreationInformation token_tx.TokenCreationInformation
	accountAddr              common.Address
	accountSigner            *key.MemoryKey
	TokenPublicHash          string
	formulatorList           []*formulator.Formulator
}

func (eh *DappStarterEventHandler) AfterProcessBlock(kn *kernel.Kernel, b *block.Block, s *block.ObserverSigned, ctx *data.Context) {
	for i, t := range b.Body.Transactions {
		switch tx := t.(type) {
		case *token_tx.TokenCreation:
			// if tx's coordinate == dapp chain coord
			coord := common.NewCoordinate(b.Header.Height(), uint16(i))
			log.Println("token_tx.TokenCreation", coord.Height, coord.Index)

			addr := common.NewAddress(coord, 0)
			address.ADDR.MainTokenAccount.Addr = addr
			// check hash and genesis context hash
			if tx.TokenPublicHash.String() != eh.TokenPublicHash {
				continue
			}
			go func(addr common.Address) {
				{
					// start Transfer
					cc, err := eh.mainkn.Loader().Transactor().NewByTypeName("fleta.Transfer")
					if err != nil {
						panic(err)
					}
					t := cc.(*account_tx.Transfer)

					t.Seq_ = eh.mainkn.Loader().Seq(eh.accountAddr) + 1
					t.From_ = address.ADDR.MainAccount.Addr
					t.To = address.ADDR.MainTokenAccount.Addr
					t.Amount = amount.NewCoinAmount(500000, 0)

					sig1, _ := address.ADDR.MainAccount.Signer.Sign(t.Hash())
					sigs1 := []common.Signature{sig1}

					eh.mainkn.AddTransaction(t, sigs1)
					// end Transfer
				}
			}(addr)

		case *account_tx.Transfer:
			if tx.From_ != address.ADDR.MainAccount.Addr || tx.To != address.ADDR.MainTokenAccount.Addr {
				continue
			}

			Height := binary.LittleEndian.Uint32(address.ADDR.MainTokenAccount.Addr[:4])
			Index := binary.LittleEndian.Uint16(address.ADDR.MainTokenAccount.Addr[4:6])
			log.Println("account_tx.Transfer", Height, Index)
			coord := common.NewCoordinate(Height, Index)

			go func(coord *common.Coordinate) {
				address.DappInitAddr(coord)
				log.Println("coord.Index ", coord.Index, "coord.Height", coord.Height)

				dappkn, frls := dappchain.InitDappChain(coord)
				hash, err := dappkn.Provider().Hash(0)
				if err != nil {
					panic(err)
				}

				dappkn.AddEventHandler(&DappEventHandler{
					mainkn: eh.mainkn,
				})

				eh.dappkn = dappkn
				eh.formulatorList = frls

				// genHash := eh.DApp.GenesisHash()
				eh.TokenCreationInformation = token_tx.TokenCreationInformation{
					GenesisContextHash: hash,
					ObserverInfos: []token_tx.ObserverInfo{
						{Hash: address.ADDR.DAppObserver[0].Hash, URL: "opserver_0"},
						{Hash: address.ADDR.DAppObserver[1].Hash, URL: "opserver_1"},
						{Hash: address.ADDR.DAppObserver[2].Hash, URL: "opserver_2"},
						{Hash: address.ADDR.DAppObserver[3].Hash, URL: "opserver_3"},
						{Hash: address.ADDR.DAppObserver[4].Hash, URL: "opserver_4"},
					},
				}

				{
					// start CreateContract
					cc, err := eh.mainkn.Loader().Transactor().NewByTypeName("fleta.ChainInitialization")
					if err != nil {
						panic(err)
					}
					t := cc.(*token_tx.ChainInitialization)
					t.From_ = address.ADDR.MainTokenAccount.Addr
					t.TokenCreationInformation = eh.TokenCreationInformation
					t.Seq_ = eh.mainkn.Loader().Seq(address.ADDR.MainTokenAccount.Addr) + 1

					sig0, _ := address.ADDR.MainTokenAccount.Signer.Sign(t.Hash())
					sigs0 := []common.Signature{sig0}

					eh.mainkn.AddTransaction(t, sigs0)
					// end CreateContract
				}
			}(coord)

		case *token_tx.ChainInitialization:
			//check chaininfo
			go func(tx *token_tx.ChainInitialization) {
				genHash, err := eh.dappkn.Provider().Hash(0)
				if err != nil {
					panic(err)
				}
				if tx.TokenCreationInformation.GenesisContextHash != genHash {
					return
				}

				if !eh.TokenCreationInformation.Equal(&tx.TokenCreationInformation) {
					return
				}

				for _, fr := range eh.formulatorList {
					go func(fr *formulator.Formulator) {
						fr.Run()
					}(fr)
				}
			}(tx)
		}
	}

}

// OnProcessBlock called when processing a block to the chain (error prevent processing block)
func (eh *DappStarterEventHandler) OnProcessBlock(kn *kernel.Kernel, b *block.Block, s *block.ObserverSigned, ctx *data.Context) error {
	return nil
}

// OnPushTransaction called when pushing a transaction to the transaction pool (error prevent push transaction)
func (eh *DappStarterEventHandler) OnPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) error {
	return nil
}

// AfterPushTransaction called when pushed a transaction to the transaction pool
func (eh *DappStarterEventHandler) AfterPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) {
}

// DoTransactionBroadcast called when a transaction need to be broadcast
func (eh *DappStarterEventHandler) DoTransactionBroadcast(kn *kernel.Kernel, msg *message_def.TransactionMessage) {
}

// DebugLog TEMP
func (eh *DappStarterEventHandler) DebugLog(kn *kernel.Kernel, args ...interface{}) {
}

type DappEventHandler struct {
	mainkn *kernel.Kernel
}

func (eh *DappEventHandler) AfterProcessBlock(kn *kernel.Kernel, b *block.Block, s *block.ObserverSigned, ctx *data.Context) {
	height := b.Header.Height()
	if height%10 == 0 {
		{
			// start CreateContract
			cc, err := eh.mainkn.Loader().Transactor().NewByTypeName("fleta.EngraveDapp")
			if err != nil {
				panic(err)
			}
			t := cc.(*token_tx.EngraveDapp)
			t.From_ = address.ADDR.MainTokenAccount.Addr
			t.Seq_ = eh.mainkn.Loader().Seq(address.ADDR.MainTokenAccount.Addr) + 1

			t.Height = height
			t.BlockHash = b.Header.Hash()

			sig0, _ := address.ADDR.MainTokenAccount.Signer.Sign(t.Hash())
			sigs0 := []common.Signature{sig0}

			eh.mainkn.AddTransaction(t, sigs0)
			// end CreateContract
		}
	}
}

// OnProcessBlock called when processing a block to the chain (error prevent processing block)
func (eh *DappEventHandler) OnProcessBlock(kn *kernel.Kernel, b *block.Block, s *block.ObserverSigned, ctx *data.Context) error {
	return nil
}

// OnPushTransaction called when pushing a transaction to the transaction pool (error prevent push transaction)
func (eh *DappEventHandler) OnPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) error {
	return nil
}

// AfterPushTransaction called when pushed a transaction to the transaction pool
func (eh *DappEventHandler) AfterPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) {
}

// DoTransactionBroadcast called when a transaction need to be broadcast
func (eh *DappEventHandler) DoTransactionBroadcast(kn *kernel.Kernel, msg *message_def.TransactionMessage) {
}

// DebugLog TEMP
func (eh *DappEventHandler) DebugLog(kn *kernel.Kernel, args ...interface{}) {
}
