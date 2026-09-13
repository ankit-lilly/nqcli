import type { GraphEntities } from "@/domain";

import type { GraphWorkspacePort } from "../ports";
import { createVertexTypeLookup } from "../workspace";

export class AddEntitiesToWorkspace {
	constructor(private readonly workspace: GraphWorkspacePort) {}

	async execute(entities: Partial<GraphEntities>): Promise<void> {
		const vertices = entities.vertices ?? [];
		const edges = entities.edges ?? [];
		if (vertices.length === 0 && edges.length === 0) {
			return;
		}

		const incomingVertices = new Map(
			vertices.map((vertex) => [vertex.id, vertex]),
		);
		const lookup = createVertexTypeLookup(
			incomingVertices,
			this.workspace.vertices(),
		);

		this.workspace.add(entities);
		this.workspace.mergeSchema(entities, lookup);
		await this.workspace.persist();
	}
}
