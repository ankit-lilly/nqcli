package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEnvValueCreatesFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	if err := WriteEnvValue(envPath, "NEPTUNE_TOKEN", "abc"); err != nil {
		t.Fatalf("WriteEnvValue returned error: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read env file: %v", err)
	}

	got := string(data)
	want := "NEPTUNE_TOKEN=\"abc\"\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWriteEnvValueUpdatesExistingEntry(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	seed := "NEPTUNE_URL=\"url\"\nNEPTUNE_TOKEN=\"old\"\n"
	if err := os.WriteFile(envPath, []byte(seed), 0o600); err != nil {
		t.Fatalf("failed to seed env file: %v", err)
	}

	if err := WriteEnvValue(envPath, "NEPTUNE_TOKEN", "new"); err != nil {
		t.Fatalf("WriteEnvValue returned error: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read env file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if lines[len(lines)-1] != `NEPTUNE_TOKEN="new"` {
		t.Fatalf("expected updated token, got %v", lines)
	}
}

func TestResolveEnvFileForWriteDefault(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	path, err := ResolveEnvFileForWrite("")
	if err != nil {
		t.Fatalf("ResolveEnvFileForWrite returned error: %v", err)
	}

	resolvedTmp, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		resolvedTmp = tmp
	}
	expected := filepath.Join(resolvedTmp, ".env")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}
