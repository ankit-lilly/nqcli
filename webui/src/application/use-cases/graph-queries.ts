import type {
	CountVerticesCommand,
	EdgeDetailsCommand,
	ExecuteQueryCommand,
	FindNeighborsCommand,
	NeighborCountsCommand,
	SearchVerticesCommand,
	VertexDetailsCommand,
} from "@/domain";

import type { GraphQueryPort, OperationOptions } from "../ports";

/** Framework-independent graph exploration use cases. */
export class GraphQueries {
	constructor(private readonly queries: GraphQueryPort) {}

	execute(command: ExecuteQueryCommand, options?: OperationOptions) {
		return this.queries.execute(command, options);
	}

	countVertices(command: CountVerticesCommand, options?: OperationOptions) {
		return this.queries.countVertices(command, options);
	}

	neighbors(command: FindNeighborsCommand, options?: OperationOptions) {
		return this.queries.neighbors(command, options);
	}

	neighborCounts(command: NeighborCountsCommand, options?: OperationOptions) {
		return this.queries.neighborCounts(command, options);
	}

	search(command: SearchVerticesCommand, options?: OperationOptions) {
		return this.queries.search(command, options);
	}

	vertexDetails(command: VertexDetailsCommand, options?: OperationOptions) {
		return this.queries.vertexDetails(command, options);
	}

	edgeDetails(command: EdgeDetailsCommand, options?: OperationOptions) {
		return this.queries.edgeDetails(command, options);
	}
}
