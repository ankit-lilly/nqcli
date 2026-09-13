import { describe, expect, it } from "vitest";

import { createNewConfigurationId, createVertexType } from "@/core";

import {
	mergeConfiguration,
	normalizeConnection,
	normalizeUrl,
} from "./configuration";

describe("normalizeConnection", () => {
	it("normalizes the API URL and defaults to Gremlin", () => {
		expect(normalizeConnection({ url: "  http://localhost:8080/\n" })).toEqual({
			url: "http://localhost:8080",
			queryEngine: "gremlin",
		});
	});

	it("preserves UI request options", () => {
		expect(
			normalizeConnection({
				url: "http://localhost:8080",
				queryEngine: "openCypher",
				fetchTimeoutMs: 1000,
				nodeExpansionLimit: 20,
			}),
		).toMatchObject({
			queryEngine: "openCypher",
			fetchTimeoutMs: 1000,
			nodeExpansionLimit: 20,
		});
	});
});

describe("mergeConfiguration", () => {
	it("combines schema data with domain defaults", () => {
		const type = createVertexType("Study");
		const result = mergeConfiguration(
			{
				vertices: [{ type, attributes: [] }],
				edges: [],
				edgeConnections: [],
			},
			{
				id: createNewConfigurationId(),
				connection: { url: "http://localhost:8080" },
			},
			new Map(),
			new Map(),
		);

		expect(result.schema.vertices[0]).toMatchObject({ type });
		expect(result.schema.edgeConnections).toStrictEqual([]);
	});
});

describe("normalizeUrl", () => {
	it("handles missing URLs", () => expect(normalizeUrl(undefined)).toBe(""));
});
