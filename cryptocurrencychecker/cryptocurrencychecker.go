package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type apiResponse struct {
	Status string `json:"status"`
	Stats  map[string]struct {
		Latest string `json:"latest"`
	} `json:"stats"`
}

func GetExchangeRate(source, destination string) (string, error) {

	src := strings.ToLower(strings.TrimSpace(source))
	dst := strings.ToLower(strings.TrimSpace(destination))

	if src == "" {
		return "", fmt.Errorf("source currency cannot be empty")
	}
	if dst == "" {
		dst = "rls"
	}

	url := fmt.Sprintf("http://localhost:4001/rates?srcCurrency=%s&dstCurrency=%s", src, dst)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status: %s", resp.Status)
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if strings.ToUpper(data.Status) != "OK" {
		return "", fmt.Errorf("api status: %s", data.Status)
	}

	key := fmt.Sprintf("%s-%s", src, dst)
	stat, ok := data.Stats[key]
	if !ok || stat.Latest == "" {
		return "", fmt.Errorf("rate not found for %s", key)
	}

	return stat.Latest, nil
}
