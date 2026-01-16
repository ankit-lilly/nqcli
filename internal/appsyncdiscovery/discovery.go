package appsyncdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/aws/aws-sdk-go-v2/service/appsync/types"
)

type ResolveOptions struct {
	Profile string
	APIName string
	APIID   string
}

const cacheVersion = 1

type cacheFile struct {
	Version int                    `json:"version"`
	Entries map[string]*cacheEntry `json:"entries"`
}

type cacheEntry struct {
	URL       string    `json:"url"`
	APIName   string    `json:"api_name,omitempty"`
	APIID     string    `json:"api_id,omitempty"`
	Region    string    `json:"region,omitempty"`
	Profile   string    `json:"profile,omitempty"`
	FetchedAt time.Time `json:"fetched_at"`
}

func ResolveAppSyncURL(ctx context.Context, awsCfg aws.Config, opts ResolveOptions) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if awsCfg.Region == "" {
		return "", fmt.Errorf("AWS region is required to discover the AppSync endpoint")
	}

	cacheKey := cacheKeyFor(opts.Profile, awsCfg.Region)
	if cached := readCacheEntry(cacheKey); cached != "" {
		return cached, nil
	}

	client := appsync.NewFromConfig(awsCfg)

	var (
		selected *types.GraphqlApi
		err      error
	)
	switch {
	case opts.APIID != "":
		selected, err = fetchAPIByID(ctx, client, opts.APIID)
	case opts.APIName != "":
		selected, err = fetchAPIByName(ctx, client, opts.APIName)
	default:
		selected, err = fetchSingleAPI(ctx, client)
	}
	if err != nil {
		return "", err
	}

	url, err := graphqlURL(ctx, client, selected)
	if err != nil {
		return "", err
	}

	writeCacheEntry(cacheKey, awsCfg.Region, opts.Profile, selected, url)

	return url, nil
}

func fetchAPIByID(ctx context.Context, client *appsync.Client, apiID string) (*types.GraphqlApi, error) {
	resp, err := client.GetGraphqlApi(ctx, &appsync.GetGraphqlApiInput{ApiId: aws.String(apiID)})
	if err != nil {
		return nil, fmt.Errorf("get AppSync API %q: %w", apiID, err)
	}
	if resp.GraphqlApi == nil {
		return nil, fmt.Errorf("AppSync API %q not found", apiID)
	}
	return resp.GraphqlApi, nil
}

func fetchAPIByName(ctx context.Context, client *appsync.Client, apiName string) (*types.GraphqlApi, error) {
	apis, err := listAPIs(ctx, client)
	if err != nil {
		return nil, err
	}

	var matches []types.GraphqlApi
	for _, api := range apis {
		if api.Name != nil && *api.Name == apiName {
			matches = append(matches, api)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no AppSync API named %q found", apiName)
	case 1:
		return &matches[0], nil
	default:
		return nil, fmt.Errorf("multiple AppSync APIs named %q found; use NEPTUNE_APPSYNC_API_ID instead", apiName)
	}
}

func fetchSingleAPI(ctx context.Context, client *appsync.Client) (*types.GraphqlApi, error) {
	apis, err := listAPIs(ctx, client)
	if err != nil {
		return nil, err
	}

	if len(apis) == 1 {
		return &apis[0], nil
	}

	var names []string
	for _, api := range apis {
		if api.Name != nil {
			names = append(names, *api.Name)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no AppSync APIs found for the current AWS account and region")
	}
	return nil, fmt.Errorf("multiple AppSync APIs found (%s); set NEPTUNE_APPSYNC_API_NAME or NEPTUNE_APPSYNC_API_ID", strings.Join(names, ", "))
}

func listAPIs(ctx context.Context, client *appsync.Client) ([]types.GraphqlApi, error) {
	var apis []types.GraphqlApi
	paginator := appsync.NewListGraphqlApisPaginator(client, &appsync.ListGraphqlApisInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list AppSync APIs: %w", err)
		}
		apis = append(apis, page.GraphqlApis...)
	}
	return apis, nil
}

func graphqlURL(ctx context.Context, client *appsync.Client, api *types.GraphqlApi) (string, error) {
	if api == nil {
		return "", fmt.Errorf("AppSync API selection was empty")
	}
	if api.Uris != nil {
		if url, ok := api.Uris["GRAPHQL"]; ok && url != "" {
			return url, nil
		}
	}

	if api.ApiId == nil {
		return "", fmt.Errorf("selected AppSync API is missing an ID")
	}

	refreshed, err := fetchAPIByID(ctx, client, *api.ApiId)
	if err != nil {
		return "", err
	}

	if refreshed.Uris == nil {
		return "", fmt.Errorf("AppSync API %q is missing a GraphQL URL", *api.ApiId)
	}
	if url, ok := refreshed.Uris["GRAPHQL"]; ok && url != "" {
		return url, nil
	}
	return "", fmt.Errorf("AppSync API %q is missing a GraphQL URL", *api.ApiId)
}

func cacheKeyFor(profile, region string) string {
	if profile == "" {
		profile = "default"
	}
	if region == "" {
		region = "unknown"
	}
	return profile + "|" + region
}

func readCacheEntry(key string) string {
	cache, err := readCache()
	if err != nil {
		return ""
	}
	entry, ok := cache.Entries[key]
	if !ok || entry == nil {
		return ""
	}
	return entry.URL
}

func writeCacheEntry(key, region, profile string, api *types.GraphqlApi, url string) {
	if url == "" {
		return
	}

	cache, err := readCache()
	if err != nil {
		cache = &cacheFile{Version: cacheVersion, Entries: map[string]*cacheEntry{}}
	}
	if cache.Entries == nil {
		cache.Entries = map[string]*cacheEntry{}
	}

	entry := &cacheEntry{
		URL:       url,
		Region:    region,
		Profile:   profile,
		FetchedAt: time.Now(),
	}
	if api != nil {
		if api.Name != nil {
			entry.APIName = *api.Name
		}
		if api.ApiId != nil {
			entry.APIID = *api.ApiId
		}
	}

	cache.Entries[key] = entry
	_ = writeCache(cache)
}

func readCache() (*cacheFile, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cacheFile{Version: cacheVersion, Entries: map[string]*cacheEntry{}}, nil
		}
		return nil, err
	}

	var cache cacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return &cacheFile{Version: cacheVersion, Entries: map[string]*cacheEntry{}}, nil
	}
	if cache.Entries == nil {
		cache.Entries = map[string]*cacheEntry{}
	}
	return &cache, nil
}

func writeCache(cache *cacheFile) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "appsync-cache-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cache); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nqcli", "appsync_cache.json"), nil
}
