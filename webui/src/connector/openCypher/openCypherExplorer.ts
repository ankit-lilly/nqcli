import type { NormalizedConnection } from "@/domain";
import { logger } from "@/utils";

import type {
	Explorer,
	ExplorerRequestOptions,
	SchemaResponse,
} from "../useGEFetchTypes";

import { fetchDatabaseRequest } from "../fetchDatabaseRequest";
import { ServerLoggerConnector } from "../LoggerConnector";
import { edgeDetails } from "./edgeDetails";
import fetchNeighbors from "./fetchNeighbors";
import fetchVertexTypeCounts from "./fetchVertexTypeCounts";
import keywordSearch from "./keywordSearch";
import { neighborCounts } from "./neighborCounts";
import { rawQuery } from "./rawQuery";
import { vertexDetails } from "./vertexDetails";

function _openCypherFetch(
	connection: NormalizedConnection,
	options?: ExplorerRequestOptions,
) {
	return async (queryTemplate: string) => {
		logger.debug(queryTemplate);
		return fetchDatabaseRequest(connection, `${connection.url}/openCypher`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({ query: queryTemplate }),
			...options,
		});
	};
}

export function createOpenCypherExplorer(
	connection: NormalizedConnection,
): Explorer {
	const remoteLogger = new ServerLoggerConnector(connection.url);
	return {
		connection,
		async fetchSchema(options) {
			remoteLogger.info("[openCypher Explorer] Fetching schema...");
			return fetchBackendSchema(connection, options);
		},
		async fetchVertexCountsByType(req, options) {
			remoteLogger.info(
				"[openCypher Explorer] Fetching vertex counts by type...",
			);
			return fetchVertexTypeCounts(_openCypherFetch(connection, options), req);
		},
		async fetchNeighbors(req, options) {
			remoteLogger.info("[openCypher Explorer] Fetching neighbors...");
			return fetchNeighbors(_openCypherFetch(connection, options), req);
		},
		async neighborCounts(req, options) {
			remoteLogger.info("[openCypher Explorer] Fetching neighbors count...");
			return neighborCounts(_openCypherFetch(connection, options), req);
		},
		async keywordSearch(req, options) {
			remoteLogger.info("[openCypher Explorer] Fetching keyword search...");
			return keywordSearch(_openCypherFetch(connection, options), req);
		},
		async vertexDetails(req, options) {
			remoteLogger.info("[openCypher Explorer] Fetching vertex details...");
			return vertexDetails(_openCypherFetch(connection, options), req);
		},
		async edgeDetails(req, options) {
			remoteLogger.info("[openCypher Explorer] Fetching edge details...");
			return edgeDetails(_openCypherFetch(connection, options), req);
		},
		async rawQuery(req, options) {
			remoteLogger.info("[openCypher Explorer] Fetching raw query...");
			return rawQuery(_openCypherFetch(connection, options), req);
		},
	} satisfies Explorer;
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
