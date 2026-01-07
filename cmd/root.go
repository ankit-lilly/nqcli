package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ankit-lilly/nqcli/internal/app"
	"github.com/ankit-lilly/nqcli/internal/config"
	neptune "github.com/ankit-lilly/nqcli/internal/gq"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type queryService interface {
	Execute(string, string) (string, string, error)
	ExecuteQuery(string, string) (string, string, error)
}

var (
	envFilePath string
	awsProfile  string
	awsRegion   string
	version     = "dev"
)

var newQueryService = func(ctx context.Context) (queryService, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := config.LoadConfig()

	cfgOpts := []func(*awscfg.LoadOptions) error{}
	if awsProfile != "" {
		cfgOpts = append(cfgOpts, awscfg.WithSharedConfigProfile(awsProfile))
	}
	if awsRegion != "" {
		cfgOpts = append(cfgOpts, awscfg.WithRegion(awsRegion))
	}

	awsCfg, err := awscfg.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}

	neptuneClient, err := neptune.NewClient(cfg, awsCfg)
	if err != nil {
		return nil, err
	}
	return app.NewAppService(neptuneClient), nil
}

var rootCmd = &cobra.Command{
	Use:   "nq [query_file|query]",
	Short: "Execute Gremlin or Cypher queries against a Neptune GraphQL endpoint.",
	Long: `A CLI tool to execute Gremlin or Cypher queries against a Neptune GraphQL endpoint.
	Usage:
	    echo "query" | nq [--type gremlin|cypher]
	    nq [--type gremlin|cypher] "query"
	    nq [--type gremlin|cypher] <query_file>
	`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			queryFile   string
			inlineQuery string
		)
		if len(args) > 0 {
			input := args[0]
			info, err := os.Stat(input)
			switch {
			case err == nil && info.IsDir():
				return fmt.Errorf("provided path %q is a directory, expected a file", input)
			case err == nil:
				queryFile = input
			case os.IsNotExist(err):
				inlineQuery = input
			default:
				return fmt.Errorf("failed to stat %q: %w", input, err)
			}
		}

		queryType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		appService, err := newQueryService(cmd.Context())
		if err != nil {
			return err
		}

		l := log.NewWithOptions(os.Stderr, log.Options{
			ReportTimestamp: false,
		})

		var (
			prettyJSON string
			execErr    error
		)

		if inlineQuery != "" {
			prettyJSON, _, execErr = appService.ExecuteQuery(inlineQuery, queryType)
		} else {
			prettyJSON, _, execErr = appService.Execute(queryFile, queryType)
		}
		if execErr != nil {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
			l.Error(style.Render("Error"), "details", execErr)
			return execErr
		}

		fmt.Println(prettyJSON)
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

	rootCmd.PersistentFlags().StringVar(
		&envFilePath,
		"env-file",
		"",
		"Path to a .env file to load before executing (defaults to ./ .env, then ~/.env).",
	)
	rootCmd.PersistentFlags().StringVar(
		&awsProfile,
		"aws-profile",
		"",
		"Optional AWS shared config profile to use for authentication.",
	)
	rootCmd.PersistentFlags().StringVar(
		&awsRegion,
		"aws-region",
		"",
		"Override the AWS region when signing AppSync requests.",
	)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := config.LoadEnvironment(envFilePath); err != nil {
			return err
		}
		return nil
	}

	rootCmd.Flags().String(
		"type",
		"gremlin",
		"The type of query to execute. Must be 'gremlin' or 'cypher'.",
	)

	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		queryType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		if queryType != "gremlin" && queryType != "cypher" {
			return fmt.Errorf("invalid value for --type: %s. Must be 'gremlin' or 'cypher'", queryType)
		}
		return nil
	}
}
