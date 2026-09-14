import { DesktopService } from "../../bindings/github.com/ankit-lilly/nqcli/internal/desktop";
import type * as WailsSchema from "../../bindings/github.com/ankit-lilly/nqcli/internal/schema/models";
// @ts-expect-error Wails injects this virtual module into the webview.
import { Events } from "/wails/runtime.js";
import type { SchemaPort } from "../application/schema-explorer";
import type {
	SchemaEvent,
	SchemaSnapshot,
	SchemaStatus,
} from "../domain/schema";

const statuses = new Set<SchemaStatus>(["empty", "running", "ready", "failed"]);
const schemaEventName = "schema:update";

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

	subscribe(listener: (event: SchemaEvent) => void): () => void {
		return Events.On(schemaEventName, (event: { data: WailsSchema.Event }) => {
			const data = event.data as Partial<SchemaEvent> | null;
			if (
				!data ||
				typeof data.type !== "string" ||
				typeof data.key !== "string" ||
				!statuses.has(data.status as SchemaStatus)
			) {
				return;
			}
			listener(data as SchemaEvent);
		});
	}
}
