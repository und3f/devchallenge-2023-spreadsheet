package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type CellResponse struct {
	Value  string `json:"value"`
	Result string `json:"result"`
}

func RestGetCell(url string) (string, error) {
	client := http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Failed to obtain cell value for external source")
	}

	defer resp.Body.Close()
	var cellResp CellResponse
	if err := json.NewDecoder(resp.Body).Decode(&cellResp); err != nil {
		return "", err
	}

	return cellResp.Result, nil
}
