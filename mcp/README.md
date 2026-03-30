# Neptune NQ MCP Server

This is an MCP (Model Context Protocol) server that wraps your `nq` CLI tool. It allows AI agents like Claude to query your Neptune database using natural language by executing your `nq` binary behind the scenes.

## Prerequisites

1.  **Build your CLI**: Ensure the `nq` binary is built and in your system PATH.
    ```bash
    cd ..
    make build
    sudo cp nq /usr/local/bin/
    ```
2.  **Node.js**: Ensure you have Node.js 18+ installed.

## Setup

1.  **Install dependencies**:
    ```bash
    npm install
    ```
2.  **Build the MCP server**:
    ```bash
    npm run build
    ```

## Usage with Claude Desktop

Add the following to your Claude Desktop configuration file (usually `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS or `%APPDATA%\Claude\claude_desktop_config.json` on Windows):

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
