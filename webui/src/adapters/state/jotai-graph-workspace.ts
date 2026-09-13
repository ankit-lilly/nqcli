import type { GraphWorkspacePort } from "@/application";
import { updateSchemaFromEntities, type VertexTypeLookup } from "@/application";
import {
	type AppStore,
	activeSchemaSelector,
	edgesAtom,
	nodesAtom,
	toEdgeMap,
	toNodeMap,
} from "@/core";
import type { GraphEntities } from "@/domain";
import { logger } from "@/utils";

export class JotaiGraphWorkspaceAdapter implements GraphWorkspacePort {
	constructor(
		private readonly store: AppStore,
		private readonly persistSession: () => void | Promise<void>,
	) {}

	vertices() {
		return this.store.get(nodesAtom);
	}

	add(entities: Partial<GraphEntities>): void {
		const vertices = toNodeMap(entities.vertices ?? []);
		const edges = toEdgeMap(entities.edges ?? []);

		if (vertices.size > 0) {
			logger.debug("Adding vertices to graph", vertices);
			this.store.set(
				nodesAtom,
				(previous) => new Map([...previous, ...vertices]),
			);
		}
		if (edges.size > 0) {
			logger.debug("Adding edges to graph", edges);
			this.store.set(edgesAtom, (previous) => new Map([...previous, ...edges]));
		}
	}

	mergeSchema(
		entities: Partial<GraphEntities>,
		vertexLookup: VertexTypeLookup,
	): void {
		this.store.set(activeSchemaSelector, (previous) =>
			previous
				? updateSchemaFromEntities(entities, previous, vertexLookup)
				: previous,
		);
	}

	persist(): void | Promise<void> {
		return this.persistSession();
	}
}
