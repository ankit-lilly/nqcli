import { queryOptions } from "@tanstack/react-query";

import type { VertexId } from "@/core";

import {
	getConnectionApplication,
	getStore,
	updateVertexGraphCanvasState,
} from "./helpers";

export function vertexDetailsQuery(vertexId: VertexId) {
	return queryOptions({
		queryKey: ["vertex", vertexId],
		queryFn: async ({ signal, meta }) => {
			const application = getConnectionApplication(meta);
			const store = getStore(meta);

			const results = await application.queries.vertexDetails(
				{ vertexIds: [vertexId] },
				{ signal },
			);

			const vertex = results.vertices[0] ?? null;

			if (vertex) {
				updateVertexGraphCanvasState(store, [vertex]);
			}

			return { vertex };
		},
	});
}
