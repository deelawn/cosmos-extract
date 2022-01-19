package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"

	"github.com/figment-networks/cosmos-extract/client"
	"github.com/figment-networks/cosmos-extract/report"
	"github.com/figment-networks/cosmos-worker/api"
	"github.com/figment-networks/cosmos-worker/cmd/common/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var configPath string

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.StringVar(&configPath, "config", "", "Path to config")
	flag.Parse()

	cfg, err := initConfig(configPath)
	if err != nil {
		log.Fatalf("error initializing config [ERR: %v]", err.Error())
	}

	logger.Init("console", "debug", []string{"stderr"})
	defer logger.Sync()

	if cfg.CosmosGRPCAddr == "" {
		logger.Error(fmt.Errorf("cosmos grpc address is not set"))
		return
	}

	if cfg.AuthToken == "" {
		logger.Error(fmt.Errorf("cosmos grpc token is not set"))
		return
	}

	tlsc, err := loadTLSCredentials()
	if err != nil {
		logger.Error(err)
		return
	}

	var dialOptions []grpc.DialOption
	dialOptions = append(dialOptions,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cfg.GrpcMaxRecvSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(cfg.GrpcMaxSendSize)),
		grpc.WithTransportCredentials(tlsc),
		grpc.WithPerRPCCredentials(tokenAuth{
			token: cfg.AuthToken,
		}),
		// grpc.WithInsecure(),
	)

	// switch cfg.TLSMode {
	// case "server":
	// 	tlsc, err := loadTLSCredentials()
	// 	if err != nil {
	// 		logger.Error(err)
	// 		return
	// 	}
	// 	dialOptions = append(dialOptions, grpc.WithTransportCredentials(tlsc))
	// default:
	// 	dialOptions = append(dialOptions, grpc.WithInsecure())
	// }

	grpcConn, dialErr := grpc.DialContext(ctx, cfg.CosmosGRPCAddr, dialOptions...)
	if dialErr != nil {
		logger.Error(fmt.Errorf("error dialing grpc: %w", dialErr))
		return
	}
	defer grpcConn.Close()

	cosmosClient := client.New(logger.GetLogger(), grpcConn, &api.ClientConfig{
		ReqPerSecond:        int(cfg.RequestsPerSecond),
		TimeoutBlockCall:    cfg.TimeoutBlockCall,
		TimeoutSearchTxCall: cfg.TimeoutTransactionCall,
	})

	reportRunner := report.NewRunner(logger.GetLogger(), cosmosClient)
	err = reportRunner.Run(ctx, cfg.Accounts, cfg.ReportOutput)
	if err != nil {
		logger.Error(err)
		return
	}
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
