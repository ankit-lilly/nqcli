package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveEnvFileForWrite determines which .env file should be updated. When the
// user passes --env-file we expand it, otherwise we default to ".env" in the
// current working directory.
func ResolveEnvFileForWrite(envFile string) (string, error) {
	if envFile != "" {
		return expandPath(envFile)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("determine current working directory: %w", err)
	}
	return filepath.Join(cwd, ".env"), nil
}

// WriteEnvValue updates or appends the provided key in the env file using
// double-quoted values to avoid parsing issues.
func WriteEnvValue(path, key, value string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("env key cannot be empty")
	}
	if path == "" {
		return fmt.Errorf("env file path cannot be empty")
	}

	var (
		lines   []string
		updated bool
	)

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := scanner.Text()
			if currentKey, _, found := strings.Cut(line, "="); found && strings.TrimSpace(currentKey) == key {
				line = fmt.Sprintf(`%s=%q`, key, value)
				updated = true
			}
			lines = append(lines, line)
		}
		if scanErr := scanner.Err(); scanErr != nil {
			return fmt.Errorf("read env file: %w", scanErr)
		}
	case os.IsNotExist(err):
		// We'll create the file below.
	default:
		return fmt.Errorf("read env file %q: %w", path, err)
	}

	if !updated {
		lines = append(lines, fmt.Sprintf(`%s=%q`, key, value))
	}

	content := strings.Join(lines, "\n")
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if content == "" {
		content = fmt.Sprintf(`%s=%q`+"\n", key, value)
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create env dir %q: %w", dir, err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write env file %q: %w", path, err)
	}
	return nil
}
