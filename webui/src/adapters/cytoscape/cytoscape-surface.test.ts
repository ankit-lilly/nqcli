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
	it("does not publish a single click as part of a double click", () => {
		vi.useFakeTimers();
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
		vi.runAllTimers();

		expect(onNodeClick).not.toHaveBeenCalled();
		expect(onNodeDoubleClick).toHaveBeenCalledOnce();
		surface.destroy();
		vi.useRealTimers();
	});

	it("publishes a single node click after the double-click window", () => {
		vi.useFakeTimers();
		const onNodeClick = vi.fn();
		const surface = new CytoscapeSurface({ headless: true } as never, {
			onNodeClick,
		});
		surface.mount(document.createElement("div"));
		surface.setElements([{ data: { id: "a", label: "Study" } }], []);

		surface.cytoscape?.getElementById("a").trigger("tap");
		vi.advanceTimersByTime(300);

		expect(onNodeClick).toHaveBeenCalledOnce();
		surface.destroy();
		vi.useRealTimers();
	});
});
