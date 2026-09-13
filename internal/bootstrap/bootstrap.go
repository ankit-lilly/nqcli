package bootstrap

import (
	"context"
	"fmt"

	"github.com/ankit-lilly/nqcli/internal/config"
	"github.com/ankit-lilly/nqcli/internal/core"
	"github.com/ankit-lilly/nqcli/internal/neptune"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
)

type Options struct {
	Profile string
	Region  string
	URL     string
}

func Build(ctx context.Context, opts Options) (core.QueryService, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := config.LoadConfig()

	cfgOpts := []func(*awscfg.LoadOptions) error{}
	if opts.Profile != "" {
		cfgOpts = append(cfgOpts, awscfg.WithSharedConfigProfile(opts.Profile))
	}
	if opts.Region != "" {
		cfgOpts = append(cfgOpts, awscfg.WithRegion(opts.Region))
	}

	awsCfg, err := awscfg.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}

	if awsCfg.Region == "" {
		awsCfg.Region = "us-east-2"
	}

	if opts.URL != "" {
		cfg.URL = opts.URL
	}
	if cfg.URL == "" {
		cfg.URL = config.FallbackEndpoint(opts.Profile)
	}
	if cfg.DirectURL == "" {
		if directURL, discoverErr := neptune.DiscoverWriterEndpoint(ctx, awsCfg, cfg.ClusterID); discoverErr == nil {
			cfg.DirectURL = directURL
		}
	}

	client, err := neptune.NewClient(cfg, awsCfg)
	if err != nil {
		return nil, err
	}

	return core.NewService(client), nil
}
