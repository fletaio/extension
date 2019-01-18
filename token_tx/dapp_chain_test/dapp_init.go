package dapp_chain

import (
	"log"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/consensus"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/store"
	"git.fleta.io/fleta/core/transaction"
	"git.fleta.io/fleta/extension/account_def"

	// use init func
	_ "git.fleta.io/fleta/extension/account_tx"
	_ "git.fleta.io/fleta/extension/utxo_tx"
	_ "git.fleta.io/fleta/javascript"
	_ "git.fleta.io/fleta/solidity"
)

func initDAppChainComponent(act *data.Accounter, tran *data.Transactor) error {
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
		// Formulation Transactions
		CreateFormulationTransctionType = transaction.Type(60)
		RevokeFormulationTransctionType = transaction.Type(61)
		// Solidity Transactions
		SolidityCreateContractType = transaction.Type(70)
		SolidityCallContractType   = transaction.Type(71)
		// Javascript Transactions
		JavascriptCreateAccountType  = transaction.Type(80)
		JavascriptCreateContractType = transaction.Type(81)
		JavascriptCallContractType   = transaction.Type(82)
	)

	// account_type account types
	const (
		// FLTEA Accounts
		SingleAccountType   = account.Type(10)
		MultiSigAccountType = account.Type(11)
		LockedAccountType   = account.Type(19)
		// Formulation Accounts
		FormulationAccountType = account.Type(60)
		// Solidity Accounts
		SolidityAccount = account.Type(70)
		// Javascript Accounts
		JavascriptAccount = account.Type(80)
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
		"consensus.CreateFormulation": &txFee{CreateFormulationTransctionType, amount.COIN.MulC(50000)},
		"consensus.RevokeFormulation": &txFee{RevokeFormulationTransctionType, amount.COIN.DivC(10)},
		"solidity.CreateContract":     &txFee{SolidityCreateContractType, amount.COIN.MulC(10)},
		"solidity.CallContract":       &txFee{SolidityCallContractType, amount.COIN.DivC(10)},
		"javascript.CreateAccount":    &txFee{JavascriptCreateAccountType, amount.COIN.MulC(0)},
		"javascript.CreateContract":   &txFee{JavascriptCreateContractType, amount.COIN.MulC(0)},
		"javascript.CallContract":     &txFee{JavascriptCallContractType, amount.COIN.DivC(10)},
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
		"fleta.LockedAccount":          LockedAccountType,
		"consensus.FormulationAccount": FormulationAccountType,
		"solidity.ContractAccount":     SolidityAccount,
		"javascript.ContractAccount":   JavascriptAccount,
	}
	for name, t := range AccTable {
		if err := act.RegisterType(name, t); err != nil {
			log.Println(name, t, err)
			return err
		}
	}
	return nil
}

func initDAppGenesisContextData(st *store.Store, ctd *data.ContextData, PublicHashs []string) error {
	acg := &accDAppCoordGenerator{idx: 1}
	for _, PublicHash := range PublicHashs {
		a := acg.Generate()
		s := st.ChainCoord()
		addr := common.NewAddress(a, s, 0)
		log.Println("dapp addr : ", addr.String(), a.Height, " : ", a.Index, " : ", s.Height, " : ", s.Index)
		addDAppFormulator(st.Accounter(), ctd, common.MustParsePublicHash(PublicHash), addr)
	}
	a := acg.Generate()
	s := st.ChainCoord()
	addr := common.NewAddress(a, s, 0)
	log.Println("dapp account : ", addr.String(), a.Height, " : ", a.Index, " : ", s.Height, " : ", s.Index)
	addDAppSingleAccount(st.Accounter(), s, ctd, common.MustParsePublicHash("2ndv1i9toF393VeqLbsFtuWD2xtXfCETZGmHdXDVnwo"), addr)
	// addFormulator(st, ctd, common.MustParsePublicHash("2VdGunZe8yZNm2mErqQqrFx2B7Mb4SBRPWviWnapahw"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
	return nil
}

func addDAppSingleAccount(act *data.Accounter, coord *common.Coordinate, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address) {
	a, err := act.NewByTypeName("fleta.SingleAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*account_def.SingleAccount)
	acc.Address_ = addr
	acc.KeyHash = KeyHash
	ctd.CreatedAccountHash[acc.Address_] = acc
	balance := account.NewBalance()
	balance.AddBalance(coord, amount.NewCoinAmount(10000000000, 0))
	ctd.AccountBalanceHash[acc.Address_] = balance
}

func addDAppFormulator(act *data.Accounter, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address) {
	a, err := act.NewByTypeName("consensus.FormulationAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*consensus.FormulationAccount)
	acc.Address_ = addr
	acc.KeyHash = KeyHash
	ctd.CreatedAccountHash[acc.Address_] = acc
}

type accDAppCoordGenerator struct {
	idx uint16
}

func (acg *accDAppCoordGenerator) Generate() *common.Coordinate {
	coord := common.NewCoordinate(0, acg.idx)
	acg.idx++
	return coord
}
