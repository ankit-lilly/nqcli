import { queryOptions } from "@tanstack/react-query";
import { atom } from "jotai";

import {
	activeConfigurationAtom,
	type ConfigurationId,
	schemaAtom,
} from "@/core";
import {
	activeSchemaAtom,
	type SchemaStorageModel,
} from "@/core/StateProvider/schema";
import { logger } from "@/utils";

import type { SchemaResponse } from "../useGEFetchTypes";

import { getConnectionApplication, getStore } from "./helpers";

/** Returns the query key for the schema sync query for the given connection. */
export function schemaSyncQueryKey(connectionId: ConfigurationId | null) {
	return ["schema", "discovery", connectionId] as const;
}

/**
 * Fetches the schema from the given explorer and persists it to the local cache on success.
 *
 * Uses `staleTime: Infinity` so the query only runs when no cached data exists
 * or when manually triggered via `refetch()`.
 *
 */
export function schemaSyncQuery({
	connectionId,
	activeSchema,
	hasConnection,
}: {
	connectionId: ConfigurationId | null;
	activeSchema: SchemaStorageModel | undefined;
	hasConnection: boolean;
}) {
	return queryOptions({
		queryKey: schemaSyncQueryKey(connectionId),
		staleTime: Infinity,
		initialData: activeSchema,
		enabled: hasConnection,
		queryFn: async ({ signal, meta }) => {
			const application = getConnectionApplication(meta);
			const store = getStore(meta);
			const schema = await application.schema.get({ signal });
			if (schema.status && schema.status !== "ready") {
				throw new Error(`Schema sync is ${schema.status}`);
			}

			store.set(replaceSchemaAtom, schema);
			return store.get(activeSchemaAtom);
		},
	});
}

/** Setter-only atom that replaces the stored schema with the given schema response. */
const replaceSchemaAtom = atom(null, (get, set, schema: SchemaResponse) => {
	const id = get(activeConfigurationAtom);
	if (!id) {
		logger.warn("Cannot update schema: no active configuration");
		return;
	}

	set(schemaAtom, (prev) => {
		const updated = new Map(prev);
		updated.set(id, {
			vertices: schema.vertices,
			edges: schema.edges,
			edgeConnections: schema.edgeConnections ?? [],
			totalVertices: schema.totalVertices,
			totalEdges: schema.totalEdges,
			lastUpdate: schema.lastUpdate ? new Date(schema.lastUpdate) : new Date(),
		});
		return updated;
	});
});
