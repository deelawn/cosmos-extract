package report

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/figment-networks/cosmos-extract/client"
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
	defer r.logger.Info("REPORT RUN COMPLETE in " + time.Since(startTime).String())

	// We want to get the last height for each of the months that fall within the provided time range,
	// regardless of the specific times. Start by calculating the number of months we need heights for.
	startYear := cfg.StartTime.Year()
	startMonth := int(cfg.StartTime.Month())
	endYear := cfg.EndTime.Year()
	endMonth := int(cfg.EndTime.Month())
	numMonths := (endYear-startYear)*12 + endMonth - startMonth + 1

	lastMonthHeights := make([]uint64, numMonths)
	for i := 0; i < numMonths; i++ {
		// Use the current month to get the first time of the next month. We will pass this "before time"
		// as an argument to indicate we want the last height that occurred before this time.
		nextMonthFirstTime := time.Date(startYear, time.Month(startMonth+i+1), 1, 0, 0, 0, 0, time.UTC)
		req := client.LastHeightBeforeReq{
			Network:    network,
			ChainID:    chainID,
			BeforeTime: nextMonthFirstTime,
		}

		height, err := r.client.GetLastHeightBefore(ctx, req)
		if err != nil {
			return err
		}

		lastMonthHeights[i] = height
	}

	fmt.Println(lastMonthHeights)

	// Get the last heights for each month

	// ha := structs.HeightAccount{
	// 	Network: network,
	// 	Account: accounts[0],
	// 	Height:  height,
	// }
	// resp, err := r.client.GetAccountDelegations(ctx, ha)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(resp)

	// rb, err := r.client.GetRewardBalances(ctx, []structs.HeightAccount{ha})
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(rb)

	return nil
}
