import type { GraphElement, GraphSelection } from "./graph";

export type SchemaStatus = "empty" | "running" | "ready" | "failed";

export interface SchemaAttribute {
	name: string;
	dataType: string;
}

export interface SchemaVertex {
	type: string;
	attributes: SchemaAttribute[];
	total?: number;
}

export interface SchemaEdge {
	type: string;
	attributes: SchemaAttribute[];
	total?: number;
}

export interface SchemaConnection {
	sourceVertexType: string;
	edgeType: string;
	targetVertexType: string;
}

export interface SchemaSnapshot {
	status: SchemaStatus;
	phase?: string;
	completed?: number;
	total?: number;
	error?: string;
	lastUpdate?: string;
	totalVertices?: number;
	vertices: SchemaVertex[];
	totalEdges?: number;
	edges: SchemaEdge[];
	edgeConnections: SchemaConnection[];
}

export type SchemaSelectionDetails =
	| { kind: "vertex"; vertex: SchemaVertex }
	| { kind: "connection"; connection: SchemaConnection; edge?: SchemaEdge };

export function schemaConnectionID(connection: SchemaConnection): string {
	return `${connection.sourceVertexType}\u0000${connection.edgeType}\u0000${connection.targetVertexType}`;
}

export function schemaGraphElements(snapshot: SchemaSnapshot): GraphElement[] {
	const vertices = new Set(snapshot.vertices.map((vertex) => vertex.type));
	const nodes: GraphElement[] = snapshot.vertices.map((vertex) => ({
		group: "nodes",
		data: {
			id: vertex.type,
			label: vertex.type,
			name: vertex.type,
			total: vertex.total,
			attributes: vertex.attributes,
		},
	}));
	const edges: GraphElement[] = snapshot.edgeConnections.flatMap(
		(connection) => {
			if (
				!vertices.has(connection.sourceVertexType) ||
				!vertices.has(connection.targetVertexType)
			) {
				return [];
			}
			return [
				{
					group: "edges" as const,
					data: {
						id: schemaConnectionID(connection),
						label: connection.edgeType,
						source: connection.sourceVertexType,
						target: connection.targetVertexType,
					},
				},
			];
		},
	);
	return [...nodes, ...edges];
}

export function schemaSelection(
	snapshot: SchemaSnapshot,
	selection: GraphSelection,
): SchemaSelectionDetails | undefined {
	if (!selection) return undefined;
	if (selection.group === "nodes") {
		const vertex = snapshot.vertices.find(
			(vertex) => vertex.type === selection.data.id,
		);
		return vertex ? { kind: "vertex", vertex } : undefined;
	}
	const connection = snapshot.edgeConnections.find(
		(connection) =>
			connection.sourceVertexType === selection.data.source &&
			connection.targetVertexType === selection.data.target &&
			connection.edgeType === selection.data.label,
	);
	return connection
		? {
				kind: "connection",
				connection,
				edge: snapshot.edges.find((edge) => edge.type === connection.edgeType),
			}
		: undefined;
}

export function formatCount(value?: number): string {
	if (value == null) return "Unknown";
	return new Intl.NumberFormat(undefined, { notation: "compact" }).format(
		value,
	);
}
