package gq

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ankit-lilly/nqcli/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type Client struct {
	httpClient *http.Client
	cfg        *config.Config
	awsCfg     aws.Config
	signer     *v4.Signer
	region     string
}

func NewClient(cfg *config.Config, awsCfg aws.Config) (*Client, error) {
	region := awsCfg.Region
	if region == "" {
		parsed, err := regionFromURL(cfg.URL)
		if err != nil {
			return nil, err
		}
		region = parsed
	}

	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cfg:        cfg,
		awsCfg:     awsCfg,
		signer:     v4.NewSigner(),
		region:     region,
	}, nil
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

	req.Header.Set("Content-Type", "application/json")

	ctx := context.Background()
	creds, err := c.awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS credentials: %w", err)
	}

	payloadHash := sha256.Sum256(jsonPayload)
	if err := c.signer.SignHTTP(
		ctx,
		creds,
		req,
		hex.EncodeToString(payloadHash[:]),
		"appsync",
		c.region,
		time.Now(),
	); err != nil {
		return "", fmt.Errorf("failed to sign request: %w", err)
	}

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

func regionFromURL(endpoint string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("appsync endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid appsync endpoint %q: %w", endpoint, err)
	}
	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("appsync endpoint %q missing hostname", endpoint)
	}

	parts := strings.Split(host, ".")
	for i, part := range parts {
		if part == "appsync-api" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("unable to infer AWS region from appsync endpoint %q", endpoint)
}
