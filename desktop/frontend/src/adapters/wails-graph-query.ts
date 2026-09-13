import { DesktopService } from "../../bindings/github.com/ankit-lilly/nqcli/internal/desktop";
import type {
	ExpandVertexCommand,
	GraphQueryPort,
	NeighborSummaryCommand,
	RunGraphQuery,
} from "../application/graph-explorer";
import { toGraphElements } from "../domain/graph";

export class WailsGraphQueryAdapter implements GraphQueryPort {
	async run(command: RunGraphQuery) {
		const response = await DesktopService.ExecuteGraphQuery({
			query: command.query,
			type: command.type,
			serializer: "",
		});
		if (response.error) throw new Error(response.error);
		return {
			elements: toGraphElements(response.elements),
			json: response.json ?? "",
			warning: response.warning,
		};
	}

	async expand(command: ExpandVertexCommand) {
		const response = await DesktopService.ExpandVertex({
			id: command.id,
			type: command.type,
			direction: command.direction ?? "both",
			relationship: command.relationship ?? "",
			neighborLabel: command.neighborLabel ?? "",
			limit: command.limit ?? 10,
			excludedVertexIds: command.excludedVertexIds ?? [],
		});
		if (response.error) throw new Error(response.error);
		return {
			elements: toGraphElements(response.elements),
			json: "",
			warning: response.warning,
		};
	}

	async neighborSummary(command: NeighborSummaryCommand) {
		const response = await DesktopService.GetNeighborSummary({
			id: command.id,
			type: command.type,
			direction: command.direction ?? "both",
		});
		if (response.error) throw new Error(response.error);
		return {
			nodes: response.nodes ?? [],
			relationships: response.relationships ?? [],
		};
	}

	async vertexProperties(id: string) {
		const response = await DesktopService.GetVertexProperties({ id });
		if (response.error) throw new Error(response.error);
		return response.properties ?? null;
	}
}
