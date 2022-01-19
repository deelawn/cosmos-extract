package main

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	AuthToken              string        `json:"auth_token" envconfig:"AUTH_TOKEN"`
	CosmosGRPCAddr         string        `json:"cosmos_grpc_addr" envconfig:"COSMOS_GRPC_ADDR"`
	GrpcMaxRecvSize        int           `json:"grpc_max_recv_size" envconfig:"GRPC_MAX_RECV_SIZE" default:"1073741824"` // 1024^3
	GrpcMaxSendSize        int           `json:"grpc_max_send_size" envconfig:"GRPC_MAX_SEND_SIZE" default:"1073741824"` // 1024^3
	TLSMode                string        `json:"tls_mode" envconfig:"TLS_MODE" default:""`
	RequestsPerSecond      int64         `json:"requests_per_second" envconfig:"REQUESTS_PER_SECOND" default:"33"`
	TimeoutBlockCall       time.Duration `json:"timeout_block_call" envconfig:"TIMEOUT_BLOCK_CALL" default:"30s"`
	TimeoutTransactionCall time.Duration `json:"timeout_transaction_call" envconfig:"TIMEOUT_TRANSACTION_CALL" default:"30s"`
	Accounts               []string      `json:"accounts" envconfig:"ACCOUNTS"`
	ReportOutput           string        `json:"report_output" envconfig:"REPORT_OUTPUT" default:"out.csv"`
}

func initConfig(path string) (*config, error) {
	cfg := &config{}
	if path != "" {
		if err := fromFile(path, cfg); err != nil {
			return nil, err
		}
	}

	if cfg.CosmosGRPCAddr != "" {
		return cfg, nil
	}

	if err := fromEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func fromFile(path string, config *config) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

func fromEnv(config *config) error {
	return envconfig.Process("", config)
}
