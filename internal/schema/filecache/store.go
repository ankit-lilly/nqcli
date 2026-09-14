package filecache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	graphschema "github.com/ankit-lilly/nqcli/internal/schema"
)

const formatVersion = 1

type Store struct {
	dir string
}

type record struct {
	Version  int                  `json:"version"`
	Key      string               `json:"key"`
	Snapshot graphschema.Snapshot `json:"snapshot"`
}

func Default() *Store {
	dir, err := os.UserCacheDir()
	if err != nil {
		return &Store{}
	}
	return New(filepath.Join(dir, "nq", "schema"))
}

func New(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) Load(key string) (graphschema.Snapshot, error) {
	if s.dir == "" {
		return graphschema.Snapshot{}, graphschema.ErrCacheMiss
	}
	payload, err := os.ReadFile(s.path(key))
	if errors.Is(err, os.ErrNotExist) {
		return graphschema.Snapshot{}, graphschema.ErrCacheMiss
	}
	if err != nil {
		return graphschema.Snapshot{}, fmt.Errorf("read schema cache: %w", err)
	}
	var cached record
	if err := json.Unmarshal(payload, &cached); err != nil {
		return graphschema.Snapshot{}, fmt.Errorf("decode schema cache: %w", err)
	}
	if cached.Version != formatVersion || cached.Key != key || cached.Snapshot.Status != graphschema.StatusReady {
		return graphschema.Snapshot{}, graphschema.ErrCacheMiss
	}
	return cached.Snapshot, nil
}

func (s *Store) Save(key string, snapshot graphschema.Snapshot) error {
	if s.dir == "" || snapshot.Status != graphschema.StatusReady {
		return nil
	}
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("create schema cache directory: %w", err)
	}
	payload, err := json.Marshal(record{Version: formatVersion, Key: key, Snapshot: snapshot})
	if err != nil {
		return fmt.Errorf("encode schema cache: %w", err)
	}
	temporary, err := os.CreateTemp(s.dir, ".schema-*.tmp")
	if err != nil {
		return fmt.Errorf("create schema cache file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("secure schema cache file: %w", err)
	}
	if _, err := temporary.Write(payload); err != nil {
		temporary.Close()
		return fmt.Errorf("write schema cache: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close schema cache: %w", err)
	}
	destination := s.path(key)
	if err := os.Rename(temporaryName, destination); err == nil {
		return nil
	}
	if err := os.Remove(destination); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("replace schema cache: %w", err)
	}
	if err := os.Rename(temporaryName, destination); err != nil {
		return fmt.Errorf("commit schema cache: %w", err)
	}
	return nil
}

func (s *Store) path(key string) string {
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(s.dir, hex.EncodeToString(hash[:])+".json")
}
