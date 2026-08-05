package cmd

import (
	"context"
	"fmt"
	"os"

	mcpcmd "github.com/ankit-lilly/nqcli/cmd/mcp"
	servercmd "github.com/ankit-lilly/nqcli/cmd/server"
	"github.com/ankit-lilly/nqcli/internal/app"
	"github.com/ankit-lilly/nqcli/internal/appsyncdiscovery"
	"github.com/ankit-lilly/nqcli/internal/config"
	neptune "github.com/ankit-lilly/nqcli/internal/gq"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type queryService interface {
	ExecuteCtx(context.Context, string, string) (string, string, error)
	ExecuteQueryCtx(context.Context, string, string) (string, string, error)
}

var (
	envFilePath string
	awsProfile  string
	awsRegion   string
	version     = "dev"
)

const devRESTEndpoint = "https://9nyrl8j1d5-vpce-069388414a9f87f40.execute-api.us-east-2.amazonaws.com/dev/api/v1/internal/neptune/query"

var newGQLClient = func(ctx context.Context) (*neptune.Client, error) {
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

	if awsCfg.Region == "" {
		awsCfg.Region = "us-east-2"
	}

	if cfg.URL == "" {
		profileName := awsProfile
		if profileName == "" {
			profileName = os.Getenv("AWS_PROFILE")
		}
		url, err := appsyncdiscovery.ResolveAppSyncURL(ctx, awsCfg, appsyncdiscovery.ResolveOptions{
			Profile: profileName,
			APIName: cfg.AppSyncAPIName,
			APIID:   cfg.AppSyncAPIID,
		})
		if err != nil {
			cfg.URL = devRESTEndpoint
		} else {
			cfg.URL = url
		}
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("neptune endpoint is required; set NEPTUNE_URL or configure discovery")
	}

	return neptune.NewClient(cfg, awsCfg)
}

var newQueryService = func(ctx context.Context) (queryService, error) {
	neptuneClient, err := newGQLClient(ctx)
	if err != nil {
		return nil, err
	}
	return app.NewAppService(neptuneClient), nil
}

func mcpServiceFactory(ctx context.Context) (mcpcmd.QueryService, error) {
	return newQueryService(ctx)
}

func serverServiceFactory(ctx context.Context) (servercmd.QueryService, error) {
	return newQueryService(ctx)
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

		queryType, _ := cmd.Flags().GetString("type")

		appService, err := newQueryService(cmd.Context())
		if err != nil {
			return err
		}

		var (
			prettyJSON string
			execErr    error
		)

		if inlineQuery != "" {
			prettyJSON, _, execErr = appService.ExecuteQueryCtx(cmd.Context(), inlineQuery, queryType)
		} else {
			prettyJSON, _, execErr = appService.ExecuteCtx(cmd.Context(), queryFile, queryType)
		}
		if execErr != nil {
			l := log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: false})
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

	rootCmd.PersistentFlags().StringVar(&envFilePath, "env-file", "",
		"Path to a .env file to load before executing (defaults to ./ .env, then ~/.env).")
	rootCmd.PersistentFlags().StringVar(&awsProfile, "aws-profile", "",
		"Optional AWS shared config profile to use for authentication.")
	rootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "",
		"Override the AWS region when signing AppSync requests.")

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

	rootCmd.AddCommand(mcpcmd.NewCommand(mcpServiceFactory, version))
	rootCmd.AddCommand(servercmd.NewCommand(serverServiceFactory))
}
