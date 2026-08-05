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

type QueryService interface {
	ExecuteQueryCtx(context.Context, string, string) (string, string, error)
}

func NewCommand(factory func(context.Context) (QueryService, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "server",
		Short:         "Start a web UI for running Neptune queries.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, _ := cmd.Flags().GetString("addr")

			appService, err := factory(cmd.Context())
			if err != nil {
				return err
			}

			logger := log.NewWithOptions(os.Stderr, log.Options{
				ReportTimestamp: true,
				TimeFormat:      time.RFC3339,
			})

			srv := httpserver.New(appService, logger)

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
	return cmd
}
