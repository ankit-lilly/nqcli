import {
	createVertexId,
	type EntityProperties,
	type EntityRawId,
	type VertexType,
} from "@/core";
import type { PatchedResultVertex, ResultVertex } from "@/domain";

export type { PatchedResultVertex, ResultVertex } from "@/domain";

/**
 * A vertex result from a graph database query.
 *
 * If the attributes are not defined, the vertex is assumed to be a fragment and
 * will have their details fetched from the database.
 */
/**
 * Creates a ResultVertex instance from the given options.
 */
export function createResultVertex(options: {
	id: EntityRawId;
	name?: string;
	types?: string[];
	attributes?: EntityProperties;
}): ResultVertex {
	return {
		entityType: "vertex",
		id: createVertexId(options.id),
		...(options.name ? { name: options.name } : {}),
		types: (options.types as VertexType[]) ?? [],
		...(options.attributes ? { attributes: options.attributes } : {}),
	};
}

/**
 * Creates a PatchedResultVertex instance from the given options.
 */
export function createPatchedResultVertex(options: {
	id: EntityRawId;
	types: string[];
	attributes: EntityProperties;
	/**
	 * The name of the vertex provided in the query result set. This is mainly just
	 * useful for user queries.
	 */
	name?: string;
}): PatchedResultVertex {
	return {
		entityType: "patched-vertex",
		id: createVertexId(options.id),
		types: options.types as VertexType[],
		attributes: options.attributes,
		...(options.name ? { name: options.name } : {}),
	};
}
