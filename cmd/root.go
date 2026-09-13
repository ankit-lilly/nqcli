package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	mcpcmd "github.com/ankit-lilly/nqcli/cmd/mcp"
	servercmd "github.com/ankit-lilly/nqcli/cmd/server"
	"github.com/ankit-lilly/nqcli/internal/bootstrap"
	"github.com/ankit-lilly/nqcli/internal/config"
	"github.com/ankit-lilly/nqcli/internal/core"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	envFilePath string
	awsProfile  string
	awsRegion   string
	version     = "dev"
)

var newQueryService = func(ctx context.Context) (core.QueryService, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := bootstrap.Options{
		Profile: awsProfile,
		Region:  awsRegion,
	}

	cfg := config.LoadConfig()

	opts.URL = cfg.URL
	return bootstrap.Build(ctx, opts)
}

func serviceFactory(ctx context.Context) (core.QueryService, error) {
	return newQueryService(ctx)
}

func profileServiceFactory(ctx context.Context, profile string) (core.QueryService, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := bootstrap.Options{
		Profile: profile,
		Region:  awsRegion,
	}

	cfg := config.LoadConfig()
	opts.URL = cfg.URL
	return bootstrap.Build(ctx, opts)
}

func readQuery(args []string) (query string, err error) {
	if len(args) > 0 {
		input := args[0]
		info, statErr := os.Stat(input)
		switch {
		case statErr == nil && info.IsDir():
			return "", fmt.Errorf("provided path %q is a directory, expected a file", input)
		case statErr == nil:
			content, readErr := os.ReadFile(input)
			if readErr != nil {
				return "", fmt.Errorf("failed to read query file: %w", readErr)
			}
			return string(content), nil
		case os.IsNotExist(statErr):
			return input, nil
		default:
			return "", fmt.Errorf("failed to stat %q: %w", input, statErr)
		}
	}

	fi, statErr := os.Stdin.Stat()
	if statErr != nil {
		return "", fmt.Errorf("failed to stat stdin: %w", statErr)
	}
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		content, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			return "", fmt.Errorf("failed to read stdin: %w", readErr)
		}
		return string(content), nil
	}

	return "", fmt.Errorf("no query provided. Use 'echo \"query\" | nq' or 'nq <query_file>' or 'nq \"query\"'")
}

var rootCmd = &cobra.Command{
	Use:   "nq [query_file|query]",
	Short: "Execute Gremlin or Cypher queries against SDR.",
	Long: `A CLI tool to execute Gremlin or Cypher queries against SDR Neptune.
	Usage:
	    echo "query" | nq [--type gremlin|cypher]
	    nq [--type gremlin|cypher] "query"
	    nq [--type gremlin|cypher] <query_file>
	`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		query, err := readQuery(args)
		if err != nil {
			return err
		}

		queryType, _ := cmd.Flags().GetString("type")

		service, err := newQueryService(cmd.Context())
		if err != nil {
			return err
		}

		result, execErr := service.ExecuteQuery(cmd.Context(), query, queryType, core.QueryOpts{})
		if execErr != nil {
			l := log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: false})
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
			l.Error(style.Render("Error"), "details", execErr)
			return execErr
		}

		fmt.Println(result.Processed)
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.PersistentFlags().StringVar(&envFilePath, "env-file", "",
		"Path to a .env file to load before executing (defaults to ./ .env, then ~/.env).")
	rootCmd.PersistentFlags().StringVar(&awsProfile, "aws-profile", "",
		"Optional AWS shared config profile to use for authentication.")
	rootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "",
		"Override the AWS region.")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return config.LoadEnvironment(envFilePath)
	}

	rootCmd.Flags().String("type", "gremlin",
		"The type of query to execute. Must be 'gremlin' or 'cypher'.")

	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		queryType, _ := cmd.Flags().GetString("type")
		if queryType != "gremlin" && queryType != "cypher" {
			return fmt.Errorf("invalid value for --type: %s. Must be 'gremlin' or 'cypher'", queryType)
		}
		return nil
	}

	rootCmd.AddCommand(mcpcmd.NewCommand(serviceFactory, version))
	rootCmd.AddCommand(servercmd.NewCommand(profileServiceFactory, func() string { return awsProfile }))
}
