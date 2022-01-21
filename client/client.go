package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/figment-networks/cosmos-worker/api"
	"github.com/figment-networks/indexing-engine/structs"

	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	GRPCAddr               string
	SearchAddr             string
	AuthToken              string
	GRPCMaxRecvSize        int
	GRPCMaxSendSize        int
	RequestsPerSecond      int
	TimeoutBlockCall       time.Duration
	TimeoutTransactionCall time.Duration
}

type Client interface {
	Close()
	GetAccountDelegations(
		ctx context.Context,
		params structs.HeightAccount,
	) (resp structs.GetAccountDelegationsResponse, err error)
	GetLastHeightBefore(ctx context.Context, req LastHeightBeforeReq) (height uint64, err error)
}

type client struct {
	apiClient     *api.Client
	grpcConn      *grpc.ClientConn
	stakingClient stakingTypes.QueryClient
	authToken     string
	searchAddr    string
	searchClient  http.Client
}

func New(ctx context.Context, logger *zap.Logger, cfg Config) (c Client, err error) {

	tlsc, err := loadTLSCredentials()
	if err != nil {
		return
	}

	dialOptions := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cfg.GRPCMaxRecvSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(cfg.GRPCMaxSendSize)),
		grpc.WithTransportCredentials(tlsc),
		grpc.WithPerRPCCredentials(tokenAuth{
			token: cfg.AuthToken,
		}),
	}

	grpcConn, err := grpc.DialContext(ctx, cfg.GRPCAddr, dialOptions...)
	if err != nil {
		err = fmt.Errorf("error dialing grpc: %w", err)
		return
	}

	api.InitMetrics()

	clientConfig := api.ClientConfig{
		ReqPerSecond:        cfg.RequestsPerSecond,
		TimeoutBlockCall:    cfg.TimeoutBlockCall,
		TimeoutSearchTxCall: cfg.TimeoutTransactionCall,
	}

	return &client{
		apiClient:     api.NewClient(logger, grpcConn, &clientConfig),
		stakingClient: stakingTypes.NewQueryClient(grpcConn),
		grpcConn:      grpcConn,
		authToken:     cfg.AuthToken,
		searchAddr:    cfg.SearchAddr,
		searchClient:  http.Client{},
	}, nil
}

func (c *client) Close() {
	c.grpcConn.Close()
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	ootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		RootCAs: ootCAs,
	}), nil
}
