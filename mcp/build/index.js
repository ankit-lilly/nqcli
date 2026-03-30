#!/usr/bin/env node
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { CallToolRequestSchema, ListToolsRequestSchema, } from "@modelcontextprotocol/sdk/types.js";
import { runGremlinQuery, getGraphSchema } from "./nq-runner.js";
const AWS_PROFILE = process.env.AWS_PROFILE;
const server = new Server({
    name: "nq-neptune-mcp",
    version: "1.0.0",
}, {
    capabilities: {
        tools: {},
    },
});
server.setRequestHandler(ListToolsRequestSchema, async () => {
    return {
        tools: [
            {
                name: "run_gremlin_query",
                description: "Run a Gremlin query against Neptune. The tool returns the JSON result from the database.",
                inputSchema: {
                    type: "object",
                    properties: {
                        query: {
                            type: "string",
                            description: "The Gremlin traversal string (e.g., g.V().count())",
                        },
                    },
                    required: ["query"],
                },
            },
            {
                name: "get_graph_schema",
                description: "Discovers vertex labels and their counts to help understand the graph structure.",
                inputSchema: {
                    type: "object",
                    properties: {},
                },
            },
        ],
    };
});
server.setRequestHandler(CallToolRequestSchema, async (request) => {
    const { name, arguments: args } = request.params;
    try {
        if (name === "run_gremlin_query") {
            const query = String(args?.query);
            const result = await runGremlinQuery(query, AWS_PROFILE);
            return {
                content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
            };
        }
        else if (name === "get_graph_schema") {
            const schema = await getGraphSchema(AWS_PROFILE);
            return {
                content: [{ type: "text", text: schema }],
            };
        }
        else {
            throw new Error(`Unknown tool: ${name}`);
        }
    }
    catch (error) {
        return {
            isError: true,
            content: [{ type: "text", text: `Error: ${error.message}` }],
        };
    }
});
async function main() {
    const transport = new StdioServerTransport();
    await server.connect(transport);
}
main().catch((error) => {
    console.error("Fatal error:", error);
    process.exit(1);
});
