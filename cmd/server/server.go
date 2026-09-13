package server

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpserver "github.com/ankit-lilly/nqcli/internal/server"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func NewCommand(factory httpserver.ServiceFactory, currentProfile func() string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "server",
		Short:         "Start a web UI for running Neptune queries.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, _ := cmd.Flags().GetString("addr")
			queryType, _ := cmd.Flags().GetString("type")
			if queryType != "gremlin" && queryType != "cypher" {
				return errors.New("--type must be 'gremlin' or 'cypher'")
			}
			queryEngine := queryType
			if queryType == "cypher" {
				queryEngine = "openCypher"
			}
			profile := currentProfile()

			service, err := factory(cmd.Context(), profile)
			if err != nil {
				return err
			}

			logger := log.NewWithOptions(os.Stderr, log.Options{
				ReportTimestamp: true,
				TimeFormat:      time.RFC3339,
			})

			srv := httpserver.NewWithOptions(service, logger, httpserver.Options{
				Profile:        profile,
				QueryEngine:    queryEngine,
				ServiceFactory: factory,
			})

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := srv.Start(ctx, addr); err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("server stopped with error", "error", err)
				return err
			}
			return nil
		},
	}

	cmd.Flags().String("addr", ":8080", "Address to bind the HTTP server to.")
	cmd.Flags().String("type", "gremlin", "Query language used by the web UI: gremlin or cypher.")
	return cmd
}
