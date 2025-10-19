package cmd

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

func init() {
	rootCmd.AddCommand(newServerCommand())
}

func newServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "server",
		Short:         "Start a web UI for running Neptune queries.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := cmd.Flags().GetString("addr")
			if err != nil {
				return err
			}

			appService := newQueryService()

			logger := log.NewWithOptions(os.Stderr, log.Options{
				ReportTimestamp: true,
				TimeFormat:      time.RFC3339,
			})

			server := httpserver.New(appService, logger)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := server.Start(ctx, addr); err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("server stopped with error", "error", err)
				return err
			}

			logger.Info("server stopped")
			return nil
		},
	}

	cmd.Flags().String("addr", ":8080", "Address to bind the HTTP server to.")

	return cmd
}
