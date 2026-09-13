import type { Branded } from "../shared/brand";

export type ScalarValue = string | number | boolean | Date | null;
export type EntityRawId = string | number;
export type EntityPropertyValue = ScalarValue;
export type EntityProperties = Record<string, EntityPropertyValue>;

export type VertexId = Branded<EntityRawId, "VertexId">;
export type VertexType = Branded<string, "VertexType">;
export type EdgeId = Branded<EntityRawId, "EdgeId">;
export type EdgeType = Branded<string, "EdgeType">;

export type Vertex = {
	id: VertexId;
	type: VertexType;
	types: VertexType[];
	attributes: EntityProperties;
};

export type Edge = {
	id: EdgeId;
	type: EdgeType;
	sourceId: VertexId;
	targetId: VertexId;
	attributes: EntityProperties;
};

export type GraphEntities = {
	vertices: Vertex[];
	edges: Edge[];
};

export const MISSING_VERTEX_TYPE = "\u00abNo Type\u00bb";

export function createVertexId(id: EntityRawId): VertexId {
	return id as VertexId;
}

export function createEdgeId(id: EntityRawId): EdgeId {
	return id as EdgeId;
}

export function getRawId(id: VertexId | EdgeId): EntityRawId {
	return id;
}

export function createVertexType(type: string): VertexType {
	return type as VertexType;
}

export function createEdgeType(type: string): EdgeType {
	return type as EdgeType;
}

export function createVertex(options: {
	id: EntityRawId;
	types?: string[];
	attributes?: EntityProperties;
}): Vertex {
	const givenTypes = options.types ?? [];
	const types = (
		givenTypes.length > 0 ? givenTypes : [MISSING_VERTEX_TYPE]
	) as VertexType[];

	return {
		id: createVertexId(options.id),
		type: types[0],
		types,
		attributes: options.attributes ?? {},
	};
}

export function createEdge(options: {
	id: EntityRawId;
	type: string;
	sourceId: EntityRawId;
	targetId: EntityRawId;
	attributes?: EntityProperties;
}): Edge {
	return {
		id: createEdgeId(options.id),
		type: createEdgeType(options.type),
		sourceId: createVertexId(options.sourceId),
		targetId: createVertexId(options.targetId),
		attributes: options.attributes ?? {},
	};
}
