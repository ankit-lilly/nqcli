import { beforeEach, describe, expect, it, vi } from "vitest";

import { normalizeConnection } from "@/core";
import { NetworkError, ServerConnectionError } from "@/utils";

import { fetchDatabaseRequest } from "./fetchDatabaseRequest";

const connection = normalizeConnection({
	url: "http://localhost:8080",
	queryEngine: "gremlin",
});

describe("fetchDatabaseRequest", () => {
	beforeEach(() => vi.restoreAllMocks());

	it("sends the request and decodes JSON", async () => {
		const fetchMock = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ ok: true }), {
				status: 200,
				headers: { "Content-Type": "application/json" },
			}),
		);
		vi.stubGlobal("fetch", fetchMock);

		await expect(
			fetchDatabaseRequest(connection, "/gremlin", {
				method: "POST",
				headers: { Accept: "application/json" },
			}),
		).resolves.toStrictEqual({ ok: true });
		expect(fetchMock).toHaveBeenCalledWith(
			"/gremlin",
			expect.objectContaining({
				method: "POST",
				headers: { accept: "application/json" },
			}),
		);
	});

	it("turns a JSON error response into a NetworkError", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue(
				new Response(JSON.stringify({ error: "bad query" }), {
					status: 400,
					headers: { "Content-Type": "application/json" },
				}),
			),
		);

		await expect(
			fetchDatabaseRequest(connection, "/gremlin", {}),
		).rejects.toBeInstanceOf(NetworkError);
	});

	it("turns connectivity failures into ServerConnectionError", async () => {
		vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("offline")));

		await expect(
			fetchDatabaseRequest(connection, "http://localhost:8080/gremlin", {}),
		).rejects.toBeInstanceOf(ServerConnectionError);
	});
});
