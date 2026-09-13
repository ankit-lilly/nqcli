import type {
	Edge,
	EdgeId,
	GraphEntities,
	Vertex,
	VertexId,
	VertexType,
} from "../graph";

export type AttributeFilter = {
	name: string;
	value: string;
};

export type FindNeighborsCommand = {
	vertexId: VertexId;
	excludedVertices?: Set<VertexId>;
	filterByVertexTypes?: string[];
	attributeFilters?: AttributeFilter[];
	limit?: number;
};

export type NeighborCountsCommand = {
	vertexIds: VertexId[];
};

export type CountVerticesCommand = { label: string };
export type CountVerticesResult = { total: number };

export type NeighborCount = {
	vertexId: VertexId;
	totalCount: number;
	counts: Map<VertexType, number>;
};

export type SearchVerticesCommand = {
	searchTerm?: string;
	searchByAttributes?: string[];
	vertexTypes?: string[];
	limit?: number;
	offset?: number;
	exactMatch?: boolean;
};

export type ExecuteQueryCommand = { query: string };

export type VertexDetailsResult = { vertices: Vertex[] };
export type EdgeDetailsResult = { edges: Edge[] };
export type EdgeDetailsCommand = { edgeIds: EdgeId[] };
export type VertexDetailsCommand = { vertexIds: VertexId[] };
export type NeighborsResult = GraphEntities;
export type NeighborCountsResult = { counts: NeighborCount[] };
export type SearchVerticesResult = { vertices: Vertex[] };
