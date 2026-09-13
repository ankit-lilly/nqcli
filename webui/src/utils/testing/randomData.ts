import type { Writable } from "type-fest";

import { type QueryEngine, queryEngineOptions } from "@/core";
import {
	createArray,
	createRandomBoolean,
	createRandomColor,
	createRandomDate,
	createRandomDouble,
	createRandomInteger,
	createRandomName,
	createRandomUrlString,
	createRecord,
	randomlyUndefined,
} from "@/utils/testing/random";

import {
	createPatchedResultEdge,
	createPatchedResultVertex,
	createResultEdge,
	createResultScalar,
	createResultVertex,
} from "@/connector/entities";
import {
	type ArrowStyle,
	type AttributeConfig,
	type ConfigurationId,
	type ConnectionWithId,
	createEdge,
	createEdgeId,
	createEdgeType,
	createNewConfigurationId,
	createVertex,
	createVertexId,
	createVertexType,
	type EdgeConnection,
	type EdgeId,
	type EdgeStyle,
	type EdgeStyleStorage,
	type EdgeType,
	type EdgeTypeConfig,
	type Entities,
	type EntityProperties,
	type EntityRawId,
	type GraphViewLayout,
	type LineStyle,
	type RawConfiguration,
	resolveEdgeStyle,
	resolveVertexStyle,
	type SchemaStorageModel,
	type Vertex,
	type VertexId,
	type VertexStyle,
	type VertexStyleStorage,
	type VertexType,
	type VertexTypeConfig,
} from "@/core";
import {
	graphViewSidebarItems,
	toggleableViews,
} from "@/core/StateProvider/graphViewLayoutDefaults";
import {
	type SchemaViewLayout,
	schemaViewSidebarItems,
} from "@/core/StateProvider/schemaViewLayoutDefaults";
import {
	createExportedGraph,
	type ExportedGraphConnection,
} from "@/modules/GraphViewer/exportedGraph";

/*

# Developer Note

These helper functions are provided to allow for easier test data creation.

When creating test data you should start with a random object, then set the values
that directly apply to the logic you are testing.

The randomness of all the other values ensures that the logic under test is not 
affected by those values, regardless of what they are.

*/

/**
 * Creates a random AttributeConfig object.
 * @returns A random AttributeConfig object.
 */
export function createRandomAttributeConfig(): AttributeConfig {
	const dataType = randomlyUndefined(createRandomName("dataType"));
	const displayLabel = randomlyUndefined(createRandomName("displayLabel"));

	return {
		name: createRandomName("name"),
		...(displayLabel && { displayLabel }),
		...(dataType && { dataType }),
	};
}

/**
 * Creates a random EdgeTypeConfig object.
 * @returns A random EdgeTypeConfig object.
 */
export function createRandomEdgeTypeConfig(): EdgeTypeConfig {
	const displayLabel = randomlyUndefined(createRandomName("displayLabel"));
	const hidden = randomlyUndefined(createRandomBoolean());
	return {
		type: createRandomEdgeType(),
		attributes: createArray(6, createRandomAttributeConfig),
		...(displayLabel && { displayLabel }),
		...(hidden && { hidden }),
		total: createRandomInteger(),
		// style
		lineColor: createRandomColor(),
		lineStyle: createRandomLineStyle(),
		sourceArrowStyle: createRandomArrowStyle(),
		targetArrowStyle: createRandomArrowStyle(),
	};
}

/**
 * Creates a random VertexTypeConfig object.
 * @returns A random VertexTypeConfig object.
 */
export function createRandomVertexTypeConfig(): VertexTypeConfig {
	const displayLabel = randomlyUndefined(createRandomName("displayLabel"));
	const hidden = randomlyUndefined(createRandomBoolean());
	return {
		type: createRandomVertexType(),
		attributes: createArray(6, createRandomAttributeConfig),
		...(displayLabel && { displayLabel }),
		...(hidden && { hidden }),
		total: createRandomInteger(),
		// style
		color: createRandomColor(),
		iconImageType: createRandomName("iconImageType"),
		iconUrl: createRandomUrlString(),
	};
}

/**
 * Creates a random EdgeConnection object.
 * Used for testing Schema Explorer edge connection functionality.
 * @returns A random EdgeConnection object.
 */
export function createRandomEdgeConnection(): EdgeConnection {
	return {
		edgeType: createRandomEdgeType(),
		sourceVertexType: createRandomVertexType(),
		targetVertexType: createRandomVertexType(),
		count: randomlyUndefined(createRandomInteger({ min: 1, max: 1000 })),
	};
}

/**
 * Creates a random schema object.
 * @returns A random SchemaStorageModel object.
 */
export function createRandomSchema(): SchemaStorageModel {
	const edges = createArray(3, createRandomEdgeTypeConfig);
	const vertices = createArray(3, createRandomVertexTypeConfig);
	const edgeConnections = createArray(5, createRandomEdgeConnection);
	const schema: SchemaStorageModel = {
		edges,
		vertices,
		edgeConnections,
		totalEdges: edges
			.map((e) => e.total ?? 0)
			.reduce((prev, current) => current + prev, 0),
		totalVertices: vertices
			.map((v) => v.total ?? 0)
			.reduce((prev, current) => current + prev, 0),
		lastUpdate: new Date(),
	};
	return schema;
}

/**
 * Creates random entities (nodes and edges).
 * @returns A random Entities object.
 */
export function createRandomEntities(): Entities {
	const nodes = createArray(3, createRandomVertex);
	const edges = [
		createRandomEdge(nodes[0], nodes[1]),
		createRandomEdge(nodes[0], nodes[2]),
		createRandomEdge(nodes[1], nodes[2]),

		// Reverse edges
		createRandomEdge(nodes[1], nodes[0]),
		createRandomEdge(nodes[2], nodes[0]),
		createRandomEdge(nodes[2], nodes[1]),
	];
	return { vertices: nodes, edges: edges };
}

/** Creates a random vertex ID. */
export function createRandomVertexId(): VertexId {
	return createVertexId(createRandomName("VertexId"));
}

/** Creates a random edge ID. */
export function createRandomEdgeId(): EdgeId {
	return createEdgeId(createRandomName("EdgeId"));
}

/** Creates a random configuration (connection) ID. */
export function createRandomConfigurationId(): ConfigurationId {
	return createNewConfigurationId();
}

export function createRandomVertexType(): VertexType {
	return createVertexType(createRandomName("VertexType"));
}

export function createRandomEdgeType(): EdgeType {
	return createEdgeType(createRandomName("EdgeType"));
}

/**
 * Creates a random vertex.
 * @returns A random Vertex object.
 */
export function createRandomVertex() {
	const label = createRandomName("VertexType");
	return createVertex({
		id: createRandomVertexId(),
		types: [label],
		attributes: createRecord(3, createRandomEntityAttribute),
	});
}

/**
 * Creates a random edge.
 * @returns A random Edge object.
 */
export function createRandomEdge(source?: Vertex, target?: Vertex) {
	const sourceId = source?.id ?? createRandomVertexId();
	const targetId = target?.id ?? createRandomVertexId();

	return createEdge({
		id: createRandomEdgeId(),
		type: createRandomName("EdgeType"),
		attributes: createRecord(3, createRandomEntityAttribute),
		sourceId,
		targetId,
	});
}

/**
 * Creates a testable vertex factory with random data and transformation methods.
 *
 * Generates a vertex with random properties that can be transformed into
 * different vertex types (Vertex, ResultVertex, PatchedResultVertex).
 *
 * @returns Testable vertex with methods: with(), asVertex(), asFragmentResult(), asResult(), asPatchedResult()
 */
export function createTestableVertex() {
	const createInternal = (testable: {
		id: EntityRawId;
		types: string[];
		attributes: EntityProperties;
	}) => {
		return {
			...testable,
			id: createVertexId(testable.id),
			types: testable.types as VertexType[],
			with: (newTestable: Partial<typeof testable>) => {
				return createInternal({ ...testable, ...newTestable });
			},
			asVertex: () =>
				createVertex({
					id: testable.id,
					types: testable.types,
					attributes: testable.attributes,
				}),
			asFragmentResult: (name?: string) =>
				createResultVertex({
					id: testable.id,
					name,
					types: testable.types,
				}),
			asResult: (name?: string) =>
				createResultVertex({
					id: testable.id,
					name,
					types: testable.types,
					attributes: testable.attributes,
				}),
			asPatchedResult: (name?: string) =>
				createPatchedResultVertex({
					id: testable.id,
					name,
					types: testable.types,
					attributes: testable.attributes,
				}),
		};
	};

	return createInternal({
		id: createRandomVertexId(),
		types: createArray(3, createRandomName),
		attributes: createRecord(3, createRandomEntityAttribute),
	});
}

/**
 * Creates a testable edge factory with random data and transformation methods.
 *
 * Generates an edge with random properties connecting two testable vertices
 * that can be transformed into different edge types (Edge, ResultEdge,
 * PatchedResultEdge). Maintains consistent relationships across
 * transformations.
 *
 * @returns Testable edge with methods: with(), withSource(), withTarget(), asEdge(), asFragmentResult(), asResult(), asPatchedResult()
 */
export function createTestableEdge() {
	// Factory method that creates the testable edge
	const createInternal = (testable: {
		id: EntityRawId;
		type: string;
		attributes: EntityProperties;
		source: TestableVertex;
		target: TestableVertex;
	}) => {
		return {
			...testable,
			id: createEdgeId(testable.id),
			type: testable.type as EdgeType,
			with: (newTestable: Partial<typeof testable>) => {
				return createInternal({ ...testable, ...newTestable });
			},
			withSource: (source: TestableVertex) => {
				return createInternal({ ...testable, source });
			},
			withTarget: (target: TestableVertex) => {
				return createInternal({ ...testable, target });
			},
			asEdge: () =>
				createEdge({
					id: testable.id,
					sourceId: testable.source.id,
					targetId: testable.target.id,
					type: testable.type,
					attributes: testable.attributes,
				}),
			asFragmentResult: (name?: string) =>
				createResultEdge({
					id: testable.id,
					sourceId: testable.source.id,
					targetId: testable.target.id,
					type: testable.type,
					name,
				}),
			asResult: (name?: string) =>
				createResultEdge({
					id: testable.id,
					sourceId: testable.source.id,
					targetId: testable.target.id,
					type: testable.type,
					attributes: testable.attributes,
					name,
				}),
			asPatchedResult: (name?: string) =>
				createPatchedResultEdge({
					id: testable.id,
					type: testable.type,
					attributes: testable.attributes,
					sourceVertex: testable.source.asVertex(),
					targetVertex: testable.target.asVertex(),
					name,
				}),
		};
	};

	// Use default values for a random property graph style edge
	return createInternal({
		id: createRandomEdgeId(),
		type: createRandomName("EdgeType"),
		source: createTestableVertex(),
		target: createTestableVertex(),
		attributes: createRecord(3, createRandomEntityAttribute),
	});
}

export type TestableEdge = ReturnType<typeof createTestableEdge>;
export type TestableVertex = ReturnType<typeof createTestableVertex>;

/**
 * Creates a random entity (vertex or edge) attribute.
 * @returns A random entity attribute object.
 */
export function createRandomEntityAttribute() {
	const valueTypes = ["string", "number", "boolean", "date"] as const;
	const randomIndex = Math.floor(Math.random() * valueTypes.length);
	const valueType = valueTypes[randomIndex];
	const value = (() => {
		switch (valueType) {
			case "string":
				return createRandomName("StringValue");
			case "number":
				return createRandomInteger();
			case "boolean":
				return createRandomBoolean();
			case "date":
				return createRandomDate();
		}
	})();

	return {
		key: createRandomName("EntityAttribute"),
		value,
	};
}

function pickRandomElement<T>(array: T[]): T {
	return array[Math.floor(Math.random() * array.length)];
}

export function createRandomExportedGraphConnection(): ExportedGraphConnection {
	const dbUrl = createRandomUrlString();
	const queryEngine = createRandomQueryEngine();
	return {
		dbUrl,
		queryEngine,
	};
}

export function createRandomVersion(): string {
	return `${createRandomInteger()}.${createRandomInteger()}.${createRandomInteger()}`;
}

export function createRandomExportedGraph() {
	const entities = createRandomEntities();
	const vertexIds = entities.vertices.map((v) => v.id);
	const edgeIds = entities.edges.map((e) => e.id);
	const connection = createRandomConnectionWithId();
	connection.queryEngine = pickRandomElement([...queryEngineOptions]);
	const result = createExportedGraph(vertexIds, edgeIds, connection);
	result.meta.sourceVersion = createRandomVersion();
	return result;
}

export function createRandomFile(): File {
	const fileName = createRandomName("File");
	const contentsString = createRandomName("Contents");
	const fileContent = new Blob([contentsString], { type: "text/plain" });
	const file = new File([fileContent], fileName, { type: "text/plain" });
	return file;
}

export function createRandomConnectionWithId(): ConnectionWithId {
	const fetchTimeoutMs = randomlyUndefined(createRandomInteger());
	const nodeExpansionLimit = randomlyUndefined(createRandomInteger());
	const queryEngine = createRandomQueryEngine();

	return {
		id: createNewConfigurationId(),
		displayLabel: createRandomName("displayLabel"),
		url: createRandomUrlString(),
		queryEngine,
		...(fetchTimeoutMs && { fetchTimeoutMs }),
		...(nodeExpansionLimit && { nodeExpansionLimit }),
	};
}

/**
 * Creates a random RawConfiguration object.
 * @returns A random RawConfiguration object.
 */
export function createRandomRawConfiguration(): RawConfiguration {
	const { id, displayLabel, ...connection } = createRandomConnectionWithId();

	return {
		id,
		displayLabel,
		connection,
	};
}

/**
 * Returnes a random query engine.
 * @param graphType - If "pg" then a random property graph query engine is
 * returned. If undefined then a random query engine is returned.
 * @returns The query engine.
 */
export function createRandomQueryEngine(graphType?: "pg"): QueryEngine {
	if (graphType === "pg") {
		return pickRandomElement(["gremlin", "openCypher"]);
	}
	return pickRandomElement([...queryEngineOptions]);
}

export function createRandomVertexStyleStorage(): VertexStyleStorage {
	const color = randomlyUndefined(createRandomColor());
	const borderColor = randomlyUndefined(createRandomColor());
	const longDisplayNameAttribute = randomlyUndefined(
		createRandomName("LongDisplayNameAttribute"),
	);
	const displayNameAttribute = randomlyUndefined(
		createRandomName("DisplayNameAttribute"),
	);
	const displayLabel = randomlyUndefined(createRandomName("DisplayLabel"));
	return {
		type: createRandomVertexType(),
		...(displayLabel && { displayLabel }),
		...(displayNameAttribute && { displayNameAttribute }),
		...(longDisplayNameAttribute && { longDisplayNameAttribute }),
		...(color && { color }),
		...(borderColor && { borderColor }),
	};
}

export function createRandomEdgeStyleStorage(): EdgeStyleStorage {
	const displayLabel = randomlyUndefined(createRandomName("DisplayLabel"));
	const displayNameAttribute = randomlyUndefined(
		createRandomName("DisplayNameAttribute"),
	);
	const lineColor = randomlyUndefined(createRandomColor());
	const labelColor = randomlyUndefined(createRandomColor());
	const labelBorderColor = randomlyUndefined(createRandomColor());
	const lineThickness = randomlyUndefined(createRandomInteger({ max: 25 }));
	return {
		type: createRandomEdgeType(),
		...(displayLabel && { displayLabel }),
		...(displayNameAttribute && { displayNameAttribute }),
		...(lineColor && { lineColor }),
		...(labelColor && { labelColor }),
		...(labelBorderColor && { labelBorderColor }),
		...(lineThickness && { lineThickness }),
	};
}

export function createRandomVertexStyles(): Map<
	VertexType,
	VertexStyleStorage
> {
	return new Map(
		createArray(3, createRandomVertexStyleStorage).map((style) => [
			style.type,
			style,
		]),
	);
}

export function createRandomEdgeStyles(): Map<EdgeType, EdgeStyleStorage> {
	return new Map(
		createArray(3, createRandomEdgeStyleStorage).map((style) => [
			style.type,
			style,
		]),
	);
}

export function createRandomVertexStyle(): Writable<VertexStyle> {
	const stored = createRandomVertexStyleStorage();
	return resolveVertexStyle(stored.type, stored);
}

export function createRandomEdgeStyle(): Writable<EdgeStyle> {
	const stored = createRandomEdgeStyleStorage();
	return resolveEdgeStyle(stored.type, stored);
}

export function createRandomLineStyle(): LineStyle {
	return pickRandomElement(["solid", "dotted", "dashed"]);
}

export function createRandomArrowStyle(): ArrowStyle {
	return pickRandomElement([
		"triangle",
		"triangle-tee",
		"circle-triangle",
		"triangle-cross",
		"triangle-backcurve",
		"tee",
		"vee",
		"square",
		"circle",
		"diamond",
		"none",
	]);
}

export function createRandomScalarValue() {
	const generators = [
		() => createRandomBoolean(),
		() => createRandomColor(),
		() => createRandomUrlString(),
		() => createRandomInteger(),
		() => createRandomDouble(),
		() => createRandomDate(),
		() => createRandomName(),
		() => createRandomVersion(),
		() => null,
	];
	const generator = pickRandomElement(generators);
	return generator();
}

export function createRandomeResultScalar() {
	return createResultScalar({
		name: createRandomName("name"),
		value: createRandomScalarValue(),
	});
}

/** Picks a random subset (possibly empty) of the given values. */
function randomSubset<T>(values: readonly T[]): Set<T> {
	return new Set(values.filter(() => createRandomBoolean()));
}

// Arbitrary but plausible pixel bounds for randomized panel dimensions. The
// exact values carry no meaning — they only need to be positive and varied so
// tests reading a width or height pin their own expected value.
const RANDOM_PANEL_MIN_PX = 100;
const RANDOM_PANEL_MAX_PX = 800;

/**
 * Creates a random {@link GraphViewLayout}. Every field spans its full
 * production domain — a closed sidebar (`null`), an empty toggle set, and
 * `undefined` optionals are all reachable — so tests that read layout must pin
 * the fields they depend on rather than relying on a convenient default.
 */
export function createRandomGraphViewLayout(): GraphViewLayout {
	return {
		activeSidebarItem: pickRandomElement([...graphViewSidebarItems, null]),
		sidebar: {
			width: createRandomInteger({
				min: RANDOM_PANEL_MIN_PX,
				max: RANDOM_PANEL_MAX_PX,
			}),
		},
		activeToggles: randomSubset(toggleableViews),
		tableView: randomlyUndefined({
			height: createRandomInteger({
				min: RANDOM_PANEL_MIN_PX,
				max: RANDOM_PANEL_MAX_PX,
			}),
		}),
		detailsAutoOpenOnSelection: randomlyUndefined(createRandomBoolean()),
	};
}

/**
 * Creates a random {@link SchemaViewLayout}. As with the graph view layout,
 * every field spans its full production domain, including a closed sidebar and
 * an `undefined` auto-open flag.
 */
export function createRandomSchemaViewLayout(): SchemaViewLayout {
	return {
		activeSidebarItem: pickRandomElement([...schemaViewSidebarItems, null]),
		sidebar: {
			width: createRandomInteger({
				min: RANDOM_PANEL_MIN_PX,
				max: RANDOM_PANEL_MAX_PX,
			}),
		},
		detailsAutoOpenOnSelection: randomlyUndefined(createRandomBoolean()),
	};
}
