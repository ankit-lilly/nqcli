import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useAtomValue } from "jotai";
import { useEffect, useMemo } from "react";

import { createConnectionApplication } from "@/bootstrap";
import { schemaSyncQuery } from "@/connector";
import {
	activeConfigurationAtom,
	maybeActiveSchemaAtom,
	useConfiguration,
	useExplorer,
} from "@/core";
import { logger } from "@/utils";

/** Keeps the browser cache synchronized with the backend schema cache. */
export function useSchemaRevalidation() {
	const config = useConfiguration();
	const explorer = useExplorer();
	const queryClient = useQueryClient();
	const application = useMemo(
		() =>
			createConnectionApplication(explorer, (error) => {
				logger.warn("Schema sync event stream error", error);
			}),
		[explorer],
	);

	useEffect(() => {
		const endpoint = config?.connection?.url;
		if (!endpoint) {
			return;
		}

		void application.schema.revalidate().catch((error) => {
			logger.warn("Failed to start background schema refresh", error);
		});

		const unsubscribe = application.schema.subscribe((event) => {
			if (
				event.type === "schema.sync.completed" ||
				(event.type === "schema.sync.snapshot" && event.status === "ready")
			) {
				void queryClient.invalidateQueries({ queryKey: ["schema"] });
			}
		});

		return unsubscribe;
	}, [application, config?.connection?.url, queryClient]);
}

/**
 * Subscribes to the backend-owned schema snapshot.
 *
 * The query uses `staleTime: Infinity` and `initialData` from the Jotai
 * store. This means:
 * - If cached data exists in localforage, it seeds the query cache and no
 *   fetch occurs.
 * - If no cached data exists, TanStack Query fetches automatically.
 * - Manual refetch is available via `refreshSchema()`.
 * - On refetch failure, TanStack Query preserves the previous successful data.
 */
export function useSchemaSync() {
	const config = useConfiguration();
	const explorer = useExplorer();
	const application = useMemo(
		() => createConnectionApplication(explorer),
		[explorer],
	);
	// Read the atom directly instead of useMaybeActiveSchema() because that hook
	// wraps the value in useDeferredValue, which delays the update by one render.
	// The schema and connectionId must update in the same render so the query
	// options stay consistent when switching connections.
	const activeSchema = useAtomValue(maybeActiveSchemaAtom);
	const connectionId = useAtomValue(activeConfigurationAtom);

	const schemaQuery = useQuery(
		schemaSyncQuery({
			connectionId,
			activeSchema,
			hasConnection: config != null,
		}),
	);
	const refreshSchema = async () => {
		logger.log("Refreshing schema");
		if (config?.connection?.url) {
			await application.schema.refresh();
		}
		await schemaQuery.refetch();
	};

	return {
		schemaQuery,
		refreshSchema,
		isFetching: schemaQuery.isFetching,
	};
}
