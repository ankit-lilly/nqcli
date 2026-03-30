# Neptune NQ MCP Server

## Usage with Claude Desktop

```json
{
  "mcpServers": {
    "neptune": {
      "command": "node",
      "args": ["/absolute/path/to/nqcli/mcp/build/index.js"],
      "env": {
        "AWS_PROFILE": "your-profile-name",
        "NEPTUNE_URL": "https://your-appsync-endpoint/graphql"
      }
    }
  }
}
```

## Tools Provided

1.  **`run_gremlin_query`**: Takes a Gremlin string, runs it via `nq`, and returns the JSON result.
2.  **`get_graph_schema`**: Runs a `groupCount()` query to tell the AI what labels exist in your graph.

## Why this is great for learning Gremlin

You can now ask Claude:
*   "Show me the schema of my graph."
*   "Write a Gremlin query to find the shortest path between vertex 'A' and 'B' and then execute it for me."
*   "Explain what this query result means in plain English."
