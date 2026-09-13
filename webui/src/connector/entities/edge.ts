import {
	createEdgeId,
	createEdgeType,
	createVertexId,
	type EntityProperties,
	type EntityRawId,
	type Vertex,
} from "@/core";
import type { PatchedResultEdge, ResultEdge } from "@/domain";

export type { PatchedResultEdge, ResultEdge } from "@/domain";

import { createPatchedResultVertex, type PatchedResultVertex } from "./vertex";

/**
 * An edge result from a graph database query.
 *
 * If the attributes are undefined, the edge is assumed to be a fragment and
 * will have their details fetched from the database.
 */
/**
 * Creates a ResultEdge instance from the given options.
 */
export function createResultEdge(options: {
	id: EntityRawId;
	sourceId: EntityRawId;
	targetId: EntityRawId;
	type: string;

	/**
	 * If no attributes are provided, then a future process will fetch edge
	 * details to get the attributes.
	 */
	attributes?: EntityProperties;

	/**
	 * The name of the edge provided in the query result set. This is mainly just
	 * useful for user queries.
	 */
	name?: string;
}): ResultEdge {
	return {
		entityType: "edge",
		id: createEdgeId(options.id),
		sourceId: createVertexId(options.sourceId),
		targetId: createVertexId(options.targetId),
		type: createEdgeType(options.type),
		...(options.attributes ? { attributes: options.attributes } : {}),
		...(options.name ? { name: options.name } : {}),
	};
}

/**
 * Creates a PatchedResultEdge instance from the given options.
 */
export function createPatchedResultEdge(options: {
	id: EntityRawId;
	sourceVertex: Vertex;
	targetVertex: Vertex;
	type: string;
	attributes: EntityProperties;

	/**
	 * The name of the edge provided in the query result set. This is mainly just
	 * useful for user queries.
	 */
	name?: string;
}): PatchedResultEdge {
	return {
		entityType: "patched-edge",
		id: createEdgeId(options.id),
		type: createEdgeType(options.type),
		source: createPatchedResultVertex({
			...options.sourceVertex,
			name: "source",
		}),
		target: createPatchedResultVertex({
			...options.targetVertex,
			name: "target",
		}),
		attributes: options.attributes ?? {},
		...(options.name ? { name: options.name } : {}),
	};
}
