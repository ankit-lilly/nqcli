package gq

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ankit-lilly/nqcli/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// Client is an HTTP client for executing signed GraphQL requests against an
// AWS AppSync endpoint backed by Amazon Neptune.
type Client struct {
	httpClient *http.Client
	cfg        *config.Config
	awsCfg     aws.Config
	signer     *v4.Signer
	region     string
}

// NewClient creates a Client configured with the given Neptune config and AWS
// credentials. It infers the AWS region from the endpoint URL if not set in awsCfg.
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

// GraphQLPayload is the JSON body sent to the AppSync endpoint.
type GraphQLPayload struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

// NeptuneQueryVariables holds the input variables for a Neptune query mutation.
type NeptuneQueryVariables struct {
	Input struct {
		Type  string `json:"type"`
		Query string `json:"query"`
	} `json:"input"`
}

// ExecuteQuery runs a Gremlin or Cypher query against Neptune using a default
// background context. Prefer ExecuteQueryCtx when a caller context is available.
func (c *Client) ExecuteQuery(query string, queryType string) (string, error) {
	return c.ExecuteQueryCtx(context.Background(), query, queryType)
}

// ExecuteQueryCtx runs a Neptune query with the given context for cancellation and timeouts.
func (c *Client) ExecuteQueryCtx(ctx context.Context, query string, queryType string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	variables := NeptuneQueryVariables{}
	variables.Input.Type = queryType
	variables.Input.Query = query

	return c.ExecuteGraphQLCtx(ctx, `mutation ($input: NeptuneQuery!) { executeQuery(input: $input) }`, variables)
}

// ExecuteGraphQL sends a signed GraphQL request using a default background context.
// Prefer ExecuteGraphQLCtx when a caller context is available.
func (c *Client) ExecuteGraphQL(query string, variables any) (string, error) {
	return c.ExecuteGraphQLCtx(context.Background(), query, variables)
}

// ExecuteGraphQLCtx sends a signed GraphQL request using the provided context.
func (c *Client) ExecuteGraphQLCtx(ctx context.Context, query string, variables any) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	payload := GraphQLPayload{
		Query:     query,
		Variables: variables,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
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
