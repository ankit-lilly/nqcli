package cmd

import (
	"fmt"
	"os"

	"github.com/ankit-lilly/nqcli/internal/app"
	"github.com/ankit-lilly/nqcli/internal/config"
	neptune "github.com/ankit-lilly/nqcli/internal/gq"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type queryService interface {
	Execute(string, string) (string, string, error)
	ExecuteQuery(string, string) (string, string, error)
}

var newQueryService = func() queryService {
	cfg := config.LoadConfig()
	neptuneClient := neptune.NewClient(cfg)
	return app.NewAppService(neptuneClient)
}

var rootCmd = &cobra.Command{
	Use:   "nq-cli [query_file|query]",
	Short: "Execute Gremlin or Cypher queries against a Neptune GraphQL endpoint.",
	Long: `A CLI tool to execute Gremlin or Cypher queries against a Neptune GraphQL endpoint.

Usage:
  echo "query" | nq-cli [--type gremlin|cypher]
  nq-cli [--type gremlin|cypher] "query"
  nq-cli [--type gremlin|cypher] <query_file>`,
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

		appService := newQueryService()

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
		os.Exit(1)
	}
}

func init() {
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
