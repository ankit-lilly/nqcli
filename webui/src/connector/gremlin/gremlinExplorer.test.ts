import { afterEach, describe, expect, it, vi } from "vitest";

import { normalizeConnection } from "@/core";

import { createGremlinExplorer } from "./gremlinExplorer";

describe("createGremlinExplorer", () => {
	afterEach(() => vi.unstubAllGlobals());

	it("loads the complete schema from nq", async () => {
		const schema = {
			status: "ready" as const,
			vertices: [],
			edges: [],
			edgeConnections: [],
		};
		const fetchMock = vi.fn().mockResolvedValue(
			new Response(JSON.stringify(schema), {
				headers: { "Content-Type": "application/json" },
			}),
		);
		vi.stubGlobal("fetch", fetchMock);
		const explorer = createGremlinExplorer(
			normalizeConnection({ url: "http://localhost:8080" }),
		);

		await expect(explorer.fetchSchema()).resolves.toStrictEqual(schema);
		expect(fetchMock).toHaveBeenCalledWith(
			"http://localhost:8080/schema?wait=true",
			expect.objectContaining({ method: "GET" }),
		);
	});
});
