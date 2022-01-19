package client

import (
	"github.com/figment-networks/cosmos-worker/api"

	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Client struct {
	*api.Client
	StakingClient stakingTypes.QueryClient
}

func New(logger *zap.Logger, cli *grpc.ClientConn, cfg *api.ClientConfig) *Client {
	api.InitMetrics()
	return &Client{
		Client:        api.NewClient(logger, cli, cfg),
		StakingClient: stakingTypes.NewQueryClient(cli),
	}
}
