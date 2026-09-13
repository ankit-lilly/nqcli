// @vitest-environment jsdom

import { CytoscapeSurfaceAdapter } from "./cytoscape-surface";
import { CytoscapeSurface } from "@nq/graph-surface";

describe("CytoscapeSurfaceAdapter", () => {
	it("owns the Cytoscape lifecycle and graph commands", () => {
		const surface = new CytoscapeSurfaceAdapter({ headless: true } as never);
		surface.mount(document.createElement("div"));
		surface.setElements(
			[{ data: { id: "a" } }, { data: { id: "b" } }],
			[{ data: { id: "ab", source: "a", target: "b" } }],
		);

		expect(surface.cytoscape?.nodes()).toHaveLength(2);
		expect(surface.cytoscape?.edges()).toHaveLength(1);

		const listener = vi.fn();
		surface.onSelectionChanged(listener);
		surface.setSelection({
			nodeIds: new Set(["a"]),
			edgeIds: new Set(),
			groupIds: new Set(),
		});
		expect(surface.cytoscape?.getElementById("a").selected()).toBe(true);

		surface.destroy();
		expect(surface.cytoscape).toBeUndefined();
	});
});

describe("shared Cytoscape surface", () => {
	it("publishes framework-neutral node and double-click events", () => {
		const onNodeClick = vi.fn();
		const onNodeDoubleClick = vi.fn();
		const surface = new CytoscapeSurface({ headless: true } as never, {
			onNodeClick,
			onNodeDoubleClick,
		});
		surface.mount(document.createElement("div"));
		surface.setElements([{ data: { id: "a", label: "Study" } }], []);

		const node = surface.cytoscape?.getElementById("a");
		node?.trigger("tap");
		node?.trigger("tap");

		expect(onNodeClick).toHaveBeenCalledTimes(2);
		expect(onNodeDoubleClick).toHaveBeenCalledOnce();
		surface.destroy();
	});
});
