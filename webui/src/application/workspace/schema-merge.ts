import type {
	Edge,
	EdgeConnection,
	GraphEntities,
	ScalarValue,
	SchemaAttribute,
	SchemaEdge,
	SchemaVertex,
	Vertex,
	VertexId,
	VertexType,
} from "@/domain";

export type SchemaModel = {
	vertices: SchemaVertex[];
	edges: SchemaEdge[];
	edgeConnections?: EdgeConnection[];
	lastUpdate?: Date;
	totalVertices?: number;
	totalEdges?: number;
};

export type VertexTypeLookup = {
	get(id: VertexId): VertexType[] | undefined;
	isEmpty(): boolean;
};

export function createVertexTypeLookup(
	...sources: ReadonlyMap<VertexId, { types: VertexType[] }>[]
): VertexTypeLookup {
	return {
		get(id) {
			for (const source of sources) {
				const vertex = source.get(id);
				if (vertex) {
					return vertex.types;
				}
			}
			return undefined;
		},
		isEmpty() {
			return sources.every((source) => source.size === 0);
		},
	};
}

/** Merges newly observed entities into a schema while preserving stable references. */
export function updateSchemaFromEntities<Schema extends SchemaModel>(
	entities: Partial<GraphEntities>,
	schema: Schema,
	vertexLookup: VertexTypeLookup,
): Schema {
	const vertices = entities.vertices ?? [];
	const edges = entities.edges ?? [];
	if (vertices.length === 0 && edges.length === 0) {
		return schema;
	}

	const mergedVertices = mergeVertices(schema.vertices, vertices);
	const mergedEdges = mergeEdges(schema.edges, edges);
	const existingConnections = schema.edgeConnections ?? [];
	const mergedConnections = mergeEdgeConnections(
		existingConnections,
		edges,
		vertexLookup,
	);

	if (
		mergedVertices === schema.vertices &&
		mergedEdges === schema.edges &&
		mergedConnections === existingConnections
	) {
		return schema;
	}

	return {
		...schema,
		vertices: mergedVertices,
		edges: mergedEdges,
		edgeConnections: mergedConnections,
	} as Schema;
}

function edgeConnectionKey(connection: EdgeConnection): string {
	return `${connection.sourceVertexType}-[${connection.edgeType}]->${connection.targetVertexType}`;
}

function mergeEdgeConnections(
	existing: EdgeConnection[],
	edges: Edge[],
	vertexLookup: VertexTypeLookup,
): EdgeConnection[] {
	if (edges.length === 0 || vertexLookup.isEmpty()) {
		return existing;
	}

	const existingIds = new Set(existing.map(edgeConnectionKey));
	const newConnections: EdgeConnection[] = [];

	for (const edge of edges) {
		const sourceTypes = vertexLookup.get(edge.sourceId);
		const targetTypes = vertexLookup.get(edge.targetId);
		if (!sourceTypes || !targetTypes) {
			continue;
		}

		for (const sourceVertexType of sourceTypes) {
			for (const targetVertexType of targetTypes) {
				const connection: EdgeConnection = {
					sourceVertexType,
					edgeType: edge.type,
					targetVertexType,
				};
				const id = edgeConnectionKey(connection);
				if (!existingIds.has(id)) {
					existingIds.add(id);
					newConnections.push(connection);
				}
			}
		}
	}

	return newConnections.length === 0
		? existing
		: [...existing, ...newConnections];
}

function mergeVertices<VertexSchema extends SchemaVertex>(
	existing: VertexSchema[],
	vertices: Vertex[],
): VertexSchema[] {
	if (vertices.length === 0) {
		return existing;
	}

	const byType = new Map(existing.map((vertex) => [vertex.type, vertex]));
	let hasChanges = false;

	for (const vertex of vertices) {
		const attributes = attributesFromProperties(vertex.attributes);
		for (const type of vertex.types) {
			const existingConfig = byType.get(type);
			if (!existingConfig) {
				byType.set(type, { type, attributes } as VertexSchema);
				hasChanges = true;
				continue;
			}

			const mergedAttributes = mergeAttributesFromProperties(
				existingConfig.attributes,
				vertex.attributes,
			);
			if (mergedAttributes !== existingConfig.attributes) {
				byType.set(type, {
					...existingConfig,
					attributes: mergedAttributes,
				});
				hasChanges = true;
			}
		}
	}

	return hasChanges ? Array.from(byType.values()) : existing;
}

function mergeEdges<EdgeSchema extends SchemaEdge>(
	existing: EdgeSchema[],
	edges: Edge[],
): EdgeSchema[] {
	if (edges.length === 0) {
		return existing;
	}

	const byType = new Map(existing.map((edge) => [edge.type, edge]));
	let hasChanges = false;

	for (const edge of edges) {
		const existingConfig = byType.get(edge.type);
		if (!existingConfig) {
			byType.set(edge.type, {
				type: edge.type,
				attributes: attributesFromProperties(edge.attributes),
			} as EdgeSchema);
			hasChanges = true;
			continue;
		}

		const mergedAttributes = mergeAttributesFromProperties(
			existingConfig.attributes,
			edge.attributes,
		);
		if (mergedAttributes !== existingConfig.attributes) {
			byType.set(edge.type, {
				...existingConfig,
				attributes: mergedAttributes,
			});
			hasChanges = true;
		}
	}

	return hasChanges ? Array.from(byType.values()) : existing;
}

function mergeAttributesFromProperties(
	existing: SchemaAttribute[],
	properties: Record<string, ScalarValue>,
): SchemaAttribute[] {
	const existingNames = new Set(existing.map((attribute) => attribute.name));
	const newAttributes = Object.keys(properties)
		.filter((name) => !existingNames.has(name))
		.map((name) => ({ name, dataType: detectDataType(properties[name]) }));

	return newAttributes.length === 0
		? existing
		: [...existing, ...newAttributes];
}

function attributesFromProperties(
	properties: Record<string, ScalarValue>,
): SchemaAttribute[] {
	return Object.entries(properties).map(([name, value]) => ({
		name,
		dataType: detectDataType(value),
	}));
}

export function mapVertexToTypeConfigs(vertex: Vertex): SchemaVertex[] {
	return vertex.types.map((type) => ({
		type,
		attributes: attributesFromProperties(vertex.attributes),
	}));
}

export function mapEdgeToTypeConfig(edge: Edge): SchemaEdge {
	return {
		type: edge.type,
		attributes: attributesFromProperties(edge.attributes),
	};
}

function detectDataType(value: ScalarValue): string | undefined {
	if (value === null) return undefined;
	if (value instanceof Date) return "Date";
	switch (typeof value) {
		case "string":
			return "String";
		case "number":
			return "Number";
		case "boolean":
			return "Boolean";
	}
}
