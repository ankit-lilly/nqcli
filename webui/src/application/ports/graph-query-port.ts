import type {
	CountVerticesCommand,
	CountVerticesResult,
	EdgeDetailsCommand,
	EdgeDetailsResult,
	ExecuteQueryCommand,
	FindNeighborsCommand,
	NeighborCountsCommand,
	NeighborCountsResult,
	NeighborsResult,
	QueryResult,
	SearchVerticesCommand,
	SearchVerticesResult,
	VertexDetailsCommand,
	VertexDetailsResult,
} from "@/domain";

import type { OperationOptions } from "./operation";

/** Database operations needed by graph exploration, independent of transport. */
export interface GraphQueryPort {
	execute(
		command: ExecuteQueryCommand,
		options?: OperationOptions,
	): Promise<QueryResult>;
	countVertices(
		command: CountVerticesCommand,
		options?: OperationOptions,
	): Promise<CountVerticesResult>;
	neighbors(
		command: FindNeighborsCommand,
		options?: OperationOptions,
	): Promise<NeighborsResult>;
	neighborCounts(
		command: NeighborCountsCommand,
		options?: OperationOptions,
	): Promise<NeighborCountsResult>;
	search(
		command: SearchVerticesCommand,
		options?: OperationOptions,
	): Promise<SearchVerticesResult>;
	vertexDetails(
		command: VertexDetailsCommand,
		options?: OperationOptions,
	): Promise<VertexDetailsResult>;
	edgeDetails(
		command: EdgeDetailsCommand,
		options?: OperationOptions,
	): Promise<EdgeDetailsResult>;
}
