package neptune

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
	"sync"
	"time"

	"github.com/ankit-lilly/nqcli/internal/config"
	"github.com/ankit-lilly/nqcli/internal/core"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type Backend string

const (
	BackendREST          Backend = "rest"
	BackendHTTP          Backend = "http"
	backendAuto          Backend = "auto"
	defaultGremlinAccept         = "application/json"
)

type Client struct {
	httpClient *http.Client
	cfg        *config.Config
	awsCfg     aws.Config
	signer     *v4.Signer
	region     string
	backendMu  sync.RWMutex
	backend    Backend
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
		backend:    backendAuto,
	}, nil
}

type restQueryRequest struct {
	Type  string `json:"type"`
	Query string `json:"query"`
}

func (c *Client) ExecuteQuery(ctx context.Context, query, queryType string, opts core.QueryOpts) (core.QueryPayload, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if queryType != "gremlin" || c.cfg.DirectURL == "" {
		return c.executeREST(ctx, query, queryType, opts)
	}

	switch c.selectedBackend() {
	case BackendREST:
		return c.executeREST(ctx, query, queryType, opts)
	case BackendHTTP:
		return c.executeDirectGremlin(ctx, query, opts)
	default:
		payload, err := c.executeDirectGremlin(ctx, query, opts)
		if err == nil {
			c.selectBackend(BackendHTTP)
			return payload, nil
		}

		payload, restErr := c.executeREST(ctx, query, queryType, opts)
		if restErr == nil {
			c.selectBackend(BackendREST)
		}
		return payload, restErr
	}
}

func (c *Client) selectedBackend() Backend {
	c.backendMu.RLock()
	defer c.backendMu.RUnlock()
	return c.backend
}

func (c *Client) selectBackend(backend Backend) {
	c.backendMu.Lock()
	c.backend = backend
	c.backendMu.Unlock()
}

func (c *Client) executeREST(ctx context.Context, query, queryType string, opts core.QueryOpts) (core.QueryPayload, error) {
	payload := restQueryRequest{Type: queryType, Query: query}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if opts.Serializer != "" {
		req.Header.Set("Accept", opts.Serializer)
	}

	creds, err := c.awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to load AWS credentials: %w", err)
	}

	payloadHash := sha256.Sum256(jsonPayload)
	if err := c.signer.SignHTTP(
		ctx,
		creds,
		req,
		hex.EncodeToString(payloadHash[:]),
		"execute-api",
		c.region,
		time.Now(),
	); err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := readResponse(resp.Body, opts.MaxResponseBytes)
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to read response: %w", err)
	}
	raw := string(body)
	content := normalizeRESTPayload(raw)

	if resp.StatusCode != http.StatusOK {
		return core.QueryPayload{Content: content, Raw: raw}, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, raw)
	}

	return core.QueryPayload{Content: content, Raw: raw}, nil
}

func (c *Client) executeDirectGremlin(ctx context.Context, query string, opts core.QueryOpts) (core.QueryPayload, error) {
	jsonPayload, err := json.Marshal(map[string]string{"gremlin": query})
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to marshal direct gremlin payload: %w", err)
	}

	endpoint := strings.TrimRight(c.cfg.DirectURL, "/") + "/gremlin"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to create direct gremlin request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if opts.Serializer != "" {
		req.Header.Set("Accept", opts.Serializer)
	} else {
		req.Header.Set("Accept", defaultGremlinAccept)
	}

	creds, err := c.awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to load AWS credentials: %w", err)
	}

	payloadHash := sha256.Sum256(jsonPayload)
	if err := c.signer.SignHTTP(
		ctx,
		creds,
		req,
		hex.EncodeToString(payloadHash[:]),
		"neptune-db",
		c.region,
		time.Now(),
	); err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to sign direct gremlin request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to execute direct gremlin request: %w", err)
	}
	defer resp.Body.Close()

	body, err := readResponse(resp.Body, opts.MaxResponseBytes)
	if err != nil {
		return core.QueryPayload{}, fmt.Errorf("failed to read direct gremlin response: %w", err)
	}

	raw := string(body)
	content := normalizeDirectGremlinPayload(raw, opts)
	if resp.StatusCode != http.StatusOK {
		return core.QueryPayload{Content: content, Raw: raw}, fmt.Errorf("direct Neptune returned status code %d: %s", resp.StatusCode, raw)
	}

	return core.QueryPayload{Content: content, Raw: raw}, nil
}

func normalizeRESTPayload(raw string) string {
	if unwrapped, ok := unwrapSingleDataEnvelope(raw); ok {
		return unwrapped
	}
	return raw
}

func normalizeDirectGremlinPayload(raw string, opts core.QueryOpts) string {
	if data, ok := unwrapNeptuneResultData(raw); ok {
		if opts.Serializer == "" {
			if plain, ok := unwrapGraphSON(data); ok {
				return plain
			}
		}
		return data
	}
	return normalizeRESTPayload(raw)
}

func unwrapNeptuneResultData(raw string) (string, bool) {
	var parsed struct {
		Result struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", false
	}
	if len(parsed.Result.Data) == 0 || string(parsed.Result.Data) == "null" {
		return "", false
	}
	return string(parsed.Result.Data), true
}

func unwrapSingleDataEnvelope(raw string) (string, bool) {
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", false
	}

	obj, ok := parsed.(map[string]any)
	if !ok || len(obj) != 1 {
		return "", false
	}

	data, ok := obj["data"]
	if !ok || data == nil {
		return "", false
	}

	content, err := json.Marshal(data)
	if err != nil {
		return "", false
	}

	return string(content), true
}

func regionFromURL(endpoint string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint %q: %w", endpoint, err)
	}
	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("endpoint %q missing hostname", endpoint)
	}

	parts := strings.Split(host, ".")
	for i, part := range parts {
		if (part == "appsync-api" || part == "execute-api") && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("unable to infer AWS region from endpoint %q", endpoint)
}

func readResponse(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(r)
	}
	b, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes; narrow the query or export a smaller page", maxBytes)
	}
	return b, nil
}
