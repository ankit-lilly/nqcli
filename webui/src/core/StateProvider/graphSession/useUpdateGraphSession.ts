import { useAtomCallback } from "jotai/utils";
import { useCallback } from "react";

import { logger } from "@/utils";

import { edgesAtom } from "../edges";
import { nodesAtom } from "../nodes";
import {
	activeGraphSessionAtom,
	type GraphSessionStorageModel,
	isRestorePreviousSessionAvailableAtom,
} from "./storage";

/**
 * Returns a callback that can be used to trigger an update of the graph
 * session storage for the active connection.
 */
export function useUpdateGraphSession() {
	return useAtomCallback(
		useCallback((get, set) => {
			// Get the latest graph data from the atoms
			const nodesInGraph = get(nodesAtom);
			const edgesInGraph = get(edgesAtom);

			const vertices = new Set(nodesInGraph.keys());
			const edges = new Set(edgesInGraph.keys());

			// Construct the graph storage model
			const graphSession: GraphSessionStorageModel = {
				vertices,
				edges,
			};

			// Update the session
			logger.debug("Updating graph session", graphSession);
			set(activeGraphSessionAtom, graphSession);
			set(isRestorePreviousSessionAvailableAtom, false);
		}, []),
	);
}
