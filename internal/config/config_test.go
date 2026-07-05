package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvironment_ExplicitFileNotFound(t *testing.T) {
	t.Parallel()
	err := LoadEnvironment("/nonexistent/path/.env")
	if err == nil {
		t.Fatal("expected error for missing explicit env file")
	}
}

func TestLoadEnvironment_ExplicitFileIsDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	err := LoadEnvironment(dir)
	if err == nil {
		t.Fatal("expected error when env file is a directory")
	}
}

func TestLoadEnvironment_LoadsVariables(t *testing.T) {
	// Not parallel — manipulates env vars.
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("TEST_NQ_VAR=hello123\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Cleanup(func() { os.Unsetenv("TEST_NQ_VAR") })

	if err := LoadEnvironment(envPath); err != nil {
		t.Fatalf("LoadEnvironment: %v", err)
	}

	if got := os.Getenv("TEST_NQ_VAR"); got != "hello123" {
		t.Fatalf("expected TEST_NQ_VAR=hello123, got %q", got)
	}
}

func TestLoadEnvironment_NoFileNoError(t *testing.T) {
	t.Parallel()
	// When no explicit file is given and no .env exists in cwd/home, should not error.
	// We can't guarantee cwd doesn't have .env, so this is a basic sanity check.
	err := LoadEnvironment("")
	// Either nil (no file found, which is fine) or nil (file loaded successfully).
	// Should never error when envFile is "".
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfig_ReadsEnvVars(t *testing.T) {
	// Not parallel — manipulates env vars.
	t.Setenv("NEPTUNE_URL", "https://example.com/graphql")
	t.Setenv("NEPTUNE_APPSYNC_API_NAME", " my-api ")
	t.Setenv("NEPTUNE_APPSYNC_API_ID", "abc123")

	cfg := LoadConfig()

	if cfg.URL != "https://example.com/graphql" {
		t.Errorf("URL: got %q, want %q", cfg.URL, "https://example.com/graphql")
	}
	if cfg.AppSyncAPIName != "my-api" {
		t.Errorf("AppSyncAPIName: got %q, want %q (should be trimmed)", cfg.AppSyncAPIName, "my-api")
	}
	if cfg.AppSyncAPIID != "abc123" {
		t.Errorf("AppSyncAPIID: got %q, want %q", cfg.AppSyncAPIID, "abc123")
	}
}

func TestExpandPath_Tilde(t *testing.T) {
	t.Parallel()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir available")
	}

	result, err := expandPath("~/foo/bar")
	if err != nil {
		t.Fatalf("expandPath: %v", err)
	}

	expected := filepath.Join(home, "foo", "bar")
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExpandPath_Empty(t *testing.T) {
	t.Parallel()
	_, err := expandPath("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}
