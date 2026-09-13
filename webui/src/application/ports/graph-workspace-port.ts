import type { GraphEntities, Vertex, VertexId } from "@/domain";

import type { VertexTypeLookup } from "../workspace";

export interface GraphWorkspacePort {
	vertices(): ReadonlyMap<VertexId, Vertex>;
	add(entities: Partial<GraphEntities>): void;
	mergeSchema(
		entities: Partial<GraphEntities>,
		vertexLookup: VertexTypeLookup,
	): void;
	persist(): void | Promise<void>;
}
