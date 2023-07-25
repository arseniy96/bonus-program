package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sethgrid/pester"

	"github.com/arseniy96/bonus-program/internal/logger"
)

type Client struct {
	HTTPClient *pester.Client
}

func InitClient() *Client {
	client := pester.New()
	client.MaxRetries = 3
	client.KeepLog = true

	return &Client{
		HTTPClient: client,
	}
}

func (c *Client) GetOrderRequest(url string) (*GetOrderResponse, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		logger.Log.Errorf("accrual request error: %v", c.HTTPClient.LogString())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status in accrual response: %v", resp.Status)
	}

	r := GetOrderResponse{
		Accrual: 0, // default value
	}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
