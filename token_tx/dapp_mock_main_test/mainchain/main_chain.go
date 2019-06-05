package mainchain

import (
	"encoding/hex"
	"log"
	"strconv"

	"github.com/fletaio/common"
	"github.com/fletaio/core/account"
	"github.com/fletaio/core/amount"
	"github.com/fletaio/core/consensus"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/formulator"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/core/key"
	"github.com/fletaio/core/observer"
	"github.com/fletaio/core/transaction"
	"github.com/fletaio/extension/account_def"
	"github.com/fletaio/extension/token_tx/dapp_mock_main_test/address"
	"github.com/fletaio/framework/peer"
	"github.com/fletaio/framework/router"
	"github.com/fletaio/framework/router/evilnode"
)

// consts
const (
	CreateAccountChannelSize = 100
)

// consts
const (
	BlockchainVersion = 1
)

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

func RunMainChain() (*kernel.Kernel, []*formulator.Formulator) {
	obstrs := []string{
		"cd7cca6359869f4f58bb31aa11c2c4825d4621406f7b514058bc4dbe788c29be",
		"d8744df1e76a7b76f276656c48b68f1d40804f86518524d664b676674fccdd8a",
		"387b430fab25c03313a7e987385c81f4b027199304e2381561c9707847ec932d",
		"a99fa08114f41eb7e0a261cf11efdc60887c1d113ea6602aaf19eca5c3f5c720",
		"a9878ff3837700079fbf187c86ad22f1c123543a96cd11c53b70fedc3813c27b",
	}
	obkeys := make([]key.Key, 0, len(obstrs))
	ObserverKeys := make([]common.PublicHash, 0, len(obstrs))

	NetAddressMap := map[common.PublicHash]string{}
	NetAddressMapForFr := map[common.PublicHash]string{}
	for i, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			obkeys = append(obkeys, Key)
			Num := strconv.Itoa(i + 1)
			pubhash := common.NewPublicHash(Key.PublicKey())
			NetAddressMap[pubhash] = "127.0.0.1:300" + Num
			NetAddressMapForFr[pubhash] = "127.0.0.1:500" + Num
			ObserverKeys = append(ObserverKeys, pubhash)
		}
	}
	ObserverKeyMap := map[common.PublicHash]bool{}
	for _, pubhash := range ObserverKeys {
		ObserverKeyMap[pubhash] = true
	}

	frstrs := []string{
		"67066852dd6586fa8b473452a66c43f3ce17bd4ec409f1fff036a617bb38f063",
	}

	frkeys := make([]key.Key, 0, len(frstrs))
	for _, v := range frstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			frkeys = append(frkeys, Key)
		}
	}

	ObserverHeights := []uint32{}

	obs := []*observer.Observer{}
	for _, obkey := range obkeys {
		GenCoord := common.NewCoordinate(0, 0)
		act := data.NewAccounter(GenCoord)
		tran := data.NewTransactor(GenCoord)
		evt := data.NewEventer(GenCoord)
		if err := initChainComponent(act, tran, evt); err != nil {
			panic(err)
		}
		GenesisContextData, err := initGenesisContextData(act, tran, evt)
		if err != nil {
			panic(err)
		}

		StoreRoot := "./mainchain/observer/" + common.NewPublicHash(obkey.PublicKey()).String()

		ks, err := kernel.NewStore(StoreRoot+"/kernel", 1, act, tran, evt, true)
		if err != nil {
			panic(err)
		}

		rd := &mockRewarder{}
		kn, err := kernel.NewKernel(&kernel.Config{
			ChainCoord:              GenCoord,
			ObserverKeyMap:          ObserverKeyMap,
			MaxBlocksPerFormulator:  8,
			MaxTransactionsPerBlock: 5000,
		}, ks, rd, GenesisContextData)
		if err != nil {
			panic(err)
		}

		cfg := &observer.Config{
			ChainCoord:     GenCoord,
			Key:            obkey,
			ObserverKeyMap: NetAddressMap,
		}
		ob, err := observer.NewObserver(cfg, kn)
		if err != nil {
			panic(err)
		}
		obs = append(obs, ob)

		ObserverHeights = append(ObserverHeights, kn.Provider().Height())
	}

	Formulators := []string{}
	FormulatorHeights := []uint32{}

	frs := []*formulator.Formulator{}
	var frkn *kernel.Kernel
	for _, frkey := range frkeys {
		GenCoord := common.NewCoordinate(0, 0)
		act := data.NewAccounter(GenCoord)
		tran := data.NewTransactor(GenCoord)
		evt := data.NewEventer(GenCoord)
		if err := initChainComponent(act, tran, evt); err != nil {
			panic(err)
		}
		GenesisContextData, err := initGenesisContextData(act, tran, evt)
		if err != nil {
			panic(err)
		}

		StoreRoot := "./mainchain/formulator/" + common.NewPublicHash(frkey.PublicKey()).String()

		//os.RemoveAll(StoreRoot)

		ks, err := kernel.NewStore(StoreRoot+"/kernel", 1, act, tran, evt, true)
		if err != nil {
			panic(err)
		}

		rd := &mockRewarder{}
		kn, err := kernel.NewKernel(&kernel.Config{
			ChainCoord:              GenCoord,
			ObserverKeyMap:          ObserverKeyMap,
			MaxBlocksPerFormulator:  8,
			MaxTransactionsPerBlock: 5000,
		}, ks, rd, GenesisContextData)
		if err != nil {
			panic(err)
		}

		cfg := &formulator.Config{
			Key:            frkey,
			ObserverKeyMap: NetAddressMapForFr,
			Formulator:     common.MustParseAddress("3CUsUpvEK"),
			Router: router.Config{
				Network: "tcp",
				Port:    7000,
				EvilNodeConfig: evilnode.Config{
					StorePath: StoreRoot + "/router",
				},
			},
			Peer: peer.Config{
				StorePath: StoreRoot + "/peers",
			},
		}
		fr, err := formulator.NewFormulator(cfg, kn)
		frkn = kn
		if err != nil {
			panic(err)
		}
		frs = append(frs, fr)

		Formulators = append(Formulators, cfg.Formulator.String())
		FormulatorHeights = append(FormulatorHeights, kn.Provider().Height())

	}

	for i, ob := range obs {
		go func(BindOb string, BindFr string, ob *observer.Observer) {
			ob.Run(BindOb, BindFr)
		}(":300"+strconv.Itoa(i+1), ":500"+strconv.Itoa(i+1), ob)
	}

	// for _, fr := range frs {
	// 	go func(fr *formulator.Formulator) {
	// 		fr.Run()
	// 	}(fr)
	// }

	return frkn, frs
}

type txFee struct {
	Type transaction.Type
	Fee  *amount.Amount
}

func initChainComponent(act *data.Accounter, tran *data.Transactor, evt *data.Eventer) error {
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
		"consensus.CreateFormulation": &txFee{CreateFormulationTransctionType, amount.COIN.DivC(10)},
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

func initGenesisContextData(act *data.Accounter, tran *data.Transactor, evt *data.Eventer) (*data.ContextData, error) {
	loader := data.NewEmptyLoader(act.ChainCoord(), act, tran, evt)
	ctd := data.NewContextData(loader, nil)

	addFormulator(loader, ctd, common.MustParsePublicHash("2NDLwtFxtrtUzy6Dga8mpzJDS5kapdWBKyptMhehNVB"), common.MustParseAddress("3CUsUpvEK"), "sandbox.fr00001")
	addSingleAccount(loader, ctd, common.MustParsePublicHash(address.ADDR.MainAccount.Hash), address.ADDR.MainAccount.Addr, "testaccount")
	return ctd, nil
}

func addFormulator(loader data.Loader, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address, name string) {
	a, err := loader.Accounter().NewByTypeName("consensus.FormulationAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*consensus.FormulationAccount)
	acc.Address_ = addr
	acc.Name_ = name
	acc.Balance_ = amount.NewCoinAmount(0, 0)
	acc.KeyHash = KeyHash
	ctd.CreatedAccountMap[acc.Address_] = acc
}

func addSingleAccount(loader data.Loader, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address, name string) {
	a, err := loader.Accounter().NewByTypeName("fleta.SingleAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*account_def.SingleAccount)
	acc.Address_ = addr
	acc.Name_ = name
	acc.Balance_ = amount.NewCoinAmount(10000000000, 0)
	acc.KeyHash = KeyHash
	ctd.CreatedAccountMap[acc.Address_] = acc
}

type accCoordGenerator struct {
	idx uint16
}

func (acg *accCoordGenerator) Generate() *common.Coordinate {
	coord := common.NewCoordinate(0, acg.idx)
	acg.idx++
	return coord
}
