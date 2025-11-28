package login

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type stubSecretsClient struct {
	value string
	err   error
}

func (s *stubSecretsClient) GetSecretValue(
	ctx context.Context,
	params *secretsmanager.GetSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(s.value),
	}, nil
}

func TestServiceLoginSuccess(t *testing.T) {
	secret := `{
		"ApiAuthClientSecret": "super-secret",
		"ApiAuthClientId": "client-id",
		"ApiAuthTenantId": "tenant-id",
		"ApiAuthScope": "api://scope/.default"
	}`

	stub := &stubSecretsClient{value: secret}

	var capturedForm url.Values
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		body, _ := io.ReadAll(req.Body)
		req.Body.Close()
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		capturedForm = form

		resp := `{"access_token":"token","token_type":"Bearer","expires_in":7200}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(resp)),
			Header:     make(http.Header),
		}, nil
	})

	svc := NewService("secret", stub, &http.Client{Transport: transport})
	result, err := svc.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() returned error: %v", err)
	}

	if result.AccessToken != "token" {
		t.Fatalf("expected access token 'token', got %q", result.AccessToken)
	}
	if result.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %q", result.TokenType)
	}
	if time.Until(result.ExpiresAt) < time.Hour || time.Until(result.ExpiresAt) > 3*time.Hour {
		t.Fatalf("expected expiration roughly in 2h, got %v", result.ExpiresAt)
	}

	if capturedForm.Get("client_id") != "client-id" {
		t.Fatalf("expected client_id form field, got %q", capturedForm.Get("client_id"))
	}
	if capturedForm.Get("scope") != "api://scope/.default" {
		t.Fatalf("expected scope form field, got %q", capturedForm.Get("scope"))
	}
}

func TestCredentialsValidate(t *testing.T) {
	c := Credentials{}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for missing fields, got nil")
	}

	c = Credentials{
		ClientSecret: "secret",
		ClientID:     "id",
		TenantID:     "tenant",
		Scope:        "scope",
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
