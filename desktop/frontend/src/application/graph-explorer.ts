import type { GraphElement } from "../domain/graph";

export type QueryType = "gremlin" | "cypher";
export type ExpansionDirection = "both" | "in" | "out";

export type RunGraphQuery = {
	query: string;
	type: QueryType;
};

export type ExpandVertexCommand = {
	id: string;
	type: QueryType;
	direction?: ExpansionDirection;
	relationship?: string;
	neighborLabel?: string;
	limit?: number;
	excludedVertexIds?: string[];
};

export type GraphResult = {
	elements: GraphElement[];
	warning?: string;
};

export interface GraphQueryPort {
	run(command: RunGraphQuery): Promise<GraphResult>;
	expand(command: ExpandVertexCommand): Promise<GraphResult>;
	vertexProperties(id: string): Promise<Record<string, unknown> | null>;
}

/** Framework-independent graph exploration use cases. */
export class GraphExplorer {
	constructor(private readonly queries: GraphQueryPort) {}

	run(command: RunGraphQuery) {
		return this.queries.run(command);
	}

	expand(command: ExpandVertexCommand) {
		return this.queries.expand(command);
	}

	vertexProperties(id: string) {
		return this.queries.vertexProperties(id);
	}
}
