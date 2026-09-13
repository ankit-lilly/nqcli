import { v4 } from "uuid";

import type { NormalizedConnection } from "@/domain";
import { logger } from "@/utils";

import type {
	Explorer,
	ExplorerRequestOptions,
	SchemaResponse,
} from "../useGEFetchTypes";
import type { GremlinFetch } from "./types";

import { fetchDatabaseRequest } from "../fetchDatabaseRequest";
import { ServerLoggerConnector } from "../LoggerConnector";
import { edgeDetails } from "./edgeDetails";
import fetchNeighbors from "./fetchNeighbors";
import fetchVertexTypeCounts from "./fetchVertexTypeCounts";
import keywordSearch from "./keywordSearch";
import { neighborCounts } from "./neighborCounts";
import { rawQuery } from "./rawQuery";
import { vertexDetails } from "./vertexDetails";

function _gremlinFetch(
	connection: NormalizedConnection,
	options?: ExplorerRequestOptions,
): GremlinFetch {
	return async (queryTemplate: string) => {
		logger.debug(queryTemplate);
		const body = JSON.stringify({ query: queryTemplate });
		const headers: HeadersInit = {
			"Content-Type": "application/json",
			Accept: "application/vnd.gremlin-v3.0+json",
		};
		if (options?.queryId) {
			headers.queryId = options.queryId;
		}

		return fetchDatabaseRequest(connection, `${connection.url}/gremlin`, {
			method: "POST",
			headers,
			body,
			...options,
		});
	};
}

async function fetchBackendSchema(
	connection: NormalizedConnection,
	options?: RequestInit,
): Promise<SchemaResponse> {
	return fetchDatabaseRequest(
		connection,
		`${connection.url}/schema?wait=true`,
		{
			method: "GET",
			...options,
		},
	);
}

export function createGremlinExplorer(
	connection: NormalizedConnection,
): Explorer {
	const remoteLogger = new ServerLoggerConnector(connection.url);
	return {
		connection: connection,
		async fetchSchema(options) {
			remoteLogger.info("[Gremlin Explorer] Fetching schema...");
			return fetchBackendSchema(connection, options);
		},
		async fetchVertexCountsByType(req, options) {
			remoteLogger.info("[Gremlin Explorer] Fetching vertex counts by type...");
			return fetchVertexTypeCounts(_gremlinFetch(connection, options), req);
		},
		async fetchNeighbors(req, options) {
			remoteLogger.info("[Gremlin Explorer] Fetching neighbors...");
			return fetchNeighbors(_gremlinFetch(connection, options), req);
		},
		async neighborCounts(req, options) {
			remoteLogger.info("[Gremlin Explorer] Fetching neighbors count...");
			return neighborCounts(_gremlinFetch(connection, options), req);
		},
		async keywordSearch(req, options) {
			options ??= {};
			options.queryId = v4();

			remoteLogger.info("[Gremlin Explorer] Fetching keyword search...");
			return keywordSearch(_gremlinFetch(connection, options), req);
		},
		async vertexDetails(req, options) {
			options ??= {};
			options.queryId = v4();

			remoteLogger.info("[Gremlin Explorer] Fetching vertex details...");
			const result = await vertexDetails(
				_gremlinFetch(connection, options),
				req,
			);
			return result;
		},
		async edgeDetails(req, options) {
			options ??= {};
			options.queryId = v4();

			remoteLogger.info("[Gremlin Explorer] Fetching edge details...");
			const result = await edgeDetails(_gremlinFetch(connection, options), req);
			return result;
		},
		async rawQuery(req, options) {
			options ??= {};
			options.queryId = v4();
			remoteLogger.info("[Gremlin Explorer] Fetching raw query...");
			const result = await rawQuery(_gremlinFetch(connection, options), req);
			return result;
		},
	} satisfies Explorer;
}
