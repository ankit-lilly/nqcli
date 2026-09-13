import { describe, expect, it, vi } from "vitest";

import { fetchDefaultConnection, mapToConnection } from "./defaultConnection";

describe("default connection", () => {
	it("maps the nq API contract", () => {
		expect(
			mapToConnection({
				apiBaseUrl: "http://localhost:8080",
				queryEngine: "openCypher",
				profile: "dsoadev",
			}),
		).toMatchObject({
			id: "profile:dsoadev:openCypher",
			displayLabel: "dsoadev",
			connection: {
				url: "http://localhost:8080",
				queryEngine: "openCypher",
			},
		});
	});

	it("loads the connection from the same-origin server", async () => {
		vi.stubGlobal("location", { origin: "http://localhost:8080" });
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue(
				new Response(
					JSON.stringify({
						apiBaseUrl: "http://localhost:8080",
						queryEngine: "gremlin",
						profile: "dsoadev",
					}),
					{ status: 200, headers: { "Content-Type": "application/json" } },
				),
			),
		);

		const result = await fetchDefaultConnection();

		expect(fetch).toHaveBeenCalledWith(
			"http://localhost:8080/defaultConnection",
		);
		expect(result?.displayLabel).toBe("dsoadev");
	});
});
