package dapp_chain

import (
	"encoding/hex"
	"log"
	"os"
	"testing"
	"time"

	"git.fleta.io/fleta/framework/peer"
	"git.fleta.io/fleta/framework/router"
	"git.fleta.io/fleta/framework/router/evilnode"

	"git.fleta.io/fleta/extension/account_def"
	"git.fleta.io/fleta/extension/token_tx"
	_ "git.fleta.io/fleta/extension/utxo_tx"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/block"
	"git.fleta.io/fleta/core/chain"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/generator"
	"git.fleta.io/fleta/core/kernel"
	"git.fleta.io/fleta/core/key"
	"git.fleta.io/fleta/core/observer"
	"git.fleta.io/fleta/core/store"
	"git.fleta.io/fleta/core/transaction"
)

func Test_dapp_chain(t *testing.T) {
	os.RemoveAll("./_data")

	t.Run("dapp_test", func(t *testing.T) {
		pHash1 := "Qdd6uH7cz4GJmgNAAXPZwSFDRd2CLhSh5c6x24z3eS"
		ObserverPhashs := []token_tx.ObserverInfo{
			{Hash: "3usuNFwAzsMYCjgExThqRn8NPziNdD5rNoqM5xuUn7d", URL: "opserver_0"},
			{Hash: "4ntZxpT6QHiuhULDvEMyUvCFjx3qZhYXNTeBdKXyMef", URL: "opserver_1"},
			{Hash: "2dvU82rd2h175pqanTWfVamXxadgaVuJ3ED2wfesD44", URL: "opserver_2"},
			{Hash: "3gvjLyf3m1ZzC8VPv2CkjJcp5L7bBHTL8WfPxH84Y79", URL: "opserver_3"},
			{Hash: "2BCRboUG6aQDi653n47zsbQ8hpaC7nDfgTScSJr4c5Y", URL: "opserver_4"},
		}

		Addr := common.MustParseAddress("5QxWJopN3")
		Hash := common.MustParsePublicHash(pHash1)

		DApp := &DApp{
			secureKey: "f0507dc42f6ce962c85bc23770f39b33d0a89033b4e4bf7f075bde9b67605972",
		}
		DApp.InitDAppChain(Hash, Addr, ObserverPhashs)
		genHash := DApp.GenesisHash()
		DApp.RunDAppChain(Hash, Addr, genHash)
	})
}

type DappStarterEventHandler struct {
	kernel.EventHandlerBase
	DApp                     *DApp
	kn                       *kernel.Kernel
	TokenCreationInformation token_tx.TokenCreationInformation
	accountAddr              common.Address
	accountSigner            *key.MemoryKey
	pHash                    string
}

func (eh *DappStarterEventHandler) AfterAppendBlock(b *block.Block, s *block.ObserverSigned, ctx *data.Context) error {
	for i, t := range b.Transactions {
		switch tx := t.(type) {
		case *token_tx.TokenCreation:
			// if tx's coordinate == dapp chain coord
			coord := common.NewCoordinate(b.Header.Height, uint16(i))
			addr := common.NewAddress(coord, &b.Header.ChainCoord, 0)
			if addr != common.MustParseAddress("5QxWJopN3") {
				break
			}
			// check hash and genesis context hash
			if tx.KeyHash.String() != "Qdd6uH7cz4GJmgNAAXPZwSFDRd2CLhSh5c6x24z3eS" {
				break
			}

			eh.DApp.InitDAppChain(tx.KeyHash, addr, eh.TokenCreationInformation.ObserverInfos)
			genHash := eh.DApp.GenesisHash()
			eh.TokenCreationInformation.GenesisContextHash = genHash

			// dapp_chain.RunChain(tx.KeyHash, addr, ObserverPhashs)
			{
				// start CreateContract
				cc, err := eh.kn.Loader().Transactor().NewByTypeName("fleta.ChainInitialization")
				if err != nil {
					panic(err)
				}
				t := cc.(*token_tx.ChainInitialization)
				t.From_ = eh.accountAddr
				t.TokenCreationInformation = eh.TokenCreationInformation
				t.Seq_ = eh.kn.Loader().Seq(eh.accountAddr) + 1

				sig0, _ := eh.accountSigner.Sign(t.Hash())
				sigs0 := []common.Signature{sig0}

				eh.kn.AddTransaction(t, sigs0)
				// end CreateContract
			}
		case *token_tx.ChainInitialization:
			//check chaininfo
			genHash := eh.DApp.GenesisHash()
			if tx.TokenCreationInformation.GenesisContextHash != genHash {
				break
			}

			if !eh.TokenCreationInformation.Equal(&tx.TokenCreationInformation) {
				break
			}

			go eh.DApp.RunDAppChain(common.MustParsePublicHash(eh.pHash), common.MustParseAddress("5QxWJopN3"), genHash)
		}
	}

	return nil
}

func Test_JStest(t *testing.T) {
	os.RemoveAll("./_data")

	t.Run("dapp_test", func(t *testing.T) {
		SeedNodes := []string{"formulator_0:3000"}

		ObserverPhashs := []string{
			"3e5PNobd577YEdjeb59zG6N7BBZbyRKMja2s55QQMQE",
			"4ry8UmsCbo1BPTmUhqjMWgy9UtLDFcebEsc4Lgjo9ba",
			"4KT4crmdp5GDihPXufUonmAujjC1YH4viej8C1udjc4",
			"3suqtMQWdUFUwDGMzH53KRJPFzqaP6YYRqRGMPLhhp5",
			"3HLUjZYeUDc7nGKqCaRyqB8yJHwj3BgFMWUYNmMSYfE",
		}
		obSks := []string{
			"cca49818f6c49cf57b6c420cdcd98fcae08850f56d2ff5b8d287fddc7f9ede08",
			"39f1a02bed5eff3f6247bb25564cdaef20d410d77ef7fc2c0181b1d5b31ce877",
			"2b97bc8f21215b7ed085cbbaa2ea020ded95463deef6cbf31bb1eadf826d4694",
			"3b43d728deaa62d7c8790636bdabbe7148a6641e291fd1f94b157673c0172425",
			"e6cf2724019000a3f703db92829ecbd646501c0fd6a5e97ad6774d4ad621f949",
		}

		MainChainCoord := common.NewCoordinate(0, 0)

		sk0 := "13db949719b42eac09a8d7eeb7d9d259d595657f810c50aeb249250483652f98"
		pHash0 := "2xASBuEWw6LcQGjYxeGZH9w1DUsEDt7fvUh8p3auxyN"
		kn0Addr := common.NewAddress(common.NewCoordinate(0, 1), MainChainCoord, 0)

		accountSk := "5e0dc680d12a728f60a708dcdbfb8d2c2aaea3ee5748d12bd9358f1015e3d18b"
		// accountPHash := "3An3VbzXozCehPiWYTCmngK6NQxokSruhgUUXurEgWa"
		accountAddr := common.NewAddress(common.NewCoordinate(0, 2), MainChainCoord, 0)
		var accountSigner *key.MemoryKey
		{
			data0, err := hex.DecodeString(accountSk)
			if err != nil {
				panic(err)
			}
			accountSigner, err = key.NewMemoryKeyFromBytes(data0)
			if err != nil {
				panic(err)
			}
		}

		sk1 := "ef475a14258d0a6f061293628e299a78e6abd7d46f0eb544c473045c84dffa31"
		pHash1 := "Qdd6uH7cz4GJmgNAAXPZwSFDRd2CLhSh5c6x24z3eS"
		kn1Addr := common.MustParseAddress("5QxWJopN3")

		PublicHashs := []string{
			pHash0,
		}

		ObConfig := GetConfig("observer_0", SeedNodes, "", MainChainCoord, "observer_0", ObserverPhashs)
		obKernel := CreateKernel("observer_0", ObConfig, PublicHashs)
		ob, err := NewObserver(&ObConfig.Observer, obKernel, obSks)
		if err != nil {
			panic(err)
		}
		go ob.Start()

		FmConfig0 := GetConfig("formulator_0", SeedNodes, kn0Addr.String(), MainChainCoord, "observer_0", ObserverPhashs)
		kn0 := CreateKernel("formulator_0", FmConfig0, PublicHashs)
		SetupFormulator(kn0, sk0)
		FmConfig1 := GetConfig("formulator_1", SeedNodes, kn1Addr.String(), MainChainCoord, "", ObserverPhashs)
		kn1 := CreateKernel("formulator_1", FmConfig1, PublicHashs)
		SetupFormulator(kn1, sk1)

		kn0.Start()
		kn1.Start()

		kn0.PeerManager.EnforceConnect()
		kn1.PeerManager.EnforceConnect()

		time.Sleep(time.Second)

		go kn0.TryGenerateBlock()

		dapp := &DApp{
			secureKey:  "f0507dc42f6ce962c85bc23770f39b33d0a89033b4e4bf7f075bde9b67605972",
			DAppRouter: kn1.Router,
		}

		kn0.AddEventHandler(&DappStarterEventHandler{
			kn:   kn0,
			DApp: dapp,
			TokenCreationInformation: token_tx.TokenCreationInformation{
				ObserverInfos: []token_tx.ObserverInfo{
					{Hash: "bqanMFRvFDywQgy6bgSeduG8qfhPTG7p54k55rnknL", URL: "opserver_0"},
					{Hash: "3An3VbzXozCehPiWYTCmngK6NQxokSruhgUUXurEgWa", URL: "opserver_1"},
					{Hash: "48EQBqbezsxM35rvSt44ECgZYsSENJ3qDgNmGsrmqNs", URL: "opserver_2"},
					{Hash: "27ZEfGy339mPF8coWxRrFa64bspwj2Ymf6U9hK8iuff", URL: "opserver_3"},
					{Hash: "2Yy3PMi1Hkk5hNxUhJfV8tMoudvGmtdfTnEHY7fg5Pb", URL: "opserver_4"},
				},
			},
			accountAddr:   accountAddr,
			accountSigner: accountSigner,
			pHash:         pHash1,
		})

		time.Sleep(time.Millisecond * 300)

		{
			// start CreateContract
			cc, err := kn0.Loader().Transactor().NewByTypeName("fleta.TokenCreation")
			if err != nil {
				panic(err)
			}
			t := cc.(*token_tx.TokenCreation)
			t.From_ = accountAddr
			t.KeyHash = common.MustParsePublicHash(pHash1)
			t.Seq_ = kn0.Loader().Seq(accountAddr) + 1

			sig0, _ := accountSigner.Sign(t.Hash())
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

func initMainGenesisContextData(st *store.Store, ctd *data.ContextData, PublicHashs []string) error {
	acg := &accMainCoordGenerator{idx: 1}
	for _, PublicHash := range PublicHashs {
		AddFormulator(st, ctd, common.MustParsePublicHash(PublicHash), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
	}
	// 5e0dc680d12a728f60a708dcdbfb8d2c2aaea3ee5748d12bd9358f1015e3d18b : 3An3VbzXozCehPiWYTCmngK6NQxokSruhgUUXurEgWa
	addSingleAccount(st, ctd, common.MustParsePublicHash("3An3VbzXozCehPiWYTCmngK6NQxokSruhgUUXurEgWa"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
	// addFormulator(st, ctd, common.MustParsePublicHash("2VdGunZe8yZNm2mErqQqrFx2B7Mb4SBRPWviWnapahw"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
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

type accMainCoordGenerator struct {
	idx uint16
}

func (acg *accMainCoordGenerator) Generate() *common.Coordinate {
	coord := common.NewCoordinate(0, acg.idx)
	acg.idx++
	return coord
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

func CreateKernel(ID string, Config *kernel.Config, PublicHashs []string) *kernel.Kernel {
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

	if err := initMainGenesisContextData(st, GenesisContextData, PublicHashs); err != nil {
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
