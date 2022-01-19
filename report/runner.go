package report

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/cosmos-extract/client"
	"github.com/figment-networks/indexing-engine/structs"
	"go.uber.org/zap"
)

const (
	network string = "cosmos"
)

type Runner interface {
	Run(ctx context.Context, accounts []string, outputPath string) error
}

func NewRunner(logger *zap.Logger, cosmosClient *client.Client) Runner {
	return &runner{logger: logger, client: cosmosClient}
}

type runner struct {
	logger *zap.Logger
	client *client.Client
}

func (r *runner) Run(ctx context.Context, accounts []string, outputPath string) error {

	startTime := time.Now()
	r.logger.Info("Starting report run...")
	defer r.logger.Info("REPORT RUN COMPLETE in " + time.Since(startTime).String())

	ha := structs.HeightAccount{
		Network: network,
		Account: accounts[0],
		Height:  8795000,
	}
	resp, err := r.client.GetAccountDelegations(ctx, ha)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	rb, err := r.client.GetRewardBalances(ctx, []structs.HeightAccount{ha})
	if err != nil {
		return err
	}

	fmt.Println(rb)

	return nil
}
