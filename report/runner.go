package report

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/figment-networks/cosmos-extract/client"
	"github.com/figment-networks/indexing-engine/structs"
	"go.uber.org/zap"
)

const (
	network string = "cosmos"
	chainID string = "cosmoshub-4"
)

type Runner interface {
	Run(ctx context.Context, config *Config) error
}

func NewRunner(logger *zap.Logger, cosmosClient client.Client) Runner {
	return &runner{logger: logger, client: cosmosClient}
}

type Config struct {
	StartTime time.Time
	EndTime   time.Time
	// This will always be by month unless we need it otherwise.
	// GroupBy time.Duration
	Accounts   []string
	OutputPath string
}

type runner struct {
	logger *zap.Logger
	client client.Client
}

func (r *runner) Run(ctx context.Context, cfg *Config) error {

	if cfg == nil {
		return errors.New("no config provided")
	}

	startTime := time.Now()
	r.logger.Info("Starting report run...")

	// Periods are built -- it is a chronologically ordered slice of month start and end
	// times and the the last height for each month. This data allows us to efficiently
	// query account delegation balances and rewards.
	periods, err := r.buildOrderedPeriods(ctx, cfg.StartTime, cfg.EndTime)
	if err != nil {
		return err
	}

	accounts := cfg.Accounts
	results := initAccountResults(accounts)

	// For each period, get data for each account.
	for _, period := range periods {
		for _, acc := range accounts {

			durationResult := initDurationResult(period.startTime)

			// Step 1: Get the delegation balances by validator.
			heightAccount := structs.HeightAccount{
				Height:  period.endHeight,
				Account: acc,
				Network: network,
				ChainID: chainID,
			}

			r.logger.Info("Getting account delegations", zap.String("account", acc), zap.Time("period", period.startTime))
			delegationsResp, err := r.client.GetAccountDelegations(ctx, heightAccount)
			if err != nil {
				return fmt.Errorf("could not get account delegations for %+v: %w", heightAccount, err)
			}

			for _, d := range delegationsResp.Delegations {
				durationResult.validators[string(d.Validator)] = true
				durationResult.delegations[string(d.Validator)] = d.Balance.Numeric
			}

			// Step 2: Get rewards earned by validator.
			rewReq := client.RewardsReq{
				Network:   network,
				ChainID:   chainID,
				Account:   acc,
				StartTime: period.startTime,
				EndTime:   period.nextStartTime,
			}

			r.logger.Info("Getting account rewards", zap.String("account", acc), zap.Time("period", period.startTime))
			rewSum, feeSum, err := r.client.GetRewardsAndFeesSum(ctx, rewReq)
			if err != nil {
				return fmt.Errorf("could not get rewards for %+v: %w", rewReq, err)
			}

			// Not possible to have fees without rewards, so just check rewards.
			durationResult.rewards = rewSum
			for v := range rewSum {
				durationResult.validators[v] = true
			}
			durationResult.fees = feeSum

			results[acc] = append(results[acc], durationResult)
		}
	}

	r.logger.Info("REPORT RUN COMPLETE in " + time.Since(startTime).String())

	return results.writeToDisk(cfg.Accounts, cfg.OutputPath)
}
