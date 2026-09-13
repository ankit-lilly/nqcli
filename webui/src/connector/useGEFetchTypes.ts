import type {
	AttributeFilter,
	CountVerticesCommand,
	CountVerticesResult,
	EdgeDetailsCommand,
	EdgeDetailsResult,
	FindNeighborsCommand,
	GraphEntities,
	NeighborCount,
	NeighborCountsCommand,
	NeighborCountsResult,
	NormalizedConnection,
	QueryResult,
	SchemaEdge,
	SchemaSnapshot,
	SchemaVertex,
	SearchVerticesCommand,
	SearchVerticesResult,
	VertexDetailsCommand,
	VertexDetailsResult,
} from "@/domain";

export type VertexSchemaResponse = SchemaVertex;
export type EdgeSchemaResponse = SchemaEdge;
export type SchemaResponse = SchemaSnapshot;
export type CountsByTypeRequest = CountVerticesCommand;
export type CountsByTypeResponse = CountVerticesResult;
export type NeighborsRequest = FindNeighborsCommand;
export type NeighborsResponse = GraphEntities;
export type NeighborCountsRequest = NeighborCountsCommand;
export type NeighborCountsResponse = NeighborCountsResult;
export type KeywordSearchRequest = SearchVerticesCommand;
export type KeywordSearchResponse = SearchVerticesResult;
export type VertexDetailsRequest = VertexDetailsCommand;
export type VertexDetailsResponse = VertexDetailsResult;
export type EdgeDetailsRequest = EdgeDetailsCommand;
export type EdgeDetailsResponse = EdgeDetailsResult;
export type RawQueryRequest = { query: string };
export type RawQueryResponse = QueryResult;

export type { AttributeFilter, NeighborCount };

export type ErrorResponse = {
	code: string;
	detailedMessage: string;
};

/** Transitional HTTP options accepted by the legacy connector adapters. */
export type ExplorerRequestOptions = RequestInit & {
	queryId?: string;
};

/**
 * Compatibility facade for the original Graph Explorer connectors.
 * New application code depends on the focused ports in `src/application`.
 */
export type Explorer = {
	connection: NormalizedConnection;
	fetchSchema: (options?: ExplorerRequestOptions) => Promise<SchemaResponse>;
	fetchVertexCountsByType: (
		req: CountsByTypeRequest,
		options?: ExplorerRequestOptions,
	) => Promise<CountsByTypeResponse>;
	fetchNeighbors: (
		req: NeighborsRequest,
		options?: ExplorerRequestOptions,
	) => Promise<NeighborsResponse>;
	neighborCounts: (
		req: NeighborCountsRequest,
		options?: ExplorerRequestOptions,
	) => Promise<NeighborCountsResponse>;
	keywordSearch: (
		req: KeywordSearchRequest,
		options?: ExplorerRequestOptions,
	) => Promise<KeywordSearchResponse>;
	vertexDetails: (
		req: VertexDetailsRequest,
		options?: ExplorerRequestOptions,
	) => Promise<VertexDetailsResponse>;
	edgeDetails: (
		req: EdgeDetailsRequest,
		options?: ExplorerRequestOptions,
	) => Promise<EdgeDetailsResponse>;
	rawQuery: (
		req: RawQueryRequest,
		options?: ExplorerRequestOptions,
	) => Promise<RawQueryResponse>;
};
