package client

import (
	"context"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type validatorCommission struct {
	value       *big.Int
	lastChanged time.Time
}

func (c *client) getValidatorCommission(ctx context.Context, validator string) (vc validatorCommission, err error) {

	resp, err := c.stakingClient.Validator(ctx, &types.QueryValidatorRequest{ValidatorAddr: validator})
	if err != nil {
		return
	}

	vc.value = resp.Validator.Commission.Rate.BigInt()
	vc.lastChanged = resp.Validator.Commission.UpdateTime

	return
}
