package gq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ankit-lilly/nqcli/internal/config"
)

type Client struct {
	httpClient *http.Client
	cfg        *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cfg:        cfg,
	}
}

type GraphQLPayload struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

type NeptuneQueryVariables struct {
	Input struct {
		Type  string `json:"type"`
		Query string `json:"query"`
	} `json:"input"`
}

func (c *Client) ExecuteQuery(query string, queryType string) (string, error) {
	variables := NeptuneQueryVariables{}
	variables.Input.Type = queryType
	variables.Input.Query = query

	payload := GraphQLPayload{
		Query:     `mutation ($input: NeptuneQuery!) { executeQuery(input: $input) }`,
		Variables: variables,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.cfg.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf.String(), nil
}
