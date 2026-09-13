import { SseSchemaEventsAdapter } from "@/adapters/events";
import {
	ExplorerQueryAdapter,
	NqDefaultConnectionAdapter,
	NqProfileAdapter,
	NqSchemaAdapter,
} from "@/adapters/nq-api";
import {
	ConnectionService,
	GraphQueries,
	ProfileService,
	SchemaService,
} from "@/application";
import type { Explorer } from "@/connector";
import { createGremlinExplorer } from "@/connector/gremlin/gremlinExplorer";
import { createOpenCypherExplorer } from "@/connector/openCypher/openCypherExplorer";
import type { NormalizedConnection } from "@/domain";

export function createConnectionApplication(
	explorer: Explorer,
	onSchemaEventError?: (error: unknown) => void,
) {
	const events =
		typeof EventSource === "undefined"
			? undefined
			: new SseSchemaEventsAdapter(
					explorer.connection.url,
					undefined,
					onSchemaEventError,
				);

	return {
		queries: new GraphQueries(new ExplorerQueryAdapter(explorer)),
		schema: new SchemaService(new NqSchemaAdapter(explorer), events),
	};
}

/** Composition root for any presentation framework given an nq connection. */
export function createApplicationForConnection(
	connection: NormalizedConnection,
	onSchemaEventError?: (error: unknown) => void,
) {
	const explorer =
		connection.queryEngine === "openCypher"
			? createOpenCypherExplorer(connection)
			: createGremlinExplorer(connection);
	return createConnectionApplication(explorer, onSchemaEventError);
}

export function createProfileApplication(baseUrl = "") {
	return new ProfileService(new NqProfileAdapter(baseUrl));
}

export function createDefaultConnectionApplication(baseUrl?: string) {
	return new ConnectionService(new NqDefaultConnectionAdapter(baseUrl));
}
