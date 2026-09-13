import { createDefaultConnectionApplication } from "@/bootstrap";
import type { DefaultConnection } from "@/domain";
import type {
	ConfigurationId,
	RawConfiguration,
} from "./ConfigurationProvider";

export type DefaultConnectionData = DefaultConnection;

/** Loads the single connection exposed by the nq server. */
export async function fetchDefaultConnection(): Promise<RawConfiguration> {
	const connections = createDefaultConnectionApplication();
	return mapToConnection(await connections.getDefault());
}

export function mapToConnection(data: DefaultConnectionData): RawConfiguration {
	const profile = data.profile || "default";
	return {
		id: `profile:${profile}:${data.queryEngine}` as ConfigurationId,
		displayLabel: profile,
		connection: {
			url: data.apiBaseUrl,
			queryEngine: data.queryEngine,
		},
	};
}
