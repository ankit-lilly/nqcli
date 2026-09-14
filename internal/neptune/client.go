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
	content := normalizeRESTPayload(body)

	if resp.StatusCode != http.StatusOK {
		raw := string(body)
		return core.QueryPayload{Content: content, Raw: raw}, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, responseSnippet(body))
	}

	return core.QueryPayload{Content: content, Raw: preservedRaw(body, opts)}, nil
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

	content := normalizeDirectGremlinPayload(body, opts)
	if resp.StatusCode != http.StatusOK {
		raw := string(body)
		return core.QueryPayload{Content: content, Raw: raw}, fmt.Errorf("direct Neptune returned status code %d: %s", resp.StatusCode, responseSnippet(body))
	}

	return core.QueryPayload{Content: content, Raw: preservedRaw(body, opts)}, nil
}

func preservedRaw(body []byte, opts core.QueryOpts) string {
	if !opts.PreserveRaw {
		return ""
	}
	return string(body)
}

func responseSnippet(body []byte) string {
	const limit = 2 << 10
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "... (truncated)"
}

func normalizeRESTPayload(raw []byte) string {
	if unwrapped, ok := unwrapSingleDataEnvelope(raw); ok {
		return unwrapped
	}
	return string(raw)
}

func normalizeDirectGremlinPayload(raw []byte, opts core.QueryOpts) string {
	if data, ok := unwrapNeptuneResultData(raw); ok {
		if opts.Serializer == "" {
			if plain, ok := unwrapGraphSON(data); ok {
				return plain
			}
		}
		return string(data)
	}
	return normalizeRESTPayload(raw)
}

func unwrapNeptuneResultData(raw []byte) ([]byte, bool) {
	var parsed struct {
		Result struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, false
	}
	if len(parsed.Result.Data) == 0 || bytes.Equal(parsed.Result.Data, []byte("null")) {
		return nil, false
	}
	return parsed.Result.Data, true
}

func unwrapSingleDataEnvelope(raw []byte) (string, bool) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope) != 1 {
		return "", false
	}
	data, ok := envelope["data"]
	if !ok || len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return "", false
	}
	return string(data), true
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
