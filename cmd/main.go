package main

import (
	"context"
	"flag"
	"log"

	"github.com/figment-networks/cosmos-extract/client"
	"github.com/figment-networks/cosmos-extract/report"
	"github.com/figment-networks/cosmos-worker/cmd/common/logger"
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

	if err := cfg.validate(); err != nil {
		logger.Error(err)
		return
	}

	clientConfig := client.Config{
		GRPCAddr:               cfg.CosmosGRPCAddr,
		SearchAddr:             cfg.CosmosSearchAddr,
		AuthToken:              cfg.AuthToken,
		GRPCMaxRecvSize:        cfg.GrpcMaxRecvSize,
		GRPCMaxSendSize:        cfg.GrpcMaxSendSize,
		RequestsPerSecond:      cfg.RequestsPerSecond,
		TimeoutBlockCall:       cfg.TimeoutBlockCall,
		TimeoutTransactionCall: cfg.TimeoutTransactionCall,
	}
	cosmosClient, err := client.New(ctx, logger.GetLogger(), clientConfig)
	if err != nil {
		logger.Error(err)
		return
	}
	defer cosmosClient.Close()

	reportRunner := report.NewRunner(logger.GetLogger(), cosmosClient)
	reportConfig := report.Config{
		StartTime:  cfg.StartTime,
		EndTime:    cfg.EndTime,
		Accounts:   cfg.Accounts,
		OutputPath: cfg.ReportOutput,
	}
	err = reportRunner.Run(ctx, &reportConfig)
	if err != nil {
		logger.Error(err)
		return
	}
}
