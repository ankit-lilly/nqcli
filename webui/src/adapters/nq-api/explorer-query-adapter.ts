import type { GraphQueryPort, OperationOptions } from "@/application";
import type { Explorer, ExplorerRequestOptions } from "@/connector";
import type {
	CountVerticesCommand,
	EdgeDetailsCommand,
	ExecuteQueryCommand,
	FindNeighborsCommand,
	NeighborCountsCommand,
	SearchVerticesCommand,
	VertexDetailsCommand,
} from "@/domain";

function requestOptions(
	options?: OperationOptions,
): [] | [ExplorerRequestOptions] {
	if (!options?.signal && !options?.requestId) {
		return [];
	}
	return [
		{
			signal: options.signal,
			queryId: options.requestId,
		},
	];
}

/** Compatibility adapter around the existing Gremlin/openCypher connectors. */
export class ExplorerQueryAdapter implements GraphQueryPort {
	constructor(private readonly explorer: Explorer) {}

	execute(command: ExecuteQueryCommand, options?: OperationOptions) {
		return this.explorer.rawQuery(command, ...requestOptions(options));
	}

	countVertices(command: CountVerticesCommand, options?: OperationOptions) {
		return this.explorer.fetchVertexCountsByType(
			command,
			...requestOptions(options),
		);
	}

	neighbors(command: FindNeighborsCommand, options?: OperationOptions) {
		return this.explorer.fetchNeighbors(command, ...requestOptions(options));
	}

	neighborCounts(command: NeighborCountsCommand, options?: OperationOptions) {
		return this.explorer.neighborCounts(command, ...requestOptions(options));
	}

	search(command: SearchVerticesCommand, options?: OperationOptions) {
		return this.explorer.keywordSearch(command, ...requestOptions(options));
	}

	vertexDetails(command: VertexDetailsCommand, options?: OperationOptions) {
		return this.explorer.vertexDetails(command, ...requestOptions(options));
	}

	edgeDetails(command: EdgeDetailsCommand, options?: OperationOptions) {
		return this.explorer.edgeDetails(command, ...requestOptions(options));
	}
}
