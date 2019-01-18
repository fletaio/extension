package dapp_chain

import (
	"bytes"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"git.fleta.io/fleta/common/hash"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/kernel"
	"git.fleta.io/fleta/core/key"
	"git.fleta.io/fleta/extension/account_tx"
	"git.fleta.io/fleta/extension/token_tx"
	"git.fleta.io/fleta/framework/router"
	"git.fleta.io/fleta/javascript"
)

type DApp struct {
	DAppRouter router.Router
	IsRun      bool
	runLock    sync.Mutex
	kn         *kernel.Kernel
	secureKey  string //:= "f0507dc42f6ce962c85bc23770f39b33d0a89033b4e4bf7f075bde9b67605972"
}

func (d *DApp) InitDAppChain(keyHash common.PublicHash, Address common.Address, ObserverInfos []token_tx.ObserverInfo) {
	d.runLock.Lock()
	if d.IsRun == false {
		d.IsRun = true
	} else {
		panic("double run")
	}
	d.runLock.Unlock()

	obSks := []string{
		"246fd0687a0cc717a5af8ae2067fcdd05e9227f27b37e3a5387081037b340b7a",
		"5e0dc680d12a728f60a708dcdbfb8d2c2aaea3ee5748d12bd9358f1015e3d18b",
		"cd28be0aed28023126409d79f0bcc8b07e7045f25b76f19b80333841e4b50638",
		"b884a8999a6eb905d1f51a0f8b33d480429dae2f48daa437c6b427156a09f4ac",
		"a18fcd5d9232d8e3e45e9c4d9664bdfff9fe76b0cb01912afb25a3fc707dbed5",
	}

	PublicHashs := []string{
		"3ZBXeKsYWd8hdhnopZkXPeSJ6dKMniX3yKEwfVmvcWB",
	}

	ObserverPhashs := make([]string, 0, len(ObserverInfos))
	for _, info := range ObserverInfos {
		ObserverPhashs = append(ObserverPhashs, info.Hash)
	}

	bf := bytes.NewBuffer(Address[:6])
	DAppChainCoord := &common.Coordinate{}
	DAppChainCoord.ReadFrom(bf)
	log.Println("DApp Chain coord ", DAppChainCoord.Height, " : ", DAppChainCoord.Index)

	SeedNodes := []string{"formulator_1:3000"}

	ObConfig := GetDAppConfig("dapp_observer_1", SeedNodes, "", DAppChainCoord, "dapp_observer_1", ObserverPhashs)
	r, err := router.NewRouter(&ObConfig.Router)
	if err != nil {
		panic(err)
	}
	obKernel := CreateDAppKernel("dapp_observer_1", r, ObConfig, PublicHashs)
	ob, err := NewObserver(&ObConfig.Observer, obKernel, obSks)
	if err != nil {
		panic(err)
	}
	go ob.Start()

	keyHash1 := "f0507dc42f6ce962c85bc23770f39b33d0a89033b4e4bf7f075bde9b67605972"
	kn1Addr := common.NewAddress(common.NewCoordinate(0, 1), DAppChainCoord, 0)
	log.Println("init kernel addr : ", kn1Addr.String())

	FmConfig := GetDAppConfig("formulator_1", SeedNodes, kn1Addr.String(), DAppChainCoord, "dapp_observer_1", ObserverPhashs)
	if d.DAppRouter == nil {
		log.Println("***** DAppRouter is nil *****")
		var err error
		d.DAppRouter, err = router.NewRouter(&FmConfig.Router)
		if err != nil {
			panic(err)
		}
	}

	d.kn = CreateDAppKernel("formulator_1", d.DAppRouter, FmConfig, PublicHashs)
	SetupFormulator(d.kn, keyHash1)

}

func (d *DApp) GenesisHash() hash.Hash256 {
	hash, err := d.kn.Provider().BlockHash(0)
	if err != nil {
		panic(err)
	}
	return hash
}

func (d *DApp) RunDAppChain(keyHash common.PublicHash, Address common.Address, GenHash hash.Hash256) {
	bf := bytes.NewBuffer(Address[:6])
	DAppChainCoord := &common.Coordinate{}
	DAppChainCoord.ReadFrom(bf)

	hash, err := d.kn.Provider().BlockHash(0)
	if err != nil {
		panic(err)
	}
	log.Println("GenHash, hash", GenHash, hash)
	if hash != GenHash {
		panic("different genesise Hash")
	}

	d.kn.Start()
	d.kn.PeerManager.EnforceConnect()

	time.Sleep(time.Second)

	go d.kn.TryGenerateBlock()

	var Signer1 *key.MemoryKey
	{
		data1, err := hex.DecodeString(d.secureKey)
		if err != nil {
			panic(err)
		}
		Signer1, err = key.NewMemoryKeyFromBytes(data1)
		if err != nil {
			panic(err)
		}
	}

	dappAccountSk := "0ee9aff86e07b44d9adb40246c2ecbf511000d96f22b5a0232fff27e3df8c88d"
	knAddr := common.NewAddress(common.NewCoordinate(0, 1), DAppChainCoord, 0)
	dappAddr := common.NewAddress(common.NewCoordinate(0, 2), DAppChainCoord, 0)
	var DappSigner *key.MemoryKey
	{
		data1, err := hex.DecodeString(dappAccountSk)
		if err != nil {
			panic(err)
		}
		DappSigner, err = key.NewMemoryKeyFromBytes(data1)
		if err != nil {
			panic(err)
		}
	}

	{
		// start CreateContract
		cc, err := d.kn.Loader().Transactor().NewByTypeName("fleta.Transfer")
		if err != nil {
			panic(err)
		}
		t := cc.(*account_tx.Transfer)

		printBalance(d.kn, DAppChainCoord, knAddr, dappAddr)

		t.Seq_ = d.kn.Loader().Seq(dappAddr) + 1
		// jsigner := javascriptSigner()
		t.From_ = dappAddr
		t.To = knAddr
		t.Amount = amount.NewCoinAmount(10, 0)

		sig1, _ := DappSigner.Sign(t.Hash())
		sigs1 := []common.Signature{sig1}

		d.kn.AddTransaction(t, sigs1)
		// end CreateContract

		time.Sleep(time.Second * 1)
	}

	printBalance(d.kn, DAppChainCoord, knAddr, dappAddr)

	{
		// start CreateContract
		cc, err := d.kn.Loader().Transactor().NewByTypeName("javascript.CreateContract")
		if err != nil {
			panic(err)
		}

		t := cc.(*javascript.CreateContract)

		t.Seq_ = d.kn.Loader().Seq(knAddr) + 1
		t.From_ = knAddr
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

		sig1, _ := Signer1.Sign(t.Hash())
		sigs1 := []common.Signature{sig1}

		d.kn.AddTransaction(t, sigs1)
		// end CreateContract

		time.Sleep(time.Second * 1)
	}

	log.Println("kn1Addr : ", knAddr)

	{
		//start CallContract
		cm, err := d.kn.Loader().Transactor().NewByTypeName("javascript.CallContract")
		if err != nil {
			panic(err)
		}

		t := cm.(*javascript.CallContract)

		t.From_ = dappAddr
		t.To = knAddr
		t.Method = []byte("testTRANSFERFunction")
		t.Params = [][]byte{}

		for {
			time.Sleep(time.Second * 1)
			printBalance(d.kn, DAppChainCoord, knAddr, dappAddr)

			t.Seq_ = d.kn.Loader().Seq(dappAddr) + 1
			log.Println("k1 seq : ", t.Seq_)
			sig2, _ := DappSigner.Sign(t.Hash())
			sigs2 := []common.Signature{sig2}

			d.kn.AddTransaction(t, sigs2)
		}
		//end CallContract
	}

}

func printBalance(kn *kernel.Kernel, coord *common.Coordinate, addr1 common.Address, addr2 common.Address) {
	b1, _ := kn.Loader().AccountBalance(addr1)
	b2, _ := kn.Loader().AccountBalance(addr2)
	log.Println("**********", b1.Balance(coord).DivC(amount.FractionalMax).Int, b2.Balance(coord).DivC(amount.FractionalMax).Int, "**********")

}
