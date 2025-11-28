package login

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	azureLoginURLFormat = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

// SecretsAPI defines the subset of the AWS Secrets Manager client used by the service.
type SecretsAPI interface {
	GetSecretValue(
		ctx context.Context,
		params *secretsmanager.GetSecretValueInput,
		optFns ...func(*secretsmanager.Options),
	) (*secretsmanager.GetSecretValueOutput, error)
}

// Service fetches API auth credentials from AWS Secrets Manager and exchanges
// them for an Azure AD access token.
type Service struct {
	secretName string
	secrets    SecretsAPI
	httpClient *http.Client
}

// Credentials mirrors the JSON document stored in Secrets Manager.
type Credentials struct {
	ClientSecret string `json:"ApiAuthClientSecret"`
	Role         string `json:"ApiAuthRole"`
	ClientID     string `json:"ApiAuthClientId"`
	DsClientID   string `json:"ApiAuthDsClientId"`
	TenantID     string `json:"ApiAuthTenantId"`
	Scope        string `json:"ApiAuthScope"`
}

// Result represents the token returned from Azure AD.
type Result struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
}

// tokenResponse is shaped like the Azure AD v2 client credentials response.
type tokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ExtExpires  int    `json:"ext_expires_in"`
	AccessToken string `json:"access_token"`
}

// NewService builds a Service with the provided dependencies.
func NewService(secretName string, secrets SecretsAPI, httpClient *http.Client) *Service {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return &Service{
		secretName: secretName,
		secrets:    secrets,
		httpClient: httpClient,
	}
}

// Login retrieves the credentials from Secrets Manager and requests a new token.
func (s *Service) Login(ctx context.Context) (*Result, error) {
	if s.secrets == nil {
		return nil, errors.New("secrets client cannot be nil")
	}
	if s.secretName == "" {
		return nil, errors.New("secret name is required")
	}

	creds, err := s.fetchCredentials(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.requestToken(ctx, creds)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second).UTC()
	result := &Result{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		ExpiresAt:   expiresAt,
	}

	return result, nil
}

func (s *Service) fetchCredentials(ctx context.Context) (*Credentials, error) {
	out, err := s.secrets.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.secretName),
	})
	if err != nil {
		return nil, fmt.Errorf("get secret %q: %w", s.secretName, err)
	}

	if out.SecretString == nil || strings.TrimSpace(*out.SecretString) == "" {
		return nil, errors.New("secret does not contain SecretString payload")
	}

	var creds Credentials
	if err := json.Unmarshal([]byte(*out.SecretString), &creds); err != nil {
		return nil, fmt.Errorf("parse secret JSON: %w", err)
	}
	if err := creds.Validate(); err != nil {
		return nil, err
	}

	return &creds, nil
}

func (s *Service) requestToken(ctx context.Context, creds *Credentials) (*tokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", creds.ClientID)
	form.Set("client_secret", creds.ClientSecret)
	form.Set("scope", creds.Scope)

	endpoint := fmt.Sprintf(azureLoginURLFormat, creds.TenantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var token tokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	if token.AccessToken == "" {
		return nil, errors.New("empty access token in response")
	}

	if token.ExpiresIn <= 0 {
		token.ExpiresIn = 3600
	}
	if token.TokenType == "" {
		token.TokenType = "Bearer"
	}

	return &token, nil
}

// Validate ensures the credential payload contains the expected fields.
func (c Credentials) Validate() error {
	var missing []string
	if c.ClientID == "" {
		missing = append(missing, "ApiAuthClientId")
	}
	if c.ClientSecret == "" {
		missing = append(missing, "ApiAuthClientSecret")
	}
	if c.TenantID == "" {
		missing = append(missing, "ApiAuthTenantId")
	}
	if c.Scope == "" {
		missing = append(missing, "ApiAuthScope")
	}
	if len(missing) > 0 {
		return fmt.Errorf("secret missing fields: %s", strings.Join(missing, ", "))
	}
	return nil
}
