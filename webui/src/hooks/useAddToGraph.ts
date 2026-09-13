import { useMutation } from "@tanstack/react-query";
import { useStore } from "jotai";
import { useCallback, useMemo } from "react";
import { toast } from "sonner";

import { AddEntitiesToWorkspace } from "@/application";
import { JotaiGraphWorkspaceAdapter } from "@/adapters/state";
import {
	type Edge,
	type Entities,
	useUpdateGraphSession,
	type Vertex,
} from "@/core";
import { logger } from "@/utils";
import { createDisplayError } from "@/utils/createDisplayError";

/** Returns a callback that adds an array of nodes and edges to the graph. */
export function useAddToGraph() {
	const store = useStore();
	const updateGraphStorage = useUpdateGraphSession();
	const addEntities = useMemo(
		() =>
			new AddEntitiesToWorkspace(
				new JotaiGraphWorkspaceAdapter(store, updateGraphStorage),
			),
		[store, updateGraphStorage],
	);

	return useCallback(
		(entities: Partial<Entities>) => addEntities.execute(entities),
		[addEntities],
	);
}

/** Returns a callback that adds the given vertex to the graph. */
export function useAddVertexToGraph(vertex: Vertex) {
	const callback = useAddToGraph();
	return () => callback({ vertices: [vertex] });
}

/** Returns a callback that adds the given edge to the graph. */
export function useAddEdgeToGraph(edge: Edge) {
	const callback = useAddToGraph();
	return () => callback({ edges: [edge] });
}

/**
 * Wraps sendToGraph in a mutation which allows monitoring progress and error state.
 *
 * On error, a toast notification will be shown.
 */
export function useAddToGraphMutation() {
	const sendToGraph = useAddToGraph();
	return useMutation({
		mutationFn: sendToGraph,
		onError: (error) => {
			const displayError = createDisplayError(error);
			toast.error(displayError.message);
			logger.error("Failed to add all to graph", error);
		},
	});
}
