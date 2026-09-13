/* oxlint-disable @typescript-eslint/require-await */
import type { Explorer } from "./useGEFetchTypes";

/**
 * Empty explorer for when there is no connection.
 */
export const emptyExplorer: Explorer = {
	connection: {
		url: "",
		queryEngine: "gremlin",
	},
	fetchSchema: async () => {
		return {
			totalVertices: 0,
			vertices: [],
			totalEdges: 0,
			edges: [],
			edgeConnections: [],
		};
	},
	fetchVertexCountsByType: async () => {
		return {
			total: 0,
		};
	},
	fetchNeighbors: async () => ({ vertices: [], edges: [] }),
	neighborCounts: async () => ({ counts: [] }),
	keywordSearch: async () => ({ vertices: [] }),
	vertexDetails: async () => ({ vertices: [] }),
	edgeDetails: async () => ({ edges: [] }),
	rawQuery: async () => ({ results: [], rawResponse: null }),
};
