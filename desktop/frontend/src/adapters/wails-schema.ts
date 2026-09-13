import { DesktopService } from "../../bindings/github.com/ankit-lilly/nqcli/internal/desktop";
import type * as WailsSchema from "../../bindings/github.com/ankit-lilly/nqcli/internal/schema/models";
import type { SchemaPort } from "../application/schema-explorer";
import type { SchemaSnapshot, SchemaStatus } from "../domain/schema";

const statuses = new Set<SchemaStatus>(["empty", "running", "ready", "failed"]);

export class WailsSchemaAdapter implements SchemaPort {
	async get(options: { refresh?: boolean } = {}): Promise<SchemaSnapshot> {
		const result = await DesktopService.GetSchema({
			refresh: options.refresh ?? false,
		});
		const status = String(result.status) as SchemaStatus;
		return {
			status: statuses.has(status) ? status : "empty",
			phase: result.phase,
			completed: result.completed,
			total: result.total,
			error: result.error,
			lastUpdate: result.lastUpdate ?? undefined,
			totalVertices: result.totalVertices ?? undefined,
			vertices: (result.vertices ?? []).map((vertex: WailsSchema.Vertex) => ({
				type: vertex.type,
				total: vertex.total ?? undefined,
				attributes: vertex.attributes ?? [],
			})),
			totalEdges: result.totalEdges ?? undefined,
			edges: (result.edges ?? []).map((edge: WailsSchema.Edge) => ({
				type: edge.type,
				total: edge.total ?? undefined,
				attributes: edge.attributes ?? [],
			})),
			edgeConnections: result.edgeConnections ?? [],
		};
	}
}
