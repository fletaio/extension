package dapp_chain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/block"
	"git.fleta.io/fleta/core/chain"
	"git.fleta.io/fleta/core/consensus"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/generator"
	"git.fleta.io/fleta/core/kernel"
	"git.fleta.io/fleta/core/key"
	"git.fleta.io/fleta/core/message_def"
	"git.fleta.io/fleta/core/observer"
	"git.fleta.io/fleta/core/store"
	"git.fleta.io/fleta/core/transaction"
	"git.fleta.io/fleta/framework/message"
	"git.fleta.io/fleta/framework/peer"
	"git.fleta.io/fleta/framework/router"
	"git.fleta.io/fleta/framework/router/evilnode"
	"git.fleta.io/fleta/network"
)

//GetSigner is get memoryKey from securekey
func GetSigner(sk string) *key.MemoryKey {
	javascriptHash := sk
	data, err := hex.DecodeString(javascriptHash)
	if err != nil {
		panic(err)
	}
	JavascriptSigner, err := key.NewMemoryKeyFromBytes(data)
	if err != nil {
		panic(err)
	}

	return JavascriptSigner
}

type testTx struct {
	Tx   transaction.Transaction
	Sigs []common.Signature
}

// observer errors
var (
	ErrRequiredObserverAddresses = errors.New("required observer addresses")
	ErrRequiredRequestTimeout    = errors.New("required request timeout")
	ErrNotConnected              = errors.New("not connected")
	ErrRequestTimeout            = errors.New("request timeout")
	ErrAlreadyRequested          = errors.New("already requested")
	ErrInvalidResponse           = errors.New("invalid response")
	ErrInvalidTopFormulator      = errors.New("invalid top formulator")
)

// Observer TODO
type Observer struct {
	Config     *observer.Config
	kn         *kernel.Kernel
	mm         *message.Manager
	lastSender net.Conn
	Keys       []key.Key
	conns      []net.Conn
}

// NewObserver TODO
func NewObserver(Config *observer.Config, kn *kernel.Kernel, strs []string) (*Observer, error) {
	/*
		if len(Config.Addresses) == 0 {
			return nil, ErrRequiredObserverAddresses
		}
	*/

	keys := make([]key.Key, 0, len(strs))
	for _, v := range strs {
		if bs, err := hex.DecodeString(v); err != nil {
			return nil, err
		} else if key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			return nil, err
		} else {
			keys = append(keys, key)
		}
	}
	mm := message.NewManager()
	ob := &Observer{
		Config: Config,
		kn:     kn,
		mm:     mm,
		Keys:   keys,
		conns:  make([]net.Conn, 0),
	}
	mm.ApplyMessage(message_def.BlockGenMessageType, ob.BlockGenMessageCreator, ob.RecvMessageHandler)
	mm.ApplyMessage(message_def.BlockObSignMessageType, ob.BlockObSignMessageCreator, ob.RecvMessageHandler)
	return ob, nil
}

func (ob *Observer) log(msgs ...interface{}) {
	msgs = append([]interface{}{"[Ob]", ob.Config.ID}, msgs...)
	log.Println(msgs)
}

// Start TODO
func (ob *Observer) Start() {
	go func() {
		ob.log("Listen " + ob.Config.ID + ":3001")
		lst, _ := network.Listen("mock:"+ob.Config.ID, ":3001")
		for {
			ob.log("Accept befor " + ob.Config.ID + ":3001")
			conn, _ := lst.Accept()
			ob.log("Accept after " + ob.Config.ID + ":3001")
			// TODO : handshake - know formulator address - use signature when handshake
			ob.conns = append(ob.conns, conn)
			go func() {
				for {
					if v, _, err := util.ReadUint64(conn); err != nil {
						panic(err)
					} else {
						ob.lastSender = conn
						msg, handler, err := ob.mm.ParseMessage(conn, message.Type(v))
						if err != nil {
							panic(err)
						}
						if err := handler(msg); err != nil {
							panic(err)
						}
						ob.log("ObsRecv", ob.Config.Network, msg)
					}
					// TODO : RankTable과 동일한 알고리즘으로 Conn을 정렬해주는 알고리즘 사용
					// TODO : send by rank - max 10th rank
					// TODO : 10등이 참여해야 할 때 연결이 없으면 미리 Timeout 동의 받아두기(다른 옵저버가 연결이 있으면 연결해제 합의에 해당 옵저버 주소 기재)
					// TODO : 10등 내에서 연결이 끊기면 무조건 Timeout 동의 받아두기
				}
			}()
		}
	}()
}

//BlockGenMessageCreator TODO
func (ob *Observer) BlockGenMessageCreator(r io.Reader) message.Message {
	p := message_def.NewBlockGenMessage(ob.kn.Loader().Transactor())
	p.ReadFrom(r)
	return p
}

//BlockObSignMessageCreator TODO
func (ob *Observer) BlockObSignMessageCreator(r io.Reader) message.Message {
	p := message_def.NewBlockObSignMessage()
	p.ReadFrom(r)
	return p
}

var gSpandsLock sync.Mutex
var gSpands = []time.Duration{}

// RecvMessageHandler TODO
func (ob *Observer) RecvMessageHandler(m message.Message) error {
	switch msg := m.(type) {
	case *message_def.BlockGenMessage:
		loader := ob.kn.Loader()
		provider := ob.kn.Provider()
		if is, err := ob.kn.IsMinable(msg.Block.Header.FormulationAddress, msg.Block.Header.TimeoutCount); err != nil {
			return err
		} else if !is {
			return ErrInvalidTopFormulator
		}
		if pubkey, err := common.RecoverPubkey(msg.Block.Header.Hash(), msg.Signed.GeneratorSignature); err != nil {
			return err
		} else {
			if acc, err := loader.Account(msg.Block.Header.FormulationAddress); err != nil {
				return err
			} else {
				if err := loader.Accounter().Validate(loader, acc, []common.PublicHash{common.NewPublicHash(pubkey)}); err != nil {
					return err
				}
			}
		}

		BlockHash := msg.Block.Header.Hash()
		sigs := []common.Signature{}
		ls := rand.Perm(len(ob.Keys))
		for i := 0; i < len(ob.Keys)/2+1; i++ {
			if sig, err := ob.Keys[ls[i]].Sign(BlockHash); err != nil {
				return nil
			} else {
				sigs = append(sigs, sig)
			}
		}
		// TODO : sign by consensus

		nos := &block.ObserverSigned{
			Signed: block.Signed{
				BlockHash:          BlockHash,
				GeneratorSignature: msg.Signed.GeneratorSignature,
			},
			ObserverSignatures: sigs,
		}

		if msg.Block.Header.Height == 1 {
			ctx, err := ob.kn.ContextByBlock(msg.Block)
			if err != nil {
				return err
			}
			if err := ob.kn.AppendBlock(msg.Block, nos, ctx); err != nil {
				return err
			}
		}

		var blockTime uint64
		b, err := provider.Block(1)
		if err == nil {
			blockTime = b.Header.Timestamp
		}
		halfSecond := 500 * time.Millisecond
		spand := time.Duration(msg.Block.Header.Height)*halfSecond - time.Duration(uint64(time.Now().UnixNano())-blockTime)

		gSpandsLock.Lock()
		gSpands = append(gSpands, spand)
		gSpandsLock.Unlock()

		ob.log("spand", spand)
		if spand < 0 {
			spand = 0
		}
		if spand > halfSecond {
			spand = halfSecond
		}
		time.Sleep(spand)

		obMsg := &message_def.BlockObSignMessage{
			ObserverSigned: nos,
		}
		var buffer bytes.Buffer
		if _, err := util.WriteUint64(&buffer, uint64(obMsg.Type())); err != nil {
			return err
		}
		if _, err := obMsg.WriteTo(&buffer); err != nil {
			return err
		}
		if _, err := ob.lastSender.Write(buffer.Bytes()); err != nil {
			return err
		}
		ob.log("BlockGenMessage", ob.Config.Network, obMsg.ObserverSigned.BlockHash, msg.Block.Header.Height, len(msg.Block.Transactions))
		{
			bMsg := &message_def.BlockMessage{
				Block: msg.Block,
				ObserverSigned: &block.ObserverSigned{
					Signed: block.Signed{
						BlockHash:          BlockHash,
						GeneratorSignature: msg.Signed.GeneratorSignature,
					},
					ObserverSignatures: sigs,
				},
				Tran: loader.Transactor(),
			}
			var buffer bytes.Buffer
			if _, err := util.WriteUint64(&buffer, uint64(bMsg.Type())); err != nil {
				return err
			}
			if _, err := bMsg.WriteTo(&buffer); err != nil {
				return err
			}
			for _, c := range ob.conns {
				if ob.lastSender != c {
					if _, err := c.Write(buffer.Bytes()); err != nil {
						return err
					}
					ob.log("BlockMessage", ob.Config.Network, bMsg.ObserverSigned.BlockHash, bMsg.Block.Header.Height)
				}
			}
		}
		//log.Println("Consume : ", time.Duration(profiler.Snapshot()[0].Data.TSum))
		return nil
	case *message_def.BlockObSignMessage:
		ob.log("BlockObSignMessage", ob.Config.Network, msg.ObserverSigned.BlockHash)
	}
	return nil
}

//GetDAppConfig is return the config struct from parameter
func GetDAppConfig(ID string, SeedNodes []string, Generator string, ChainCoord *common.Coordinate, obID string, ObserverSignatures []string) *kernel.Config {
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
				StorePath:    "./_data/" + ID + "/dapp/router",
			},
		},
		Peer: peer.Config{
			StorePath: "./_data/" + ID + "/dapp/peers",
		},
		Generator: generator.Config{
			Address:           Generator,
			BlockVersion:      1,
			GenTimeThreshold:  200,
			ObserverAddresses: []string{},
		},
		Observer: observer.Config{
			ID:        obID,
			Addresses: []string{obID + ":3001"},
			Network:   "mock:connector_" + ID,
		},
		StorePath: "./_data/" + ID + "/dapp",
	}

	return Config
}

func CreateDAppKernel(ID string, r router.Router, Config *kernel.Config, PublicHashs []string) *kernel.Kernel {
	act := data.NewAccounter(Config.ChainCoord)
	tran := data.NewTransactor(Config.ChainCoord)
	if err := initDAppChainComponent(act, tran); err != nil {
		panic(err)
	}

	st, err := store.NewStore(Config.StorePath, act, tran)
	if err != nil {
		panic(err)
	}

	GenesisContextData := data.NewContextData(data.NewEmptyLoader(st.ChainCoord(), st.Accounter(), st.Transactor()), nil)

	if err := initDAppGenesisContextData(st, GenesisContextData, PublicHashs); err != nil {
		panic(err)
	}
	rewarder := &Rewarder{}

	kn, err := kernel.NewKernel(Config, r, st, rewarder, GenesisContextData)
	if err != nil {
		panic(err)
	}

	return kn
}

func SetupFormulator(kn *kernel.Kernel, KeyHex string) *kernel.Kernel {
	RequestTimeout := 5 //seconds
	data, err := hex.DecodeString(KeyHex)
	if err != nil {
		panic(err)
	}
	Signer, err := key.NewMemoryKeyFromBytes(data)
	if err != nil {
		panic(err)
	}
	gn, err := generator.NewGenerator(&kn.Config.Generator, Signer)
	if err != nil {
		panic(err)
	}

	var ob *observer.Connector
	if kn.Config.Observer.ID != "" {
		ob, err = observer.NewConnector(&kn.Config.Observer, RequestTimeout, kn.Loader().Transactor())
		if err != nil {
			panic(err)
		}
	}
	if err := kn.InitFormulator(gn, ob); err != nil {
		panic(err)
	}
	return kn
}

type Rewarder struct {
}

func (rd *Rewarder) ProcessReward(FormulationAddress common.Address, ctx *data.Context) error {
	balance, err := ctx.AccountBalance(FormulationAddress)
	if err != nil {
		return err
	}
	balance.AddBalance(ctx.ChainCoord(), amount.NewCoinAmount(1, 0))
	return nil
}

func AddFormulator(st *store.Store, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address) {
	a, err := st.Accounter().NewByTypeName("consensus.FormulationAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*consensus.FormulationAccount)
	acc.Address_ = addr
	acc.KeyHash = KeyHash
	ctd.CreatedAccountHash[acc.Address_] = acc
}
