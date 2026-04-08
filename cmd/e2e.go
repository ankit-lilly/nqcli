package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ankit-lilly/nqcli/internal/e2e"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newE2ECommand())
}

func newE2ECommand() *cobra.Command {
	var (
		scenario    string
		timeout     time.Duration
		trialPrefix string
		verbose     bool
		cleanupOnly bool
	)

	cmd := &cobra.Command{
		Use:   "e2e",
		Short: "Run end-to-end integration tests against a deployed SDR environment.",
		Long: `Run integration test scenarios that exercise the full SDR pipeline:
submit data via GraphQL → async processing → verify in Neptune → cleanup.

Requires valid AWS credentials (e.g. aws sso login --profile dsoadev).

Examples:
  nq e2e --aws-profile dsoadev
  nq e2e --aws-profile dsoadev --scenario submit-and-verify --verbose
  nq e2e --aws-profile dsoadev --cleanup-only`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			gqlClient, err := newGQLClient(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize: %w", err)
			}

			client := e2e.NewSDRClient(gqlClient, verbose)

			cfg := e2e.RunnerConfig{
				Scenario:    scenario,
				Timeout:     timeout,
				TrialPrefix: trialPrefix,
				Verbose:     verbose,
				CleanupOnly: cleanupOnly,
			}

			report := e2e.Run(client, cfg)
			e2e.PrintReport(report)

			if report.HasFailures() {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&scenario, "scenario", "", "Run a specific scenario by name (default: all).")
	cmd.Flags().DurationVar(&timeout, "timeout", 120*time.Second, "Max wait time for async operations.")
	cmd.Flags().StringVar(&trialPrefix, "trial-prefix", "TST-E2", "Trial alias prefix for test data.")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed GraphQL queries and responses.")
	cmd.Flags().BoolVar(&cleanupOnly, "cleanup-only", false, "Only clean up test data without running tests.")

	return cmd
}
