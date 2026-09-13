import { atom, useAtomValue } from "jotai";
import { selectAtom } from "jotai/utils";

import type { Explorer } from "@/connector/useGEFetchTypes";

import { emptyExplorer } from "@/connector/emptyExplorer";
import { createGremlinExplorer } from "@/connector/gremlin/gremlinExplorer";
import {
	ClientLoggerConnector,
	type LoggerConnector,
	ServerLoggerConnector,
} from "@/connector/LoggerConnector";
import { createOpenCypherExplorer } from "@/connector/openCypher/openCypherExplorer";
import { logger } from "@/utils";

import {
	activeConnectionAtom,
	type NormalizedConnection,
} from "./StateProvider/configuration";

export const explorerAtom = atom((get) => {
	const explorerForTesting = get(explorerForTestingAtom);
	if (explorerForTesting) {
		return explorerForTesting;
	}
	const connection = get(activeConnectionAtom);
	if (!connection) {
		return emptyExplorer;
	}
	logger.debug("Creating explorer for connection:", connection);
	switch (connection.queryEngine) {
		case "openCypher":
			return createOpenCypherExplorer(connection);
		case "gremlin":
			return createGremlinExplorer(connection);
	}
});

/**
 * Explorer based on the active connection.
 */
export function useExplorer() {
	return useAtomValue(explorerAtom);
}

/** CAUTION: This atom is only for testing purposes. */
export const explorerForTestingAtom = atom<Explorer | null>(null);

export const queryEngineSelector = atom((get) =>
	get(
		selectAtom(activeConnectionAtom, (c) =>
			c && c.queryEngine ? c.queryEngine : "gremlin",
		),
	),
);

export function useQueryEngine() {
	return useAtomValue(queryEngineSelector);
}

/**
 * Logger based on the active connection proxy URL.
 */
export const loggerSelector = atom((get) =>
	createLoggerFromConnection(get(activeConnectionAtom)),
);

/** Creates a logger instance that will be remote if the connection is using the
 * proxy server. Otherwise it will be a client only logger. */
export function createLoggerFromConnection(
	connection: NormalizedConnection | null,
): LoggerConnector {
	// Check for a url and that we are using the proxy server
	if (!connection?.url) {
		logger.debug(
			"Connection did not contain enough information to create a remote logger, so using a client logger instead",
			connection,
		);
		return new ClientLoggerConnector();
	}

	logger.debug(
		"Creating a remote server logger using proxy server URL",
		connection.url,
	);
	return new ServerLoggerConnector(connection.url);
}
