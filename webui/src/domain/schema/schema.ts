import type { EdgeType, VertexType } from "../graph";

export type SchemaStatus = "empty" | "running" | "ready" | "failed";

export type SchemaAttribute = {
	name: string;
	dataType?: string;
};

export type SchemaVertex = {
	type: VertexType;
	attributes: SchemaAttribute[];
	displayNameAttribute?: string;
	longDisplayNameAttribute?: string;
	total?: number;
};

export type SchemaEdge = {
	type: EdgeType;
	attributes: SchemaAttribute[];
	total?: number;
};

export type EdgeConnection = {
	edgeType: EdgeType;
	sourceVertexType: VertexType;
	targetVertexType: VertexType;
	count?: number;
};

export type SchemaSnapshot = {
	status?: SchemaStatus;
	lastUpdate?: string;
	totalVertices?: number;
	vertices: SchemaVertex[];
	totalEdges?: number;
	edges: SchemaEdge[];
	edgeConnections: EdgeConnection[];
};

export type SchemaEvent = {
	type?: string;
	status?: SchemaStatus;
	phase?: string;
	completed?: number;
	total?: number;
	error?: string;
};
