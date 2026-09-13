import { beforeEach, describe, expect, it, vi } from "vitest";

import {
	activeConfigurationAtom,
	type AppStore,
	configurationAtom,
	createVertexType,
	explorerForTestingAtom,
	getAppStore,
	schemaAtom,
} from "@/core";
import { createQueryClient } from "@/core/queryClient";
import {
	createRandomEdgeConnection,
	createRandomRawConfiguration,
	FakeExplorer,
} from "@/utils/testing";

import { schemaSyncQuery } from "./schemaSyncQuery";

describe("schemaSyncQuery", () => {
	let explorer: FakeExplorer;
	let store: AppStore;

	beforeEach(() => {
		explorer = new FakeExplorer();
		store = getAppStore();
		const config = createRandomRawConfiguration();
		store.set(configurationAtom, new Map([[config.id, config]]));
		store.set(activeConfigurationAtom, config.id);
		store.set(schemaAtom, new Map());
		store.set(explorerForTestingAtom, explorer);
	});

	function options(hasConnection = true) {
		return schemaSyncQuery({
			connectionId: store.get(activeConfigurationAtom),
			activeSchema: undefined,
			hasConnection,
		});
	}

	it("stores one complete backend snapshot", async () => {
		const edgeConnection = createRandomEdgeConnection();
		const lastUpdate = "2026-09-12T12:00:00Z";
		vi.spyOn(explorer, "fetchSchema").mockResolvedValue({
			status: "ready",
			lastUpdate,
			vertices: [{ type: createVertexType("Study"), attributes: [] }],
			edges: [],
			edgeConnections: [edgeConnection],
		});

		const result = await createQueryClient().fetchQuery(options());

		expect(result.edgeConnections).toStrictEqual([edgeConnection]);
		expect(result.lastUpdate).toStrictEqual(new Date(lastUpdate));
		expect(store.get(schemaAtom).values().next().value).toStrictEqual(result);
	});

	it("normalizes a legacy response without edge connections", async () => {
		vi.spyOn(explorer, "fetchSchema").mockResolvedValue({
			vertices: [],
			edges: [],
		} as never);

		const result = await createQueryClient().fetchQuery(options());

		expect(result.edgeConnections).toStrictEqual([]);
	});

	it("does not replace cached data when fetching fails", async () => {
		const id = store.get(activeConfigurationAtom)!;
		const cached = {
			vertices: [{ type: createVertexType("Study"), attributes: [] }],
			edges: [],
			edgeConnections: [],
		};
		store.set(schemaAtom, new Map([[id, cached]]));
		vi.spyOn(explorer, "fetchSchema").mockRejectedValue(new Error("offline"));

		await expect(createQueryClient().fetchQuery(options())).rejects.toThrow(
			"offline",
		);
		expect(store.get(schemaAtom).get(id)).toBe(cached);
	});

	it("does not persist an incomplete backend snapshot", async () => {
		vi.spyOn(explorer, "fetchSchema").mockResolvedValue({
			status: "running",
			vertices: [],
			edges: [],
			edgeConnections: [],
		});

		await expect(createQueryClient().fetchQuery(options())).rejects.toThrow(
			"Schema sync is running",
		);
		expect(store.get(schemaAtom).size).toBe(0);
	});

	it("is disabled without a connection", () => {
		expect(options(false).enabled).toBe(false);
	});
});
