import { createEdge, createVertex } from "@/domain";

import type { GraphWorkspacePort } from "../ports";
import type { VertexTypeLookup } from "../workspace";

import { AddEntitiesToWorkspace } from "./add-entities-to-workspace";

describe("AddEntitiesToWorkspace", () => {
	it("coordinates graph, schema, and persistence without framework state", async () => {
		const existing = createVertex({ id: "existing", types: ["Study"] });
		const incoming = createVertex({ id: "incoming", types: ["StudyVersion"] });
		const edge = createEdge({
			id: "edge",
			type: "has_version",
			sourceId: existing.id,
			targetId: incoming.id,
		});
		let lookup: VertexTypeLookup | undefined;
		const workspace = {
			vertices: vi.fn(() => new Map([[existing.id, existing]])),
			add: vi.fn(),
			mergeSchema: vi.fn((_entities, value) => {
				lookup = value;
			}),
			persist: vi.fn(),
		} satisfies GraphWorkspacePort;

		await new AddEntitiesToWorkspace(workspace).execute({
			vertices: [incoming],
			edges: [edge],
		});

		expect(workspace.add).toHaveBeenCalledOnce();
		expect(workspace.mergeSchema).toHaveBeenCalledOnce();
		expect(workspace.persist).toHaveBeenCalledOnce();
		expect(lookup?.get(incoming.id)).toEqual(incoming.types);
		expect(lookup?.get(existing.id)).toEqual(existing.types);
	});

	it("does nothing for an empty result", async () => {
		const workspace = {
			vertices: vi.fn(() => new Map()),
			add: vi.fn(),
			mergeSchema: vi.fn(),
			persist: vi.fn(),
		} satisfies GraphWorkspacePort;

		await new AddEntitiesToWorkspace(workspace).execute({});

		expect(workspace.add).not.toHaveBeenCalled();
		expect(workspace.persist).not.toHaveBeenCalled();
	});
});
