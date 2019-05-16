package dappchain

import (
	"github.com/fletaio/common"
	"github.com/fletaio/core/data"
)

type mockRewarder struct {
}

// ProcessReward gives a reward to the block generator address
func (rd *mockRewarder) ProcessReward(addr common.Address, ctx *data.Context) error {
	return nil
}
