package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ankit-lilly/nqcli/internal/bootstrap"
	"github.com/ankit-lilly/nqcli/internal/config"
	"github.com/ankit-lilly/nqcli/internal/core"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	config.LoadEnvironment("")

	profile := os.Getenv("AWS_PROFILE")
	svc, err := bootstrap.Build(context.Background(), bootstrap.Options{Profile: profile})
	if err != nil {
		return fmt.Errorf("initialize query service: %w", err)
	}

	factory := func(ctx context.Context, p string) (core.QueryService, error) {
		return bootstrap.Build(ctx, bootstrap.Options{Profile: p})
	}

	return runApp(svc, profile, factory)
}
