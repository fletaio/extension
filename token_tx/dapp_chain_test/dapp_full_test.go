package dappChainTest

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/fletaio/framework/peer"
	"github.com/fletaio/framework/router"
	"github.com/fletaio/framework/router/evilnode"

	"github.com/fletaio/extension/account_def"
	"github.com/fletaio/extension/account_tx"
	"github.com/fletaio/extension/token_tx"
	_ "github.com/fletaio/extension/utxo_tx"

	"github.com/fletaio/common"
	"github.com/fletaio/core/account"
	"github.com/fletaio/core/amount"
	"github.com/fletaio/core/block"
	"github.com/fletaio/core/chain"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/generator"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/core/key"
	"github.com/fletaio/core/observer"
	"github.com/fletaio/core/store"
	"github.com/fletaio/core/transaction"
)

func Test_dapp_chain(t *testing.T) {
	os.RemoveAll("./_data")

	t.Run("dapp_test", func(t *testing.T) {
		ObserverPhashs := []token_tx.ObserverInfo{
			{Hash: ADDR.DAppObserver[0].Hash, URL: "opserver_0"},
			{Hash: ADDR.DAppObserver[1].Hash, URL: "opserver_1"},
			{Hash: ADDR.DAppObserver[2].Hash, URL: "opserver_2"},
			{Hash: ADDR.DAppObserver[3].Hash, URL: "opserver_3"},
			{Hash: ADDR.DAppObserver[4].Hash, URL: "opserver_4"},
		}

		dappInit(DappCoor)
		addr := common.NewAddress(DappCoor, MainCoor, 0)

		log.Println(addr.String())
		Addr := common.MustParseAddress("5QxWJopN3")
		// Hash := common.MustParsePublicHash(ADDR.MainTokenHash)

		DApp := &DApp{}
		DApp.InitDAppChain(Addr, ObserverPhashs)
		genHash := DApp.GenesisHash()
		DApp.RunDAppChain(Addr, genHash)
	})
}

type DappStarterEventHandler struct {
	kernel.EventHandlerBase
	DApp                     *DApp
	TokenCreationInformation token_tx.TokenCreationInformation
	accountAddr              common.Address
	accountSigner            *key.MemoryKey
	TokenPublicHash          string
}

func (eh *DappStarterEventHandler) AfterAppendBlock(b *block.Block, s *block.ObserverSigned, ctx *data.Context) error {
	for i, t := range b.Transactions {
		switch tx := t.(type) {
		case *token_tx.TokenCreation:
			// if tx's coordinate == dapp chain coord
			coord := common.NewCoordinate(b.Header.Height, uint16(i))
			addr := common.NewAddress(coord, &b.Header.ChainCoord, 0)
			// tx주소가 dapp의 chain 주소가 됨
			if addr != common.MustParseAddress("5QxWJopN3") {
				break
			}
			ADDR.MainToken.Addr = addr
			// check hash and genesis context hash
			if tx.TokenPublicHash.String() != eh.TokenPublicHash {
				break
			}
			go func(coord *common.Coordinate, tx *token_tx.TokenCreation) {
				dappInit(coord)
				log.Println("coord.Index ", coord.Index, "coord.Height", coord.Height)

				eh.DApp.InitDAppChain(addr, eh.TokenCreationInformation.ObserverInfos)
				genHash := eh.DApp.GenesisHash()
				eh.TokenCreationInformation.GenesisContextHash = genHash

				// dapp_chain.RunChain(tx.KeyHash, addr, ObserverPhashs)
				Seq := eh.DApp.MainKn.Loader().Seq(eh.accountAddr) + 1
				{
					// start CreateContract
					cc, err := eh.DApp.MainKn.Loader().Transactor().NewByTypeName("fleta.ChainInitialization")
					if err != nil {
						panic(err)
					}
					t := cc.(*token_tx.ChainInitialization)
					t.From_ = eh.accountAddr
					t.TokenCreationInformation = eh.TokenCreationInformation
					t.Seq_ = Seq
					Seq++

					sig0, _ := eh.accountSigner.Sign(t.Hash())
					sigs0 := []common.Signature{sig0}

					eh.DApp.MainKn.AddTransaction(t, sigs0)
					// end CreateContract
				}
				{
					printBalance(eh.DApp.MainKn, MainCoor, ADDR.MainAccount.Addr, ADDR.MainToken.Addr)
					// start Transfer
					cc, err := eh.DApp.MainKn.Loader().Transactor().NewByTypeName("fleta.Transfer")
					if err != nil {
						panic(err)
					}
					t := cc.(*account_tx.Transfer)

					t.Seq_ = Seq
					t.From_ = ADDR.MainAccount.Addr
					t.To = ADDR.MainToken.Addr
					t.Amount = amount.NewCoinAmount(500000, 0)

					sig1, _ := ADDR.MainAccount.Signer.Sign(t.Hash())
					sigs1 := []common.Signature{sig1}

					eh.DApp.MainKn.AddTransaction(t, sigs1)
					// end Transfer
					printBalance(eh.DApp.MainKn, MainCoor, ADDR.MainAccount.Addr, ADDR.MainToken.Addr)
				}
			}(coord, tx)

		case *token_tx.ChainInitialization:
			//check chaininfo
			go func(tx *token_tx.ChainInitialization) {
				genHash := eh.DApp.GenesisHash()
				if tx.TokenCreationInformation.GenesisContextHash != genHash {
					return
				}

				if !eh.TokenCreationInformation.Equal(&tx.TokenCreationInformation) {
					return
				}

				go eh.DApp.RunDAppChain(common.MustParseAddress("5QxWJopN3"), genHash)
			}(tx)
		}
	}

	return nil
}

func Test_JStest(t *testing.T) {
	os.RemoveAll("./_data")

	t.Run("dapp_test", func(t *testing.T) {
		SeedNodes := []string{"formulator_0:3000"}

		MainChainCoord := common.NewCoordinate(0, 0)

		obsk := make([]string, len(ADDR.MainObserver))
		obHash := make([]string, len(ADDR.MainObserver))
		for i, k := range ADDR.MainObserver {
			obsk[i] = k.SK
			obHash[i] = k.Hash
		}

		ObConfig := GetConfig("observer_0", SeedNodes, "", MainChainCoord, "observer_0", obHash)
		obKernel := CreateKernel("observer_0", ObConfig)
		ob, err := NewObserver(&ObConfig.Observer, obKernel, obsk)
		if err != nil {
			panic(err)
		}
		go ob.Start()

		FmConfig0 := GetConfig("formulator_0", SeedNodes, ADDR.MainFormulator[0].Addr.String(), MainChainCoord, "observer_0", obHash)
		kn0 := CreateKernel("formulator_0", FmConfig0)
		SetupFormulator(kn0, ADDR.MainFormulator[0].SK)

		kn0.Start()

		kn0.PeerManager.EnforceConnect()

		time.Sleep(time.Second)

		go kn0.TryGenerateBlock()

		dapp := &DApp{
			DAppRouter: kn0.Router,
			MainKn:     kn0,
		}

		kn0.AddEventHandler(&DappStarterEventHandler{
			DApp: dapp,
			TokenCreationInformation: token_tx.TokenCreationInformation{
				ObserverInfos: []token_tx.ObserverInfo{
					{Hash: ADDR.DAppObserver[0].Hash, URL: "opserver_0"},
					{Hash: ADDR.DAppObserver[1].Hash, URL: "opserver_1"},
					{Hash: ADDR.DAppObserver[2].Hash, URL: "opserver_2"},
					{Hash: ADDR.DAppObserver[3].Hash, URL: "opserver_3"},
					{Hash: ADDR.DAppObserver[4].Hash, URL: "opserver_4"},
				},
			},
			accountAddr:     ADDR.MainAccount.Addr,
			accountSigner:   ADDR.MainAccount.Signer,
			TokenPublicHash: ADDR.MainToken.Hash,
		})

		time.Sleep(time.Millisecond * 300)

		{
			// start CreateContract
			cc, err := kn0.Loader().Transactor().NewByTypeName("fleta.TokenCreation")
			if err != nil {
				panic(err)
			}
			t := cc.(*token_tx.TokenCreation)
			t.From_ = ADDR.MainAccount.Addr
			t.TokenPublicHash = common.MustParsePublicHash(ADDR.MainToken.Hash)
			t.Seq_ = kn0.Loader().Seq(ADDR.MainAccount.Addr) + 1

			sig0, _ := ADDR.MainAccount.Signer.Sign(t.Hash())
			sigs0 := []common.Signature{sig0}

			kn0.AddTransaction(t, sigs0)
			// end CreateContract
		}

		select {}
	})
}

func initMainChainComponent(act *data.Accounter, tran *data.Transactor) error {
	// transaction_type transaction types
	const (
		// FLETA Transactions
		TransferTransctionType              = transaction.Type(10)
		WithdrawTransctionType              = transaction.Type(18)
		BurnTransctionType                  = transaction.Type(19)
		CreateAccountTransctionType         = transaction.Type(20)
		CreateMultiSigAccountTransctionType = transaction.Type(21)
		// UTXO Transactions
		AssignTransctionType      = transaction.Type(30)
		DepositTransctionType     = transaction.Type(38)
		OpenAccountTransctionType = transaction.Type(41)
		// Token transaction
		TokenCreationTransctionType       = transaction.Type(50)
		ChainInitializationTransctionType = transaction.Type(51)
		TokenIssueTransctionType          = transaction.Type(52)
		EngraveDappTransctionType         = transaction.Type(53)
		// Formulation Transactions
		CreateFormulationTransctionType = transaction.Type(60)
		RevokeFormulationTransctionType = transaction.Type(61)
	)

	// account_type account types
	const (
		// FLTEA Accounts
		SingleAccountType   = account.Type(10)
		MultiSigAccountType = account.Type(11)
		TokenAccountType    = account.Type(12)
		LockedAccountType   = account.Type(19)
		// Formulation Accounts
		FormulationAccountType = account.Type(60)
	)

	type txFee struct {
		Type transaction.Type
		Fee  *amount.Amount
	}

	TxFeeTable := map[string]*txFee{
		"fleta.CreateAccount":         &txFee{CreateAccountTransctionType, amount.COIN.MulC(10)},
		"fleta.CreateMultiSigAccount": &txFee{CreateMultiSigAccountTransctionType, amount.COIN.MulC(10)},
		"fleta.Transfer":              &txFee{TransferTransctionType, amount.COIN.DivC(10)},
		"fleta.Withdraw":              &txFee{WithdrawTransctionType, amount.COIN.DivC(10)},
		"fleta.Burn":                  &txFee{BurnTransctionType, amount.COIN.DivC(10)},
		"fleta.Assign":                &txFee{AssignTransctionType, amount.COIN.DivC(2)},
		"fleta.Deposit":               &txFee{DepositTransctionType, amount.COIN.DivC(2)},
		"fleta.OpenAccount":           &txFee{OpenAccountTransctionType, amount.COIN.MulC(10)},
		"fleta.TokenCreation":         &txFee{TokenCreationTransctionType, amount.COIN.MulC(10)},
		"fleta.ChainInitialization":   &txFee{ChainInitializationTransctionType, amount.COIN.MulC(10)},
		"fleta.TokenIssue":            &txFee{TokenIssueTransctionType, amount.COIN.MulC(10)},
		"fleta.EngraveDapp":           &txFee{EngraveDappTransctionType, amount.COIN.MulC(10)},
		"consensus.CreateFormulation": &txFee{CreateFormulationTransctionType, amount.COIN.MulC(50000)},
		"consensus.RevokeFormulation": &txFee{RevokeFormulationTransctionType, amount.COIN.DivC(10)},
	}
	for name, item := range TxFeeTable {
		if err := tran.RegisterType(name, item.Type, item.Fee); err != nil {
			log.Println(name, item, err)
			return err
		}
	}

	AccTable := map[string]account.Type{
		"fleta.SingleAccount":          SingleAccountType,
		"fleta.MultiSigAccount":        MultiSigAccountType,
		"fleta.TokenAccount":           TokenAccountType,
		"fleta.LockedAccount":          LockedAccountType,
		"consensus.FormulationAccount": FormulationAccountType,
	}
	for name, t := range AccTable {
		if err := act.RegisterType(name, t); err != nil {
			log.Println(name, t, err)
			return err
		}
	}
	return nil
}

func initMainGenesisContextData(st *store.Store, ctd *data.ContextData) error {
	for i, k := range ADDR.MainFormulator {
		PublicHash := k.Hash
		addr := ADDR.MainFormulator[i].Addr
		AddFormulator(st, ctd, common.MustParsePublicHash(PublicHash), addr)
	}

	addSingleAccount(st, ctd, common.MustParsePublicHash(ADDR.MainAccount.Hash), ADDR.MainAccount.Addr)
	return nil
}

func addSingleAccount(st *store.Store, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address) {
	a, err := st.Accounter().NewByTypeName("fleta.SingleAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*account_def.SingleAccount)
	acc.Address_ = addr
	acc.KeyHash = KeyHash
	ctd.CreatedAccountHash[acc.Address_] = acc
	balance := account.NewBalance()
	balance.AddBalance(st.ChainCoord(), amount.NewCoinAmount(10000000000, 0))
	ctd.AccountBalanceHash[acc.Address_] = balance
}

func GetConfig(ID string, SeedNodes []string, Generator string, ChainCoord *common.Coordinate, obID string, ObserverSignatures []string) *kernel.Config {
	Config := &kernel.Config{
		ChainCoord: ChainCoord,
		SeedNodes:  SeedNodes,
		Chain: chain.Config{
			Version:            1,
			ObserverSignatures: ObserverSignatures,
		},
		Router: router.Config{
			Network: "mock:" + ID,
			Port:    3000,
			EvilNodeConfig: evilnode.Config{
				BanEvilScore: 100,
				StorePath:    "./_data/" + ID + "/router",
			},
		},
		Peer: peer.Config{
			StorePath: "./_data/" + ID + "/peers",
		},
		Generator: generator.Config{
			Address:           Generator,
			BlockVersion:      1,
			GenTimeThreshold:  200,
			ObserverAddresses: []string{},
		},
		Observer: observer.Config{
			ID:        obID,
			Network:   "mock:connector_" + ID,
			Addresses: []string{obID + ":3001"},
		},
		StorePath: "./_data/" + ID,
	}

	return Config
}

func CreateKernel(ID string, Config *kernel.Config) *kernel.Kernel {
	os.RemoveAll("./_data/" + ID)
	act := data.NewAccounter(Config.ChainCoord)
	tran := data.NewTransactor(Config.ChainCoord)
	if err := initMainChainComponent(act, tran); err != nil {
		panic(err)
	}

	st, err := store.NewStore(Config.StorePath, act, tran)
	if err != nil {
		panic(err)
	}

	GenesisContextData := data.NewContextData(data.NewEmptyLoader(st.ChainCoord(), st.Accounter(), st.Transactor()), nil)

	if err := initMainGenesisContextData(st, GenesisContextData); err != nil {
		panic(err)
	}
	rewarder := &Rewarder{}

	r, err := router.NewRouter(&Config.Router)
	if err != nil {
		panic(err)
	}

	kn, err := kernel.NewKernel(Config, r, st, rewarder, GenesisContextData)
	if err != nil {
		panic(err)
	}

	return kn
}
