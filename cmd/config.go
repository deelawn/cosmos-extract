package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	AuthToken              string        `json:"auth_token" envconfig:"AUTH_TOKEN"`
	CosmosGRPCAddr         string        `json:"cosmos_grpc_addr" envconfig:"COSMOS_GRPC_ADDR"`
	CosmosSearchAddr       string        `json:"cosmos_search_addr" envconfig:"COSMOS_SEARCH_ADDR"`
	GrpcMaxRecvSize        int           `json:"grpc_max_recv_size" envconfig:"GRPC_MAX_RECV_SIZE" default:"1073741824"` // 1024^3
	GrpcMaxSendSize        int           `json:"grpc_max_send_size" envconfig:"GRPC_MAX_SEND_SIZE" default:"1073741824"` // 1024^3
	TLSMode                string        `json:"tls_mode" envconfig:"TLS_MODE" default:""`
	RequestsPerSecond      int           `json:"requests_per_second" envconfig:"REQUESTS_PER_SECOND" default:"33"`
	TimeoutBlockCall       time.Duration `json:"timeout_block_call" envconfig:"TIMEOUT_BLOCK_CALL" default:"30s"`
	TimeoutTransactionCall time.Duration `json:"timeout_transaction_call" envconfig:"TIMEOUT_TRANSACTION_CALL" default:"30s"`
	StartTime              time.Time     `json:"start_time" envconfig:"START_TIME"`
	EndTime                time.Time     `json:"end_time" envconfig:"END_TIME"`
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

func (c config) validate() error {

	if c.CosmosGRPCAddr == "" {
		return errors.New("cosmos grpc address is not set")
	}

	if c.AuthToken == "" {
		return errors.New("cosmos grpc token is not set")
	}

	if c.CosmosSearchAddr == "" {
		return errors.New("cosmos search address is not set")
	}

	if len(c.Accounts) == 0 {
		return errors.New("at least one account must be provided")
	}

	if c.StartTime.IsZero() {
		return errors.New("start time is not set")
	}

	if c.EndTime.IsZero() {
		return errors.New("end time is not set")
	}

	if c.StartTime.After(c.EndTime) {
		return errors.New("start time must come before end time")
	}

	return nil
}
