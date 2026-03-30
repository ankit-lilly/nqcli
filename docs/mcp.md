## MCP Primer for `nq`

This document is a short, practical tutorial for teams new to the Model Context Protocol (MCP), tied directly to how
`nq mcp` works in this repo.

### What MCP is (in one paragraph)

MCP is an open standard that lets AI applications connect to external systems (tools, data sources, workflows) through
a consistent protocol, so clients can discover capabilities and call them safely. The MCP client and server use a
standard JSON-RPC based exchange, while the transport layer can be local process stdio or remote HTTP.

### Core roles in MCP

MCP uses a client-server model with three roles:
- MCP Host: the AI application (Claude Desktop, Cursor, etc.) that manages connections.
- MCP Client: the per-server connection object created by the host.
- MCP Server: a program that exposes capabilities (tools, resources, prompts).

Local MCP servers typically use stdio and serve a single client; remote servers use HTTP and can serve many clients.

### Transports (local vs remote)

MCP supports two transports:
- Stdio transport for local, same-machine processes.
- Streamable HTTP transport for remote servers with standard HTTP authentication patterns.

The protocol and message shapes stay the same across transports. citeturn1view0

### What servers can expose

Servers can expose three capability types:
- Tools: callable functions with JSON Schema inputs and structured outputs.
- Resources: read-only data sources (files, schemas, API responses).
- Prompts: reusable templates to guide model interactions.

Tools are model-controlled and can require user approval in clients. Resources and prompts are user- or application-
controlled.

### The basic MCP lifecycle (tools path)

In typical use:
1. Client initializes the connection and negotiates capabilities.
2. Client lists tools with `tools/list`.
3. Client calls a tool with `tools/call` using the tool’s JSON Schema input.

This is the core loop you’ll see in Claude/Cursor.

### How local MCP servers are configured

Local MCP servers are usually launched by the client and configured with a JSON `mcpServers` block that specifies the
command and args to start the server.

### How MCP works in this repo

`nq mcp` is a local MCP server that uses stdio transport. The MCP client (Claude Desktop, Cursor, etc.) launches the
process and communicates over stdin/stdout.

Capabilities exposed by `nq mcp`:
- `run_gremlin_query`: executes a Gremlin traversal via the same AppSync-backed path as the CLI.
- `get_graph_schema`: returns a static, embedded schema for the clinical-trials graph model. Set
  `NQ_MCP_SCHEMA_SOURCE=dynamic` to run live schema discovery instead (labels, properties, edge
  patterns, counts, and low-cardinality enums).

Internally, the MCP handlers call the existing `AppService.ExecuteQuery(...)` code path, which signs AppSync requests
with your AWS credentials and runs inside your local machine.

### What happens when you ask: “How many studies are in my dev Neptune database?”

Here is the concrete flow when MCP is configured and you ask that question in Claude/Codex:

1. You ask the question in the MCP-enabled client (Claude Desktop/Cursor/Codex).
2. The client decides it needs data from your graph and selects the `run_gremlin_query` tool.
3. The client sends a `tools/call` request to the local `nq mcp` process over stdio with a Gremlin query, for example:
   `g.V().hasLabel('Study').count()`
4. `nq mcp` receives the call and routes it to the Go handler in `nqcli/cmd/mcp.go`.
5. The handler calls `AppService.ExecuteQuery(query, "gremlin")`.
6. `AppService` creates (or reuses) a Neptune client using your local AWS credentials and AppSync endpoint resolution.
7. The Neptune client:
   - Builds the GraphQL mutation payload: `executeQuery(input: { type: "gremlin", query: ... })`
   - Signs the request with SigV4 using `AWS_PROFILE` (e.g., `dsoadev`).
   - Sends the request to the AppSync endpoint in your VPC.
8. AppSync runs the query (via the Lambda in the VPC) against Neptune.
9. The GraphQL response comes back to `nq`.
10. `nq` extracts the JSON result and returns it to the MCP server.
11. The MCP server returns the JSON as the tool response.
12. The client presents the result and writes a natural-language answer.

The important point: the LLM never connects to Neptune directly. It only calls the MCP tool, and the tool uses your
local credentials and VPC-connected AppSync/Lambda path to execute the query.

### Environment selection (dev/qa)

Use separate MCP server entries in your client config, each with a different `AWS_PROFILE`. This lets you switch
environments without restarting the client. See the example in `nqcli/README.md`.
