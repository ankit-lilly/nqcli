// @vitest-environment happy-dom
import { createArray, createRandomName } from "@/utils/testing/random";
import { act } from "@testing-library/react";
import { useAtomValue } from "jotai";

import {
	activeConfigurationAtom,
	configurationAtom,
	createEdgeConnectionId,
	schemaAtom,
} from "@/core";
import { LABELS } from "@/utils";
import {
	createRandomEdge,
	createRandomEdgeConnection,
	createRandomRawConfiguration,
	createRandomSchema,
	createRandomVertex,
	createRandomVertexType,
	DbState,
	renderHookWithJotai,
	renderHookWithState,
} from "@/utils/testing";

import {
	createEdgeConnection,
	type EdgeTypeConfig,
	type VertexTypeConfig,
} from "../ConfigurationProvider";
import {
	createEdge,
	createEdgeType,
	createVertex,
	createVertexId,
	createVertexType,
	type EntityProperties,
} from "../entities";
import {
	activeSchemaSelector,
	createVertexTypeLookup,
	mapEdgeToTypeConfig,
	mapVertexToTypeConfigs,
	maybeActiveSchemaAtom,
	type SchemaStorageModel,
	updateSchemaFromEntities,
	useActiveSchema,
	useGraphSchema,
	useHasActiveSchema,
	useMaybeActiveSchema,
	useUpdateSchemaFromEntities,
} from "./schema";

const EMPTY_VERTEX_TYPE_LOOKUP = createVertexTypeLookup();

describe("schema", () => {
	describe("mapVertexToTypeConfigs", () => {
		it("should work with vertex", () => {
			const entity = createVertex({
				id: "1",
				types: ["Person", "Manager"],
				attributes: {
					name: "John",
					age: 30,
				},
			});
			const expected: VertexTypeConfig[] = [
				{
					type: createVertexType("Person"),
					attributes: [
						{
							name: "name",
							dataType: "String",
						},
						{
							name: "age",
							dataType: "Number",
						},
					],
				},
				{
					type: createVertexType("Manager"),
					attributes: [
						{
							name: "name",
							dataType: "String",
						},
						{
							name: "age",
							dataType: "Number",
						},
					],
				},
			];
			const result = mapVertexToTypeConfigs(entity);
			expect(result).toStrictEqual(expected);
		});
	});

	describe("mapEdgeToTypeConfig", () => {
		it("should work with edge", () => {
			const entity = createEdge({
				id: "1",
				type: "WORKS_FOR",
				attributes: {
					since: 2020,
				},
				sourceId: "2",
				targetId: "3",
			});
			const expected: EdgeTypeConfig = {
				type: createEdgeType("WORKS_FOR"),
				attributes: [
					{
						name: "since",
						dataType: "Number",
					},
				],
			};
			const result = mapEdgeToTypeConfig(entity);
			expect(result).toStrictEqual(expected);
		});
	});

	describe("createVertexTypeLookup", () => {
		it("should resolve vertex types by ID", () => {
			const v1 = createVertex({ id: "v1", types: ["Person"], attributes: {} });
			const v2 = createVertex({
				id: "v2",
				types: ["Dog", "Pet"],
				attributes: {},
			});
			const map = new Map([
				[v1.id, v1],
				[v2.id, v2],
			]);

			const lookup = createVertexTypeLookup(map);

			expect(lookup.get(v1.id)).toStrictEqual(v1.types);
			expect(lookup.get(v2.id)).toStrictEqual(v2.types);
		});

		it("should chain multiple maps", () => {
			const v1 = createVertex({ id: "v1", types: ["Person"], attributes: {} });
			const v2 = createVertex({ id: "v2", types: ["Dog"], attributes: {} });

			const lookup = createVertexTypeLookup(
				new Map([[v1.id, v1]]),
				new Map([[v2.id, v2]]),
			);

			expect(lookup.get(v1.id)).toStrictEqual(v1.types);
			expect(lookup.get(v2.id)).toStrictEqual(v2.types);
		});

		it("should return types from the first map that contains the vertex", () => {
			const v1 = createVertex({ id: "v1", types: ["Person"], attributes: {} });
			const v1Duplicate = createVertex({
				id: "v1",
				types: ["Employee"],
				attributes: {},
			});

			const lookup = createVertexTypeLookup(
				new Map([[v1.id, v1]]),
				new Map([[v1Duplicate.id, v1Duplicate]]),
			);

			expect(lookup.get(v1.id)).toStrictEqual(v1.types);
		});

		it("should return undefined for unknown vertex IDs", () => {
			const v1 = createVertex({ id: "v1", types: ["Person"], attributes: {} });
			const lookup = createVertexTypeLookup(new Map([[v1.id, v1]]));

			expect(lookup.get(createVertexId("v99"))).toBeUndefined();
		});

		it("should report isEmpty when all maps are empty", () => {
			const lookup = createVertexTypeLookup(new Map(), new Map());

			expect(lookup.isEmpty()).toBe(true);
		});

		it("should report not empty when any map has entries", () => {
			const v1 = createVertex({ id: "v1", types: ["Person"], attributes: {} });
			const lookup = createVertexTypeLookup(new Map(), new Map([[v1.id, v1]]));

			expect(lookup.isEmpty()).toBe(false);
		});
	});

	describe("updateSchemaFromEntities edge connections", () => {
		it("should not change edgeConnections when vertex lookup is empty", () => {
			const schema = createRandomSchema();
			const edge = createRandomEdge();

			const result = updateSchemaFromEntities(
				{ edges: [edge] },
				schema,
				EMPTY_VERTEX_TYPE_LOOKUP,
			);

			expect(result.edgeConnections).toBe(schema.edgeConnections);
		});

		it("should infer connections when schema has no prior edgeConnections", () => {
			const source = createVertex({
				id: "v1",
				types: ["Person"],
				attributes: {},
			});
			const target = createVertex({ id: "v2", types: ["Dog"], attributes: {} });
			const edge = createEdge({
				id: "e1",
				type: "owner",
				sourceId: "v1",
				targetId: "v2",
				attributes: {},
			});

			const schema: SchemaStorageModel = { vertices: [], edges: [] };
			expect(schema.edgeConnections).toBeUndefined();

			const vertexLookup = createVertexTypeLookup(
				new Map([
					[source.id, source],
					[target.id, target],
				]),
			);

			const result = updateSchemaFromEntities(
				{ edges: [edge], vertices: [source, target] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toStrictEqual([
				createEdgeConnection({
					source: "Person",
					edge: "owner",
					target: "Dog",
				}),
			]);
		});

		it("should not change edgeConnections when entities have no edges", () => {
			const schema = createRandomSchema();
			const vertex = createRandomVertex();

			const vertexLookup = createVertexTypeLookup(
				new Map([[vertex.id, vertex]]),
			);

			const result = updateSchemaFromEntities(
				{ vertices: [vertex] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toBe(schema.edgeConnections);
		});

		it("should infer edge connections from edges and vertex lookup", () => {
			const source = createVertex({
				id: "v1",
				types: ["Person"],
				attributes: {},
			});
			const target = createVertex({ id: "v2", types: ["Dog"], attributes: {} });
			const edge = createEdge({
				id: "e1",
				type: "owner",
				sourceId: "v1",
				targetId: "v2",
				attributes: {},
			});

			const schema: SchemaStorageModel = { vertices: [], edges: [] };
			const vertexLookup = createVertexTypeLookup(
				new Map([
					[source.id, source],
					[target.id, target],
				]),
			);

			const result = updateSchemaFromEntities(
				{ edges: [edge], vertices: [source, target] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toStrictEqual([
				createEdgeConnection({
					source: "Person",
					edge: "owner",
					target: "Dog",
				}),
			]);
		});

		it("should create connections for all type combinations of multi-label vertices", () => {
			const source = createVertex({
				id: "v1",
				types: ["Person", "Employee"],
				attributes: {},
			});
			const target = createVertex({
				id: "v2",
				types: ["Company"],
				attributes: {},
			});
			const edge = createEdge({
				id: "e1",
				type: "worksAt",
				sourceId: "v1",
				targetId: "v2",
				attributes: {},
			});

			const schema: SchemaStorageModel = { vertices: [], edges: [] };
			const vertexLookup = createVertexTypeLookup(
				new Map([
					[source.id, source],
					[target.id, target],
				]),
			);

			const result = updateSchemaFromEntities(
				{ edges: [edge], vertices: [source, target] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toStrictEqual([
				createEdgeConnection({
					source: "Person",
					edge: "worksAt",
					target: "Company",
				}),
				createEdgeConnection({
					source: "Employee",
					edge: "worksAt",
					target: "Company",
				}),
			]);
		});

		it("should skip edges where source or target vertex type cannot be resolved", () => {
			const source = createVertex({
				id: "v1",
				types: ["Person"],
				attributes: {},
			});
			const edge = createEdge({
				id: "e1",
				type: "knows",
				sourceId: "v1",
				targetId: "v2",
				attributes: {},
			});

			const schema: SchemaStorageModel = {
				vertices: [],
				edges: [],
				edgeConnections: [],
			};
			const vertexLookup = createVertexTypeLookup(
				new Map([[source.id, source]]),
			);

			const result = updateSchemaFromEntities(
				{ edges: [edge], vertices: [source] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toBe(schema.edgeConnections);
		});

		it("should preserve existing edge connections including count", () => {
			const existingConnection = createEdgeConnection({
				source: "Cat",
				edge: "chases",
				target: "Mouse",
				count: 7,
			});
			const source = createVertex({
				id: "v1",
				types: ["Person"],
				attributes: {},
			});
			const target = createVertex({ id: "v2", types: ["Dog"], attributes: {} });
			const edge = createEdge({
				id: "e1",
				type: "owner",
				sourceId: "v1",
				targetId: "v2",
				attributes: {},
			});

			const schema: SchemaStorageModel = {
				vertices: [],
				edges: [],
				edgeConnections: [existingConnection],
			};
			const vertexLookup = createVertexTypeLookup(
				new Map([
					[source.id, source],
					[target.id, target],
				]),
			);

			const result = updateSchemaFromEntities(
				{ edges: [edge], vertices: [source, target] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toStrictEqual([
				existingConnection,
				createEdgeConnection({
					source: "Person",
					edge: "owner",
					target: "Dog",
				}),
			]);
		});

		it("should not add duplicate edge connections", () => {
			const source = createVertex({
				id: "v1",
				types: ["Person"],
				attributes: {},
			});
			const target = createVertex({ id: "v2", types: ["Dog"], attributes: {} });
			const existingConnection = createEdgeConnection({
				source: "Person",
				edge: "owner",
				target: "Dog",
				count: 42,
			});
			const edge = createEdge({
				id: "e1",
				type: "owner",
				sourceId: "v1",
				targetId: "v2",
				attributes: {},
			});

			const schema: SchemaStorageModel = {
				vertices: [],
				edges: [],
				edgeConnections: [existingConnection],
			};
			const vertexLookup = createVertexTypeLookup(
				new Map([
					[source.id, source],
					[target.id, target],
				]),
			);

			const result = updateSchemaFromEntities(
				{ edges: [edge], vertices: [source, target] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toBe(schema.edgeConnections);
		});

		it("should preserve edgeConnections reference when no new connections are discovered", () => {
			const source = createVertex({
				id: "v1",
				types: ["Person"],
				attributes: {},
			});
			const target = createVertex({ id: "v2", types: ["Dog"], attributes: {} });
			const existingConnection = createEdgeConnection({
				source: "Person",
				edge: "owner",
				target: "Dog",
			});
			const edge = createEdge({
				id: "e1",
				type: "owner",
				sourceId: "v1",
				targetId: "v2",
				attributes: {},
			});

			const schema: SchemaStorageModel = {
				vertices: [],
				edges: [],
				edgeConnections: [existingConnection],
			};
			const vertexLookup = createVertexTypeLookup(
				new Map([
					[source.id, source],
					[target.id, target],
				]),
			);

			const result = updateSchemaFromEntities(
				{ edges: [edge], vertices: [source, target] },
				schema,
				vertexLookup,
			);

			expect(result.edgeConnections).toBe(schema.edgeConnections);
		});
	});

	describe("updateSchemaFromEntities", () => {
		it("should do nothing when no entities", () => {
			const originalSchema = createRandomSchema();
			const result = updateSchemaFromEntities(
				{ vertices: [], edges: [] },
				originalSchema,
				EMPTY_VERTEX_TYPE_LOOKUP,
			);

			expect(result).toEqual(originalSchema);
		});

		it("should add missing entities to schema", () => {
			const originalSchema = createRandomSchema();
			const newNodes = createArray(3, () => createRandomVertex());
			const newEdges = createArray(3, () => createRandomEdge());
			const result = updateSchemaFromEntities(
				{ vertices: newNodes, edges: newEdges },
				originalSchema,
				EMPTY_VERTEX_TYPE_LOOKUP,
			);

			expect(result).toEqual({
				...originalSchema,
				vertices: [
					...originalSchema.vertices,
					...newNodes.flatMap(mapVertexToTypeConfigs),
				],
				edges: [...originalSchema.edges, ...newEdges.map(mapEdgeToTypeConfig)],
			});
		});

		it("should merge entities that have duplicate IDs and types but different attributes", () => {
			const originalSchema = createRandomSchema();
			const newNode1 = createRandomVertex();
			const newNode2 = createRandomVertex();
			const newEdge1 = createRandomEdge();
			const newEdge2 = createRandomEdge();
			newNode2.id = newNode1.id;
			newEdge2.id = newEdge1.id;
			newNode2.type = newNode1.type;
			newNode2.types = newNode1.types;
			newEdge2.type = newEdge1.type;

			const result = updateSchemaFromEntities(
				{ vertices: [newNode1, newNode2], edges: [newEdge1, newEdge2] },
				originalSchema,
				EMPTY_VERTEX_TYPE_LOOKUP,
			);

			const node1Configs = mapVertexToTypeConfigs(newNode1);
			const node2Configs = mapVertexToTypeConfigs(newNode2);
			const mergedNodeConfigs = node1Configs.map((config1) => ({
				...config1,
				attributes: [
					...config1.attributes,
					...node2Configs.find((c) => c.type === config1.type)!.attributes,
				],
			}));

			expect(result).toEqual({
				...originalSchema,
				vertices: [...originalSchema.vertices, ...mergedNodeConfigs],
				edges: [
					...originalSchema.edges,
					{
						...mapEdgeToTypeConfig(newEdge1),
						attributes: [
							...mapEdgeToTypeConfig(newEdge1).attributes,
							...mapEdgeToTypeConfig(newEdge2).attributes,
						],
					},
				],
			});
		});

		it("should add missing attributes to schema", () => {
			const originalSchema = createRandomSchema();
			const newNode = createRandomVertex();
			const newEdge = createRandomEdge();
			newNode.type = originalSchema.vertices[0].type;
			newNode.types = [originalSchema.vertices[0].type];
			newEdge.type = originalSchema.edges[0].type;

			const extractedNodeConfigs = mapVertexToTypeConfigs(newNode);
			const extractedEdgeConfig = mapEdgeToTypeConfig(newEdge);

			const result = updateSchemaFromEntities(
				{ vertices: [newNode], edges: [newEdge] },
				originalSchema,
				EMPTY_VERTEX_TYPE_LOOKUP,
			);

			expect(result.vertices.flatMap((v) => v.attributes)).toEqual(
				expect.arrayContaining([
					...originalSchema.vertices.flatMap((v) => v.attributes),
					...extractedNodeConfigs.flatMap((c) => c.attributes),
				]),
			);

			expect(result.edges.flatMap((v) => v.attributes)).toEqual(
				expect.arrayContaining([
					...originalSchema.edges.flatMap((v) => v.attributes),
					...extractedEdgeConfig.attributes,
				]),
			);
		});

		it("should add vertices with no type to the schema", () => {
			const originalSchema = createRandomSchema();
			originalSchema.vertices = [];
			originalSchema.edges = [];

			const newNodes = createArray(3, () => createRandomVertex()).map(
				(vertex) => createVertex({ ...vertex, types: [] }),
			);

			const result = updateSchemaFromEntities(
				{ vertices: newNodes, edges: [] },
				originalSchema,
				EMPTY_VERTEX_TYPE_LOOKUP,
			);

			const emptyTypeVtConfigs = result.vertices.filter(
				(v) => v.type === LABELS.MISSING_TYPE,
			);
			expect(emptyTypeVtConfigs).toHaveLength(1);
			expect(emptyTypeVtConfigs[0].attributes).toEqual(
				newNodes.flatMap(mapVertexToTypeConfigs).flatMap((n) => n.attributes),
			);
		});

		it("should add all types from a multi-label vertex", () => {
			const schema: SchemaStorageModel = {
				vertices: [],
				edges: [],
			};

			const vertex = createVertex({
				id: "1",
				types: ["Person", "Employee"],
				attributes: { name: "Alice" },
			});

			const result = updateSchemaFromEntities(
				{ vertices: [vertex] },
				schema,
				EMPTY_VERTEX_TYPE_LOOKUP,
			);

			expect(result.vertices).toHaveLength(2);
			expect(result.vertices.map((v) => v.type)).toStrictEqual([
				createVertexType("Person"),
				createVertexType("Employee"),
			]);
			expect(result.vertices[0].attributes).toStrictEqual([
				{ name: "name", dataType: "String" },
			]);
			expect(result.vertices[1].attributes).toStrictEqual([
				{ name: "name", dataType: "String" },
			]);
		});
	});
});

describe("referential integrity", () => {
	it("should preserve schema reference when no changes", () => {
		const schema = createRandomSchema();
		const vertex = createRandomVertex();
		vertex.type = schema.vertices[0].type;
		vertex.types = [schema.vertices[0].type];
		vertex.attributes = schema.vertices[0].attributes.reduce((acc, attr) => {
			acc[attr.name] = createRandomName("value");
			return acc;
		}, {} as EntityProperties);

		const result = updateSchemaFromEntities(
			{ vertices: [vertex] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result).toBe(schema);
	});

	it("should preserve vertex config reference when no changes", () => {
		const schema = createRandomSchema();
		const vertex = createRandomVertex();
		vertex.type = schema.vertices[0].type;
		vertex.types = [schema.vertices[0].type];
		vertex.attributes = schema.vertices[0].attributes.reduce((acc, attr) => {
			acc[attr.name] = createRandomName("value");
			return acc;
		}, {} as EntityProperties);

		const result = updateSchemaFromEntities(
			{ vertices: [vertex] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result.vertices[0]).toBe(schema.vertices[0]);
	});

	it("should preserve edge config reference when no changes", () => {
		const schema = createRandomSchema();
		const edge = createRandomEdge();
		edge.type = schema.edges[0].type;
		edge.attributes = schema.edges[0].attributes.reduce((acc, attr) => {
			acc[attr.name] = createRandomName("value");
			return acc;
		}, {} as EntityProperties);

		const result = updateSchemaFromEntities(
			{ edges: [edge] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result.edges[0]).toBe(schema.edges[0]);
	});

	it("should preserve attributes reference when no changes", () => {
		const schema = createRandomSchema();
		const vertex = createRandomVertex();
		vertex.type = schema.vertices[0].type;
		vertex.types = [schema.vertices[0].type];
		vertex.attributes = schema.vertices[0].attributes.reduce((acc, attr) => {
			acc[attr.name] = createRandomName("value");
			return acc;
		}, {} as EntityProperties);

		const result = updateSchemaFromEntities(
			{ vertices: [vertex] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result.vertices[0].attributes).toBe(schema.vertices[0].attributes);
	});

	it("should create new vertex config when attributes change", () => {
		const schema = createRandomSchema();
		const vertex = createRandomVertex();
		vertex.type = schema.vertices[0].type;
		vertex.types = [schema.vertices[0].type];

		const result = updateSchemaFromEntities(
			{ vertices: [vertex] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result.vertices[0]).not.toBe(schema.vertices[0]);
		expect(result.vertices[0].attributes).not.toBe(
			schema.vertices[0].attributes,
		);
	});

	it("should create new edge config when attributes change", () => {
		const schema = createRandomSchema();
		const edge = createRandomEdge();
		edge.type = schema.edges[0].type;

		const result = updateSchemaFromEntities(
			{ edges: [edge] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result.edges[0]).not.toBe(schema.edges[0]);
		expect(result.edges[0].attributes).not.toBe(schema.edges[0].attributes);
	});

	it("should preserve vertices array reference when only edges change", () => {
		const schema = createRandomSchema();
		const edge = createRandomEdge();

		const result = updateSchemaFromEntities(
			{ edges: [edge] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result.vertices).toBe(schema.vertices);
	});

	it("should preserve edges array reference when only vertices change", () => {
		const schema = createRandomSchema();
		const vertex = createRandomVertex();

		const result = updateSchemaFromEntities(
			{ vertices: [vertex] },
			schema,
			EMPTY_VERTEX_TYPE_LOOKUP,
		);

		expect(result.edges).toBe(schema.edges);
	});
});

describe("useGraphSchema edgeConnections", () => {
	test("should return empty edgeConnections when schema has none", () => {
		const state = new DbState();
		state.activeSchema.edgeConnections = [];

		const { result } = renderHookWithState(() => useGraphSchema(), state);

		expect(result.current.edgeConnections.byVertexType.size).toBe(0);
		expect(result.current.edgeConnections.byEdgeConnectionId.size).toBe(0);
	});

	test("should group connections by vertex type", () => {
		const sharedVertexType = createRandomVertexType();
		const conn1 = createRandomEdgeConnection();
		const conn2 = createRandomEdgeConnection();
		conn1.sourceVertexType = sharedVertexType;
		conn2.targetVertexType = sharedVertexType;

		const state = new DbState();
		state.activeSchema.edgeConnections = [conn1, conn2];

		const { result } = renderHookWithState(() => useGraphSchema(), state);

		const connectionsForShared =
			result.current.edgeConnections.byVertexType.get(sharedVertexType);
		expect(connectionsForShared).toStrictEqual([conn1, conn2]);
	});

	test("forVertexType should return connections for a vertex type", () => {
		const vertexType = createRandomVertexType();
		const conn = createRandomEdgeConnection();
		conn.sourceVertexType = vertexType;

		const state = new DbState();
		state.activeSchema.edgeConnections = [conn];

		const { result } = renderHookWithState(() => useGraphSchema(), state);

		expect(
			result.current.edgeConnections.forVertexType(vertexType),
		).toStrictEqual([conn]);
	});

	test("forVertexType should return empty array for unknown vertex type", () => {
		const state = new DbState();
		state.activeSchema.edgeConnections = [createRandomEdgeConnection()];

		const { result } = renderHookWithState(() => useGraphSchema(), state);

		expect(
			result.current.edgeConnections.forVertexType(createRandomVertexType()),
		).toStrictEqual([]);
	});

	test("should not duplicate connection when source equals target", () => {
		const vertexType = createRandomVertexType();
		const conn = createRandomEdgeConnection();
		conn.sourceVertexType = vertexType;
		conn.targetVertexType = vertexType;

		const state = new DbState();
		state.activeSchema.edgeConnections = [conn];

		const { result } = renderHookWithState(() => useGraphSchema(), state);

		expect(
			result.current.edgeConnections.forVertexType(vertexType),
		).toStrictEqual([conn]);
	});

	test("byEdgeConnectionId should map connections by their ID", () => {
		const conn1 = createRandomEdgeConnection();
		const conn2 = createRandomEdgeConnection();
		const state = new DbState();
		state.activeSchema.edgeConnections = [conn1, conn2];

		const { result } = renderHookWithState(() => useGraphSchema(), state);

		expect(result.current.edgeConnections.byEdgeConnectionId.size).toBe(2);
	});

	test("byEdgeConnectionId should allow lookup by ID", () => {
		const conn = createRandomEdgeConnection();
		const id = createEdgeConnectionId(conn);
		const state = new DbState();
		state.activeSchema.edgeConnections = [conn];

		const { result } = renderHookWithState(() => useGraphSchema(), state);

		expect(
			result.current.edgeConnections.byEdgeConnectionId.get(id),
		).toStrictEqual(conn);
	});
});

describe("useUpdateSchemaFromEntities", () => {
	it("should not churn schemaAtom when active config has no schema entry", () => {
		const state = new DbState().withNoActiveSchema();

		const { result } = renderHookWithState(
			() => ({
				updateSchema: useUpdateSchemaFromEntities(),
				schemaMap: useAtomValue(schemaAtom),
			}),
			state,
		);

		const schemaMapBefore = result.current.schemaMap;

		act(() => {
			result.current.updateSchema({
				vertices: [createRandomVertex()],
			});
		});

		expect(result.current.schemaMap).toBe(schemaMapBefore);
	});

	it("should not churn schemaAtom when entities already match schema", () => {
		const state = new DbState();
		const existingVertex = createRandomVertex();
		existingVertex.type = state.activeSchema.vertices[0].type;
		existingVertex.types = [state.activeSchema.vertices[0].type];
		existingVertex.attributes =
			state.activeSchema.vertices[0].attributes.reduce((acc, attr) => {
				acc[attr.name] = createRandomName("value");
				return acc;
			}, {} as EntityProperties);

		const { result } = renderHookWithState(
			() => ({
				updateSchema: useUpdateSchemaFromEntities(),
				schemaMap: useAtomValue(schemaAtom),
			}),
			state,
		);

		const schemaMapBefore = result.current.schemaMap;

		act(() => {
			result.current.updateSchema({
				vertices: [existingVertex],
			});
		});

		expect(result.current.schemaMap).toBe(schemaMapBefore);
	});

	it("should infer edge connections from entities and canvas vertices", () => {
		const state = new DbState();
		state.activeSchema.edgeConnections = [];

		// Source vertex is on the canvas
		const source = createVertex({
			id: "v1",
			types: ["Person"],
			attributes: {},
		});
		state.addVertexToGraph(source);

		// Target vertex and edge arrive via entities
		const target = createVertex({
			id: "v2",
			types: ["Dog"],
			attributes: {},
		});
		const edge = createEdge({
			id: "e1",
			type: "owner",
			sourceId: "v1",
			targetId: "v2",
			attributes: {},
		});

		const { result } = renderHookWithState(
			() => ({
				updateSchema: useUpdateSchemaFromEntities(),
				schema: useAtomValue(activeSchemaSelector),
			}),
			state,
		);

		act(() => {
			result.current.updateSchema({
				vertices: [target],
				edges: [edge],
			});
		});

		expect(result.current.schema?.edgeConnections).toStrictEqual([
			createEdgeConnection({ source: "Person", edge: "owner", target: "Dog" }),
		]);
	});

	it("should prefer entity vertices over canvas vertices for edge connections", () => {
		const state = new DbState();
		state.activeSchema.edgeConnections = [];

		// Canvas has v1 as "Employee"
		const canvasSource = createVertex({
			id: "v1",
			types: ["Employee"],
			attributes: {},
		});
		state.addVertexToGraph(canvasSource);

		// Entities provide v1 as "Person" (should take priority)
		const entitySource = createVertex({
			id: "v1",
			types: ["Person"],
			attributes: {},
		});
		const target = createVertex({
			id: "v2",
			types: ["Dog"],
			attributes: {},
		});
		const edge = createEdge({
			id: "e1",
			type: "owner",
			sourceId: "v1",
			targetId: "v2",
			attributes: {},
		});

		const { result } = renderHookWithState(
			() => ({
				updateSchema: useUpdateSchemaFromEntities(),
				schema: useAtomValue(activeSchemaSelector),
			}),
			state,
		);

		act(() => {
			result.current.updateSchema({
				vertices: [entitySource, target],
				edges: [edge],
			});
		});

		expect(result.current.schema?.edgeConnections).toStrictEqual([
			createEdgeConnection({ source: "Person", edge: "owner", target: "Dog" }),
		]);
	});
});

describe("useHasActiveSchema", () => {
	test("should return false when schema has no lastUpdate", () => {
		const state = new DbState();
		state.activeSchema.lastUpdate = undefined;

		const { result } = renderHookWithState(() => useHasActiveSchema(), state);

		expect(result.current).toBe(false);
	});

	test("should return true when schema has lastUpdate", () => {
		const state = new DbState();
		state.activeSchema.lastUpdate = new Date();

		const { result } = renderHookWithState(() => useHasActiveSchema(), state);

		expect(result.current).toBe(true);
	});
});

describe("maybeActiveSchemaAtom", () => {
	it("returns undefined when no active configuration exists", () => {
		const { result } = renderHookWithJotai(() =>
			useAtomValue(maybeActiveSchemaAtom),
		);

		expect(result.current).toBeUndefined();
	});

	it("returns undefined when active config has no schema", () => {
		const config = createRandomRawConfiguration();

		const { result } = renderHookWithJotai(
			() => useAtomValue(maybeActiveSchemaAtom),
			(store) => {
				store.set(configurationAtom, new Map([[config.id, config]]));
				store.set(activeConfigurationAtom, config.id);
				store.set(schemaAtom, new Map());
			},
		);

		expect(result.current).toBeUndefined();
	});

	it("returns the schema when one exists", () => {
		const config = createRandomRawConfiguration();
		const schema = createRandomSchema();

		const { result } = renderHookWithJotai(
			() => useAtomValue(maybeActiveSchemaAtom),
			(store) => {
				store.set(configurationAtom, new Map([[config.id, config]]));
				store.set(activeConfigurationAtom, config.id);
				store.set(schemaAtom, new Map([[config.id, schema]]));
			},
		);

		expect(result.current).toStrictEqual(schema);
	});
});

describe("useMaybeActiveSchema", () => {
	it("returns undefined when no active configuration exists", () => {
		const { result } = renderHookWithJotai(() => useMaybeActiveSchema());

		expect(result.current).toBeUndefined();
	});

	it("returns undefined when active config has no schema", () => {
		const config = createRandomRawConfiguration();

		const { result } = renderHookWithJotai(
			() => useMaybeActiveSchema(),
			(store) => {
				store.set(configurationAtom, new Map([[config.id, config]]));
				store.set(activeConfigurationAtom, config.id);
				store.set(schemaAtom, new Map());
			},
		);

		expect(result.current).toBeUndefined();
	});

	it("returns the schema when one exists", () => {
		const config = createRandomRawConfiguration();
		const schema = createRandomSchema();

		const { result } = renderHookWithJotai(
			() => useMaybeActiveSchema(),
			(store) => {
				store.set(configurationAtom, new Map([[config.id, config]]));
				store.set(activeConfigurationAtom, config.id);
				store.set(schemaAtom, new Map([[config.id, schema]]));
			},
		);

		expect(result.current).toStrictEqual(schema);
	});
});

describe("useActiveSchema", () => {
	it("returns empty schema when no active configuration exists", () => {
		const { result } = renderHookWithJotai(() => useActiveSchema());

		expect(result.current).toStrictEqual({
			vertices: [],
			edges: [],
			edgeConnections: [],
		});
	});

	it("returns empty schema when active config has no schema", () => {
		const config = createRandomRawConfiguration();

		const { result } = renderHookWithJotai(
			() => useActiveSchema(),
			(store) => {
				store.set(configurationAtom, new Map([[config.id, config]]));
				store.set(activeConfigurationAtom, config.id);
				store.set(schemaAtom, new Map());
			},
		);

		expect(result.current).toStrictEqual({
			vertices: [],
			edges: [],
			edgeConnections: [],
		});
	});

	it("returns the schema when one exists", () => {
		const config = createRandomRawConfiguration();
		const schema = createRandomSchema();

		const { result } = renderHookWithJotai(
			() => useActiveSchema(),
			(store) => {
				store.set(configurationAtom, new Map([[config.id, config]]));
				store.set(activeConfigurationAtom, config.id);
				store.set(schemaAtom, new Map([[config.id, schema]]));
			},
		);

		expect(result.current).toStrictEqual(schema);
	});
});
