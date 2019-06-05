package address

import (
	"encoding/hex"

	"github.com/fletaio/common"
	"github.com/fletaio/core/key"
)

type accCoordGenerator1 struct {
	idx uint16
}

func (acg *accCoordGenerator1) Generate() *common.Coordinate {
	coord := common.NewCoordinate(0, acg.idx)
	acg.idx++
	return coord
}

type Key struct {
	Addr   common.Address
	Hash   string
	SK     string
	Signer *key.MemoryKey
}

func NewKey(sk string) *Key {
	data, err := hex.DecodeString(sk)
	if err != nil {
		panic(err)
	}

	k := &Key{
		SK: sk,
	}

	k.Signer, err = key.NewMemoryKeyFromBytes(data)
	if err != nil {
		panic(err)
	}
	k.Hash = common.NewPublicHash(k.Signer.PublicKey()).String()
	return k
}

//Addresses TODO
type Addresses struct {
	MainObserver   []*Key
	MainFormulator []*Key
	MainAccount    *Key

	MainToken *Key

	DAppObserver   []*Key
	DAppFormulator []*Key
	DAppAccount    *Key
}

var ADDR = Addresses{
	MainObserver: []*Key{
		NewKey("cca49818f6c49cf57b6c420cdcd98fcae08850f56d2ff5b8d287fddc7f9ede08"),
		NewKey("39f1a02bed5eff3f6247bb25564cdaef20d410d77ef7fc2c0181b1d5b31ce877"),
		NewKey("2b97bc8f21215b7ed085cbbaa2ea020ded95463deef6cbf31bb1eadf826d4694"),
		NewKey("3b43d728deaa62d7c8790636bdabbe7148a6641e291fd1f94b157673c0172425"),
		NewKey("e6cf2724019000a3f703db92829ecbd646501c0fd6a5e97ad6774d4ad621f949"),
	},
	MainFormulator: []*Key{
		NewKey("ef475a14258d0a6f061293628e299a78e6abd7d46f0eb544c473045c84dffa31"),
	},
	MainAccount: NewKey("5e0dc680d12a728f60a708dcdbfb8d2c2aaea3ee5748d12bd9358f1015e3d18b"),
	MainToken:   NewKey("116034cda48d0704426ae2141a6ea8d9a4193862cebfc9a25aab53063916caac"),
	DAppObserver: []*Key{
		NewKey("246fd0687a0cc717a5af8ae2067fcdd05e9227f27b37e3a5387081037b340b7a"),
		NewKey("5e0dc680d12a728f60a708dcdbfb8d2c2aaea3ee5748d12bd9358f1015e3d18b"),
		NewKey("cd28be0aed28023126409d79f0bcc8b07e7045f25b76f19b80333841e4b50638"),
		NewKey("b884a8999a6eb905d1f51a0f8b33d480429dae2f48daa437c6b427156a09f4ac"),
		NewKey("a18fcd5d9232d8e3e45e9c4d9664bdfff9fe76b0cb01912afb25a3fc707dbed5"),
	},
	DAppFormulator: []*Key{
		NewKey("0592f595790fa9af27d560ffc957b2eb3e4b79cd67ceb710ebd111de5057e379"),
	},
	DAppAccount: NewKey("0ee9aff86e07b44d9adb40246c2ecbf511000d96f22b5a0232fff27e3df8c88d"),
}

func init() {
	{
		acg := &accCoordGenerator1{idx: 1}
		for i := 0; i < len(ADDR.MainFormulator); i++ {
			addr := common.NewAddress(acg.Generate(), 0)
			ADDR.MainFormulator[i].Addr = addr
		}
		ADDR.MainAccount.Addr = common.NewAddress(acg.Generate(), 0)
	}
}

var DappCoor *common.Coordinate

func DappInitAddr(dappCoor *common.Coordinate) {
	DappCoor = dappCoor
	{
		acg := &accCoordGenerator1{idx: 1}
		for i := 0; i < len(ADDR.DAppFormulator); i++ {
			addr := common.NewAddress(acg.Generate(), 0)
			ADDR.DAppFormulator[i].Addr = addr
		}
		ADDR.DAppAccount.Addr = common.NewAddress(acg.Generate(), 0)
	}

}
