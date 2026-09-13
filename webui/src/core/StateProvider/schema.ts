import type { Simplify } from "type-fest";

import { atom, useAtomValue } from "jotai";
import { atomFamily } from "jotai-family";
import { RESET, useAtomCallback } from "jotai/utils";
import { useCallback, useDeferredValue } from "react";

import {
	createVertexTypeLookup,
	mapEdgeToTypeConfig as mapEdgeToSchema,
	mapVertexToTypeConfigs as mapVertexToSchemas,
	type SchemaModel,
	updateSchemaFromEntities,
	type VertexTypeLookup,
} from "@/application";
import type {
	ConfigurationId,
	EdgeConnection,
	EdgeTypeConfig,
	VertexTypeConfig,
} from "@/core/ConfigurationProvider";
import type { SetStateActionWithReset } from "@/utils/jotai";

import {
	activeConfigurationAtom,
	createEdgeConnectionId,
	type Edge,
	type EdgeConnectionId,
	type EdgeType,
	type Entities,
	schemaAtom,
	type Vertex,
	type VertexType,
} from "@/core";
import { logger } from "@/utils";

import { nodesAtom, toNodeMap } from "./nodes";

/**
 * Persisted schema state for a database connection.
 *
 * This is the runtime representation of the discovered graph schema, stored in
 * Jotai atoms and persisted to IndexedDB. It gets populated from database
 * schema queries and incrementally updated as users explore the graph.
 */
export type SchemaStorageModel = SchemaModel;

/** Grabs a specific schema out of the map, or returns the empty schema */
const schemaByIdAtom = atomFamily((id: ConfigurationId | null) => {
	if (!id) {
		return atom(emptySchema);
	}
	return atom((get) => {
		logger.debug("Creating active schema", id);
		const schemaMap = get(schemaAtom);
		return schemaMap.get(id) ?? emptySchema;
	});
});

const emptySchema: SchemaStorageModel = {
	vertices: [],
	edges: [],
	edgeConnections: [],
};

/** Gets the active schema from storage, or undefined if one doesn't exist. */
export const maybeActiveSchemaAtom = atom((get) => {
	const id = get(activeConfigurationAtom);
	if (!id) {
		return undefined;
	}

	return get(schemaAtom).get(id);
});

export const activeSchemaAtom = atom((get) => {
	const id = get(activeConfigurationAtom);
	return get(schemaByIdAtom(id));
});

/**
 * Hook to check if the active schema has been synchronized from the database.
 *
 * @returns True if the active schema has a lastUpdate timestamp, indicating
 * it has been populated from a database schema query at least once.
 */
export function useHasActiveSchema() {
	const activeSchema = useAtomValue(maybeActiveSchemaAtom);
	return !!activeSchema?.lastUpdate;
}

/** Gets the stored active schema or a default empty schema */
export function useActiveSchema(): SchemaStorageModel {
	return useDeferredValue(useAtomValue(activeSchemaAtom));
}

/** Gets the stored active schema if one exists for the active connection */
export function useMaybeActiveSchema(): SchemaStorageModel | undefined {
	return useDeferredValue(useAtomValue(maybeActiveSchemaAtom));
}

function createVertexSchema(vtConfig: VertexTypeConfig) {
	return {
		type: vtConfig.type,
		attributes: vtConfig.attributes.map((attr) => ({
			name: attr.name,
			dataType: attr.dataType ?? "String",
		})),
	};
}

export type VertexSchema = Simplify<
	Readonly<ReturnType<typeof createVertexSchema>>
>;

function createEdgeSchema(etConfig: EdgeTypeConfig) {
	return {
		type: etConfig.type,
		attributes: etConfig.attributes.map((attr) => ({
			name: attr.name,
			dataType: attr.dataType ?? "String",
		})),
	};
}

export type EdgeSchema = Simplify<
	Readonly<ReturnType<typeof createEdgeSchema>>
>;

function createEdgeConnectionsSchema(edgeConnections: EdgeConnection[]) {
	const byVertexType = new Map<VertexType, EdgeConnection[]>();
	const byEdgeConnectionId = new Map<EdgeConnectionId, EdgeConnection>();

	for (const conn of edgeConnections) {
		const id = createEdgeConnectionId(conn);
		byEdgeConnectionId.set(id, conn);

		// Add to source vertex type
		const sourceConns = byVertexType.get(conn.sourceVertexType) ?? [];
		sourceConns.push(conn);
		byVertexType.set(conn.sourceVertexType, sourceConns);

		// Add to target vertex type (if different from source)
		if (conn.targetVertexType !== conn.sourceVertexType) {
			const targetConns = byVertexType.get(conn.targetVertexType) ?? [];
			targetConns.push(conn);
			byVertexType.set(conn.targetVertexType, targetConns);
		}
	}

	return {
		/** Edge connections grouped by vertex type (includes both source and target) */
		byVertexType,
		/** Edge connections by their ID */
		byEdgeConnectionId,
		/** Get all connections for a specific vertex type */
		forVertexType(type: VertexType): EdgeConnection[] {
			return byVertexType.get(type) ?? [];
		},
	};
}

export type EdgeConnectionsSchema = ReturnType<
	typeof createEdgeConnectionsSchema
>;

function createGraphSchema(stored: SchemaStorageModel) {
	logger.debug("Creating graph schema", stored);
	const vertices = new Map<VertexType, VertexSchema>();
	for (const vtConfig of stored.vertices) {
		vertices.set(vtConfig.type, createVertexSchema(vtConfig));
	}

	const edges = new Map<EdgeType, EdgeSchema>();
	for (const etConfig of stored.edges) {
		edges.set(etConfig.type, createEdgeSchema(etConfig));
	}

	const edgeConnections = createEdgeConnectionsSchema(
		stored.edgeConnections ?? [],
	);

	return { vertices, edges, edgeConnections };
}

export function useGraphSchema() {
	const activeSchema = useActiveSchema();
	return createGraphSchema(activeSchema);
}

export function useVertexSchema(type: VertexType) {
	const { vertices } = useGraphSchema();
	return vertices.get(type) ?? { type, attributes: [] };
}

export function useEdgeSchema(type: EdgeType) {
	const { edges } = useGraphSchema();
	return edges.get(type) ?? { type, attributes: [] };
}

export function useVertexTypeTotal(type: VertexType) {
	const schema = useActiveSchema();
	const vertexType = schema.vertices.find((v) => v.type === type);
	return vertexType?.total;
}

export function useEdgeTypeTotal(type: EdgeType) {
	const schema = useActiveSchema();
	const edgeType = schema.edges.find((e) => e.type === type);
	return edgeType?.total;
}

export const activeSchemaSelector = atom(
	(get) => {
		const schemaMap = get(schemaAtom);
		const id = get(activeConfigurationAtom);
		const activeSchema = id ? schemaMap.get(id) : null;
		return activeSchema;
	},
	(
		get,
		set,
		update: SetStateActionWithReset<SchemaStorageModel | undefined>,
	) => {
		const schemaId = get(activeConfigurationAtom);
		if (!schemaId) {
			return;
		}
		set(schemaAtom, (prevSchemaMap) => {
			const prev = prevSchemaMap.get(schemaId);
			const newValue = typeof update === "function" ? update(prev) : update;

			if (newValue === prev) {
				return prevSchemaMap;
			}

			const updatedSchemaMap = new Map(prevSchemaMap);

			if (newValue === RESET || !newValue) {
				if (!prev) {
					return prevSchemaMap;
				}
				updatedSchemaMap.delete(schemaId);
				return updatedSchemaMap;
			}

			updatedSchemaMap.set(schemaId, newValue);

			return updatedSchemaMap;
		});
	},
);

export { createVertexTypeLookup, updateSchemaFromEntities };
export type { VertexTypeLookup };

export function mapVertexToTypeConfigs(vertex: Vertex): VertexTypeConfig[] {
	return mapVertexToSchemas(vertex) as VertexTypeConfig[];
}

export function mapEdgeToTypeConfig(edge: Edge): EdgeTypeConfig {
	return mapEdgeToSchema(edge) as EdgeTypeConfig;
}

/** Updates the schema with any new vertex or edge types and attributes. */
export function useUpdateSchemaFromEntities() {
	return useAtomCallback(
		useCallback((get, set, entities: Partial<Entities>) => {
			const vertices = entities.vertices ?? [];
			const edges = entities.edges ?? [];
			if (vertices.length === 0 && edges.length === 0) {
				return;
			}

			// Incoming entities take priority over canvas vertices
			const vertexLookup = createVertexTypeLookup(
				toNodeMap(vertices),
				get(nodesAtom),
			);

			set(activeSchemaSelector, (prev) => {
				if (!prev) {
					return prev;
				}
				return updateSchemaFromEntities(entities, prev, vertexLookup);
			});
		}, []),
	);
}

export type UpdateSchemaHandler = ReturnType<
	typeof useUpdateSchemaFromEntities
>;
