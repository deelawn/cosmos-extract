package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/figment-networks/indexing-engine/structs"
)

type LastHeightBeforeReq struct {
	Network    string
	ChainID    string
	BeforeTime time.Time
}

type txSearchReqBody struct {
	Network    string    `json:"network"`
	ChainIDs   []string  `json:"chain_ids"`
	BeforeTime time.Time `json:"before_time"`
	Limit      int       `json:"limit"`
}

type heightResp struct {
	Height uint64 `json:"height"`
}

func (c client) GetLastHeightBefore(ctx context.Context, req LastHeightBeforeReq) (height uint64, err error) {

	searchReq := txSearchReqBody{
		Network:    req.Network,
		ChainIDs:   []string{req.ChainID},
		BeforeTime: req.BeforeTime,
		Limit:      1,
	}

	rawBody, err := json.Marshal(searchReq)
	if err != nil {
		return
	}

	body := bytes.NewReader(rawBody)
	url := c.searchAddr
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "transactions_search"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return
	}

	httpReq.Header = http.Header{"Authorization": []string{c.authToken}}
	resp, err := c.searchClient.Do(httpReq)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		rawB, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("%w, %s", err, string(rawB))
	}

	var hr []heightResp
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&hr); err != nil {
		return
	}

	if len(hr) == 0 {
		err = errors.New("no heights found before time " + req.BeforeTime.String())
		return
	}

	return hr[0].Height, nil
}

type RewardsReq struct {
	Network   string    `json:"network"`
	ChainID   string    `json:"chain_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Account   string    `json:"account"`
}

func (c client) GetRewardsAndFeesSum(ctx context.Context, req RewardsReq) (rewards map[string]*big.Int, fees map[string]*big.Int, err error) {

	validatorCommisions := map[string]validatorCommission{}

	url := c.searchAddr
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "rewards?"

	params := neturl.Values{}
	params.Add("network", req.Network)
	params.Add("chain_id", req.ChainID)
	params.Add("start_time", req.StartTime.Format(time.RFC3339))
	params.Add("end_time", req.EndTime.Format(time.RFC3339))
	params.Add("account", req.Account)
	url += params.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	httpReq.Header = http.Header{"Authorization": []string{c.authToken}}
	resp, err := c.searchClient.Do(httpReq)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		rawB, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("%w, %s", err, string(rawB))
	}

	var dailySumm []structs.RewardSummary
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&dailySumm); err != nil {
		return
	}

	rewards = map[string]*big.Int{}
	fees = map[string]*big.Int{}
	for _, entry := range dailySumm {

		// First sum all amounts present in this entry.
		entrySubtotal := big.NewInt(0)
		for _, amount := range entry.Amount {
			entrySubtotal = entrySubtotal.Add(entrySubtotal, amount.Numeric)
		}

		// Then create or add to the running total for this validator.
		if rew, ok := rewards[string(entry.Validator)]; ok {
			rewards[string(entry.Validator)] = rew.Add(rew, entrySubtotal)
		} else {
			rewards[string(entry.Validator)] = entrySubtotal
		}

		comm, ok := validatorCommisions[string(entry.Validator)]
		if !ok {
			comm, err = c.getValidatorCommission(ctx, string(entry.Validator))
			if err != nil {
				fmt.Printf("error getting validator %s: %s\n", string(entry.Validator), err.Error())
			}
			validatorCommisions[string(entry.Validator)] = comm
		}

		if comm.lastChanged.After(req.StartTime) {
			fmt.Printf("validator %s fee last changed on %s\n", entry.Validator, comm.lastChanged.String())
		}

		newFee := big.NewInt(0)
		// 10^18 is the numerator needed to get the commission rate
		commNumerator, ok := big.NewInt(0).SetString("1000000000000000000", 10)
		if !ok {
			panic("no can numerator")
		}

		newFee = newFee.Div(big.NewInt(0).Mul(entrySubtotal, comm.value), commNumerator)

		if fee, ok := fees[string(entry.Validator)]; ok {
			fees[string(entry.Validator)] = fee.Add(fee, newFee)
		} else {
			fees[string(entry.Validator)] = newFee
		}
	}

	return rewards, fees, nil

}
