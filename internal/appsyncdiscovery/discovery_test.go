package appsyncdiscovery

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheKeyFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		profile string
		region  string
		want    string
	}{
		{"myprofile", "us-east-1", "myprofile|us-east-1"},
		{"", "us-east-2", "default|us-east-2"},
		{"prod", "", "prod|unknown"},
		{"", "", "default|unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := cacheKeyFor(tt.profile, tt.region)
			if got != tt.want {
				t.Errorf("cacheKeyFor(%q, %q) = %q, want %q", tt.profile, tt.region, got, tt.want)
			}
		})
	}
}

func TestCacheRoundTrip(t *testing.T) {
	// Uses temp dir to avoid polluting real cache.
	dir := t.TempDir()
	cacheFilePath := filepath.Join(dir, "nested", "appsync_cache.json")

	cache := &cacheFile{
		Version: cacheVersion,
		Entries: map[string]*cacheEntry{
			"dev|us-east-2": {
				URL:       "https://example.com/graphql",
				APIName:   "test-api",
				APIID:     "abc123",
				Region:    "us-east-2",
				Profile:   "dev",
				FetchedAt: time.Now(),
			},
		},
	}

	if err := writeCacheToPath(cacheFilePath, cache); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	readCache, err := readCacheFromPath(cacheFilePath)
	if err != nil {
		t.Fatalf("read cache: %v", err)
	}

	if readCache.Version != cacheVersion {
		t.Errorf("version: got %d, want %d", readCache.Version, cacheVersion)
	}
	entry, ok := readCache.Entries["dev|us-east-2"]
	if !ok {
		t.Fatal("expected cache entry for dev|us-east-2")
	}
	if entry.URL != "https://example.com/graphql" {
		t.Errorf("URL: got %q, want %q", entry.URL, "https://example.com/graphql")
	}
	if entry.APIName != "test-api" {
		t.Errorf("APIName: got %q, want %q", entry.APIName, "test-api")
	}
}

func TestWriteCache_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	cacheFilePath := filepath.Join(dir, "nqcli", "subdir", "appsync_cache.json")

	cache := &cacheFile{
		Version: cacheVersion,
		Entries: map[string]*cacheEntry{
			"test|us-east-1": {
				URL:       "https://test.appsync-api.us-east-1.amazonaws.com/graphql",
				FetchedAt: time.Now(),
			},
		},
	}

	if err := writeCacheToPath(cacheFilePath, cache); err != nil {
		t.Fatalf("writeCacheToPath: %v", err)
	}

	info, err := os.Stat(cacheFilePath)
	if err != nil {
		t.Fatalf("expected cache file to exist: %v", err)
	}
	if info.IsDir() {
		t.Fatal("expected cache path to be a file")
	}
}

func TestReadCache_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	cacheFilePath := filepath.Join(dir, "nqcli", "appsync_cache.json")

	if err := os.MkdirAll(filepath.Dir(cacheFilePath), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(cacheFilePath, []byte("not json{{{"), 0o644); err != nil {
		t.Fatalf("write corrupted file: %v", err)
	}

	cache, err := readCacheFromPath(cacheFilePath)
	if err != nil {
		t.Fatalf("readCacheFromPath should handle corruption gracefully, got error: %v", err)
	}
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.Version != cacheVersion {
		t.Errorf("version: got %d, want %d", cache.Version, cacheVersion)
	}
	if len(cache.Entries) != 0 {
		t.Fatalf("expected corrupted cache to reset entries, got %d", len(cache.Entries))
	}
}
