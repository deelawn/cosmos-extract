package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
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
