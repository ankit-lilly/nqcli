package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	URL string
}

var (
	defaultEnvOnce sync.Once
)

// LoadEnvironment attempts to load environment variables from a .env file.
// If envFile is provided it must exist; otherwise the function searches the
// current working directory and the user's home directory for ".env".
func LoadEnvironment(envFile string) error {
	candidates, err := buildEnvCandidates(envFile)
	if err != nil {
		return err
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			if envFile != "" {
				return fmt.Errorf("failed to stat env file %q: %w", candidate, err)
			}
			continue
		}

		if info.IsDir() {
			if envFile != "" {
				return fmt.Errorf("env file %q is a directory", candidate)
			}
			continue
		}

		if err := godotenv.Load(candidate); err != nil {
			if envFile != "" {
				return fmt.Errorf("failed to load env file %q: %w", candidate, err)
			}
			continue
		}
		return nil
	}

	if envFile != "" {
		return fmt.Errorf("env file %q not found", envFile)
	}

	return nil
}

func buildEnvCandidates(envFile string) ([]string, error) {
	if envFile != "" {
		expanded, err := expandPath(envFile)
		if err != nil {
			return nil, err
		}
		return []string{expanded}, nil
	}

	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, ".env"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".env"))
	}
	return candidates, nil
}

func expandPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("empty path")
	}

	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return filepath.Abs(p)
}

func ensureDefaultEnvLoaded() {
	defaultEnvOnce.Do(func() {
		_ = LoadEnvironment("")
	})
}

func LoadConfig() *Config {
	ensureDefaultEnvLoaded()

	cfg := &Config{
		URL: os.Getenv("NEPTUNE_URL"),
	}

	if cfg.URL == "" {
		cfg.URL = "https://rqhlqaisn5c6to736svujsowwm.appsync-api.us-east-2.amazonaws.com/graphql"
	}

	return cfg
}
