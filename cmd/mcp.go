package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

type runGremlinArgs struct {
	Query string `json:"query" jsonschema:"The Gremlin traversal string to execute"`
}

func init() {
	rootCmd.AddCommand(newMcpCommand())
}

func newMcpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "mcp",
		Short:         "Start an MCP server over stdio for running Neptune queries.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			appService, err := newQueryService(cmd.Context())
			if err != nil {
				return err
			}

			server := mcp.NewServer(&mcp.Implementation{
				Name:    "nq-neptune-mcp",
				Version: version,
			}, nil)

			mcp.AddTool(
				server,
				&mcp.Tool{
					Name:        "run_gremlin_query",
					Description: "Run a Gremlin query against Neptune. Returns the JSON result from the database.",
				},
				func(ctx context.Context, req *mcp.CallToolRequest, args runGremlinArgs) (*mcp.CallToolResult, any, error) {
					query := strings.TrimSpace(args.Query)
					if query == "" {
						return nil, nil, fmt.Errorf("query cannot be empty")
					}

					prettyJSON, _, execErr := appService.ExecuteQuery(query, "gremlin")
					if execErr != nil {
						return nil, nil, execErr
					}

					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: prettyJSON},
						},
					}, nil, nil
				},
			)

			mcp.AddTool(
				server,
				&mcp.Tool{
					Name:        "get_graph_schema",
					Description: "Discovers vertex labels and their counts to help understand the graph structure.",
				},
				func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
					prettyJSON, execErr := buildGraphSchema(ctx, appService)
					if execErr != nil {
						return nil, nil, execErr
					}

					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: prettyJSON},
						},
					}, nil, nil
				},
			)

			return server.Run(cmd.Context(), &mcp.StdioTransport{})
		},
	}

	return cmd
}
