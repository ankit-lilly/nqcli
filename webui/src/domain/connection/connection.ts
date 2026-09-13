export const queryEngineOptions = ["gremlin", "openCypher"] as const;

export type QueryEngine = (typeof queryEngineOptions)[number];

/** Browser-facing connection to the nq HTTP API. */
export type ConnectionConfig = {
	url: string;
	queryEngine?: QueryEngine;
	fetchTimeoutMs?: number;
	nodeExpansionLimit?: number;
};

export type NormalizedConnection = ConnectionConfig & {
	queryEngine: QueryEngine;
};

export type DefaultConnection = {
	apiBaseUrl: string;
	queryEngine: QueryEngine;
	profile: string;
};

export function normalizeUrl(url: string | undefined): string {
	return (
		url
			?.replace(/[\r\n]/g, "")
			.trim()
			.replace(/\/$/, "") ?? ""
	);
}

export function normalizeConnection(
	connection: ConnectionConfig,
): NormalizedConnection {
	return {
		...connection,
		url: normalizeUrl(connection.url),
		queryEngine: connection.queryEngine ?? "gremlin",
	};
}
