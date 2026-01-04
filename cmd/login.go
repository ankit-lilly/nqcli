package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ankit-lilly/nqcli/internal/config"
	"github.com/ankit-lilly/nqcli/internal/login"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const defaultSecretName = "lrl-dtf-sdr-api-auth-secrets"

func init() {
	rootCmd.AddCommand(newLoginCommand())
}

func newLoginCommand() *cobra.Command {
	var (
		secretName = defaultSecretName
		awsProfile string
		awsRegion  string
		noWrite    bool
		printToken bool
	)

	cmd := &cobra.Command{
		Use:           "login",
		Short:         "Fetch a fresh NEPTUNE_TOKEN using Azure AD client credentials.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			logger := log.NewWithOptions(os.Stderr, log.Options{
				ReportTimestamp: false,
			})

			cfgOpts := []func(*awscfg.LoadOptions) error{}
			if awsProfile != "" {
				cfgOpts = append(cfgOpts, awscfg.WithSharedConfigProfile(awsProfile))
			}
			if awsRegion != "" {
				cfgOpts = append(cfgOpts, awscfg.WithRegion(awsRegion))
			}

			awsCfg, err := awscfg.LoadDefaultConfig(ctx, cfgOpts...)
			if err != nil {
				return fmt.Errorf("load AWS configuration: %w", err)
			}

			secretsClient := secretsmanager.NewFromConfig(awsCfg)
			loginService := login.NewService(secretName, secretsClient, nil)

			result, err := loginService.Login(ctx)
			if err != nil {
				return err
			}

			if !noWrite {
				envPath, err := config.ResolveEnvFileForWrite(envFilePath)
				if err != nil {
					return err
				}
				if err := config.WriteEnvValue(envPath, "NEPTUNE_TOKEN", result.AccessToken); err != nil {
					return err
				}
				logger.Info("Updated NEPTUNE_TOKEN", "env_file", envPath)
			}

			if printToken {
				fmt.Fprintln(cmd.OutOrStdout(), result.AccessToken)
			}

			expires := result.ExpiresAt.Local().Format(time.RFC1123)
			logger.Info("Token acquired", "token_type", result.TokenType, "expires", expires)
			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", secretName, "AWS Secrets Manager secret that holds the API auth credentials.")
	cmd.Flags().StringVar(&awsProfile, "aws-profile", "dsoadev", "Optional AWS shared config profile to use for authentication.")
	cmd.Flags().StringVar(&awsRegion, "aws-region", "us-east-2", "Override the AWS region when loading credentials.")
	cmd.Flags().BoolVar(&noWrite, "no-write", false, "Do not write NEPTUNE_TOKEN to an env file.")
	cmd.Flags().BoolVar(&printToken, "print-token", false, "Print the raw access token to stdout (use with caution).")

	return cmd
}
