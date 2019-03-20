package dappChainTest

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/fletaio/common/hash"

	"github.com/fletaio/common"
	"github.com/fletaio/core/amount"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/extension/account_tx"
	"github.com/fletaio/extension/token_tx"
	"github.com/fletaio/framework/router"
	"github.com/fletaio/javascript"
)

//DApp TODO
type DApp struct {
	DAppRouter router.Router
	IsRun      bool
	runLock    sync.Mutex
	MainKn     *kernel.Kernel
	DAppKn     *kernel.Kernel
}

//InitDAppChain is initialization dapp chain
func (d *DApp) InitDAppChain(Address common.Address, ObserverInfos []token_tx.ObserverInfo) {
	d.runLock.Lock()
	if d.IsRun == false {
		d.IsRun = true
	} else {
		panic("double run")
	}
	d.runLock.Unlock()

	ObserverPhashs := make([]string, 0, len(ObserverInfos))
	for _, info := range ObserverInfos {
		ObserverPhashs = append(ObserverPhashs, info.Hash)
	}

	bf := bytes.NewBuffer(Address[:6])
	DAppChainCoord := &common.Coordinate{}
	DAppChainCoord.ReadFrom(bf)

	SeedNodes := []string{"formulator_1:3000"}

	ObConfig := GetDAppConfig("dapp_observer_1", SeedNodes, "", DAppChainCoord, "dapp_observer_1", ObserverPhashs)
	r, err := router.NewRouter(&ObConfig.Router, kn.ChainCoord())
	if err != nil {
		panic(err)
	}
	obKernel := CreateDAppKernel("dapp_observer_1", r, ObConfig)
	obsk := make([]string, len(ADDR.DAppObserver))
	for i, k := range ADDR.DAppObserver {
		obsk[i] = k.SK
	}
	ob, err := NewObserver(&ObConfig.Observer, obKernel, obsk)
	if err != nil {
		panic(err)
	}
	go ob.Start()

	FmConfig := GetDAppConfig("formulator_1", SeedNodes, ADDR.DAppFormulator[0].Addr.String(), DAppChainCoord, "dapp_observer_1", ObserverPhashs)
	if d.DAppRouter == nil {
		log.Println("***** DAppRouter is nil *****")
		var err error
		d.DAppRouter, err = router.NewRouter(&FmConfig.Router, kn.ChainCoord())
		if err != nil {
			panic(err)
		}
	}

	d.DAppKn = CreateDAppKernel("formulator_1", d.DAppRouter, FmConfig)
	SetupFormulator(d.DAppKn, ADDR.DAppFormulator[0].SK)

}

//GenesisHash is return genesis block's hash
func (d *DApp) GenesisHash() hash.Hash256 {
	hash, err := d.DAppKn.Provider().BlockHash(0)
	if err != nil {
		panic(err)
	}
	return hash
}

//RunDAppChain is run dapp chain
func (d *DApp) RunDAppChain(Address common.Address, GenHash hash.Hash256) {
	bf := bytes.NewBuffer(Address[:6])
	DAppChainCoord := &common.Coordinate{}
	DAppChainCoord.ReadFrom(bf)

	hash := d.GenesisHash()
	if hash != GenHash {
		panic("different genesise Hash")
	}

	d.DAppKn.Start()
	d.DAppKn.PeerManager.EnforceConnect()

	time.Sleep(time.Second)

	go d.DAppKn.TryGenerateBlock()

	{
		// start CreateContract
		cc, err := d.DAppKn.Loader().Transactor().NewByTypeName("fleta.Transfer")
		if err != nil {
			panic(err)
		}
		t := cc.(*account_tx.Transfer)

		printBalance(d.DAppKn, DAppChainCoord, ADDR.DAppFormulator[0].Addr, ADDR.DAppAccount.Addr)

		t.Seq_ = d.DAppKn.Loader().Seq(ADDR.DAppAccount.Addr) + 1
		// jsigner := javascriptSigner()
		t.From_ = ADDR.DAppAccount.Addr
		t.To = ADDR.DAppFormulator[0].Addr
		t.Amount = amount.NewCoinAmount(10, 0)

		sig1, _ := ADDR.DAppAccount.Signer.Sign(t.Hash())
		sigs1 := []common.Signature{sig1}

		d.DAppKn.AddTransaction(t, sigs1)
		// end CreateContract

		time.Sleep(time.Second * 1)
	}

	printBalance(d.DAppKn, DAppChainCoord, ADDR.DAppFormulator[0].Addr, ADDR.DAppAccount.Addr)

	{
		// start CreateContract
		cc, err := d.DAppKn.Loader().Transactor().NewByTypeName("javascript.CreateContract")
		if err != nil {
			panic(err)
		}

		t := cc.(*javascript.CreateContract)

		t.Seq_ = d.DAppKn.Loader().Seq(ADDR.DAppFormulator[0].Addr) + 1
		t.From_ = ADDR.DAppFormulator[0].Addr
		t.Name = "testTRANSFERFunction"
		t.Code = []byte(`
function testTRANSFERFunction () {
	var t = __TRANSFER__(15.1111)
	if (typeof msg.callCount === "undefined") {
		msg.callCount = 0
	}
	msg.callCount++;msg.testValue = t;
}
`)

		sig1, _ := ADDR.DAppFormulator[0].Signer.Sign(t.Hash())
		sigs1 := []common.Signature{sig1}

		d.DAppKn.AddTransaction(t, sigs1)
		// end CreateContract

		time.Sleep(time.Second * 1)
	}

	log.Println("kn1Addr : ", ADDR.DAppFormulator[0].Addr)

	go d.HashEngraveOnMain()

	{
		//start CallContract
		cm, err := d.DAppKn.Loader().Transactor().NewByTypeName("javascript.CallContract")
		if err != nil {
			panic(err)
		}

		t := cm.(*javascript.CallContract)

		t.From_ = ADDR.DAppAccount.Addr
		t.To = ADDR.DAppFormulator[0].Addr
		t.Method = []byte("testTRANSFERFunction")
		t.Params = [][]byte{}

		for {
			time.Sleep(time.Second * 1)
			printBalance(d.DAppKn, DAppChainCoord, ADDR.DAppFormulator[0].Addr, ADDR.DAppAccount.Addr)

			t.Seq_ = d.DAppKn.Loader().Seq(ADDR.DAppAccount.Addr) + 1
			log.Println("k1 seq : ", t.Seq_)
			sig2, _ := ADDR.DAppAccount.Signer.Sign(t.Hash())
			sigs2 := []common.Signature{sig2}

			d.DAppKn.AddTransaction(t, sigs2)
		}
		//end CallContract
	}

}

//HashEngraveOnMain is dapp block hash engrave on main chain
func (d *DApp) HashEngraveOnMain() {
	cm, err := d.MainKn.Loader().Transactor().NewByTypeName("fleta.EngraveDapp")
	if err != nil {
		panic(err)
	}

	t := cm.(*token_tx.EngraveDapp)

	tokenAddr := common.MustParseAddress("5QxWJopN3")

	t.From_ = tokenAddr
	t.TokenAddress = tokenAddr

	for {
		//start CallContract

		t.Height = d.DAppKn.Provider().Height()
		t.BlockHash, err = d.DAppKn.Provider().BlockHash(t.Height)

		printBalance(d.MainKn, MainCoor, tokenAddr, tokenAddr)

		t.Seq_ = d.MainKn.Loader().Seq(tokenAddr) + 1
		log.Println("tokenAddr seq : ", t.Seq_)
		sig2, _ := ADDR.MainToken.Signer.Sign(t.Hash())
		sigs2 := []common.Signature{sig2}

		d.MainKn.AddTransaction(t, sigs2)
		//end CallContract
		time.Sleep(time.Second * 3)
	}
}

func printBalance(kn *kernel.Kernel, coord *common.Coordinate, addr1 common.Address, addr2 common.Address) {
	b1, _ := kn.Loader().AccountBalance(addr1)
	b2, _ := kn.Loader().AccountBalance(addr2)
	log.Println("**********", b1.Balance(coord).DivC(amount.FractionalMax).Int, b2.Balance(coord).DivC(amount.FractionalMax).Int, "**********")

}
