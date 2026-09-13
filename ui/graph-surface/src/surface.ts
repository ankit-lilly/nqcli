import cytoscape, { type Core, type ElementDefinition } from "cytoscape";

export type GraphElement = ElementDefinition & {
	data: { id: string; [key: string]: unknown };
};

export type GraphSelection = {
	nodeIds: Set<string>;
	edgeIds: Set<string>;
	groupIds: Set<string>;
};

export type SurfaceConfig = cytoscape.CytoscapeOptions;

export type SurfaceCallbacks = {
	onLayoutRunningChanged?: (running: boolean) => void;
	onZoomChanged?: (zoom: number) => void;
	onPanChanged?: (pan: { x: number; y: number }) => void;
	onNodeClick?: (data: Record<string, unknown>) => void;
	onNodeDoubleClick?: (data: Record<string, unknown>) => void;
	onEdgeClick?: (data: Record<string, unknown>) => void;
	onCanvasClick?: () => void;
};

const doubleClickDelay = 300;

/** Framework-neutral Cytoscape lifecycle, events, and viewport commands. */
export class CytoscapeSurface {
	#cy?: Core;
	#callbacks: SurfaceCallbacks;
	#selectionListeners = new Set<(selection: GraphSelection) => void>();
	#zoomTimer?: ReturnType<typeof setTimeout>;
	#panTimer?: ReturnType<typeof setTimeout>;
	#tapTimer?: ReturnType<typeof setTimeout>;
	#lastTapped?: cytoscape.SingularElementReturnValue;

	constructor(
		private readonly config: Partial<SurfaceConfig> = {},
		callbacks: SurfaceCallbacks = {},
	) {
		this.#callbacks = callbacks;
	}

	get cytoscape(): Core | undefined {
		return this.#cy;
	}

	setCallbacks(callbacks: SurfaceCallbacks): void {
		this.#callbacks = callbacks;
	}

	mount(container: HTMLElement): void {
		this.destroy();
		const cy = cytoscape({
			container,
			style: [],
			...this.config,
		} as SurfaceConfig);

		cy.on("layoutstart", () => {
			cy.userPanningEnabled(false);
			cy.userZoomingEnabled(false);
			this.#callbacks.onLayoutRunningChanged?.(true);
		});
		cy.on("layoutstop", () => {
			cy.userPanningEnabled(this.config.userPanningEnabled ?? true);
			cy.userZoomingEnabled(this.config.userZoomingEnabled ?? true);
			if (this.config.autolock) cy.nodes().lock();
			this.#callbacks.onLayoutRunningChanged?.(false);
		});
		cy.on("zoom", () => {
			clearTimeout(this.#zoomTimer);
			this.#zoomTimer = setTimeout(
				() => this.#callbacks.onZoomChanged?.(cy.zoom()),
				100,
			);
		});
		cy.on("pan", () => {
			clearTimeout(this.#panTimer);
			this.#panTimer = setTimeout(
				() => this.#callbacks.onPanChanged?.(cy.pan()),
				100,
			);
		});
		cy.on("select unselect", () => this.#publishSelection());
		cy.on("tap", (event) => this.#handleTap(event));
		cy.on("tap", "node", (event) =>
			this.#callbacks.onNodeClick?.(event.target.data()),
		);
		cy.on("tap", "edge", (event) =>
			this.#callbacks.onEdgeClick?.(event.target.data()),
		);
		cy.on("tap", (event) => {
			if (event.target === cy) this.#callbacks.onCanvasClick?.();
		});
		cy.on("doubleTap", "node", (event) =>
			this.#callbacks.onNodeDoubleClick?.(event.target.data()),
		);

		this.#cy = cy;
	}

	destroy(): void {
		clearTimeout(this.#zoomTimer);
		clearTimeout(this.#panTimer);
		clearTimeout(this.#tapTimer);
		this.#lastTapped = undefined;
		if (!this.#cy) return;
		this.#cy.removeAllListeners();
		this.#cy.destroy();
		this.#cy = undefined;
	}

	setElements(nodes: GraphElement[], edges: GraphElement[]): void {
		this.#cy?.json({
			elements: structuredClone({ nodes, edges }),
		});
	}

	setSelection(selection: GraphSelection): void {
		const cy = this.#cy;
		if (!cy) return;
		const selected = new Set([
			...selection.nodeIds,
			...selection.edgeIds,
			...selection.groupIds,
		]);
		cy.batch(() => {
			cy.elements(":selected").forEach((element) => {
				if (!selected.has(element.id())) element.unselect();
			});
			for (const id of selected) cy.getElementById(id).select();
		});
	}

	applyLayout(options: cytoscape.LayoutOptions): void {
		this.#cy?.layout(options).run();
	}

	zoomIn(step = 0.15): void {
		if (this.#cy) this.#cy.zoom(this.#cy.zoom() + step);
	}

	zoomOut(step = 0.15): void {
		if (this.#cy) this.#cy.zoom(this.#cy.zoom() - step);
	}

	fit(padding = 30): void {
		this.#cy?.fit(this.#cy.elements(), padding);
	}

	center(): void {
		this.#cy?.center(this.#cy.elements());
	}

	exportPNG(): Blob | undefined {
		return this.#cy?.png({ output: "blob", full: true }) as Blob | undefined;
	}

	onSelectionChanged(
		listener: (selection: GraphSelection) => void,
	): () => void {
		this.#selectionListeners.add(listener);
		return () => this.#selectionListeners.delete(listener);
	}

	#handleTap(event: cytoscape.EventObject): void {
		const tapped = event.target as cytoscape.SingularElementReturnValue;
		if (this.#lastTapped === tapped) {
			clearTimeout(this.#tapTimer);
			this.#lastTapped = undefined;
			tapped.trigger("doubleTap", [event]);
			return;
		}
		this.#lastTapped = tapped;
		clearTimeout(this.#tapTimer);
		this.#tapTimer = setTimeout(() => {
			this.#lastTapped = undefined;
		}, doubleClickDelay);
	}

	#publishSelection(): void {
		const cy = this.#cy;
		if (!cy || this.#selectionListeners.size === 0) return;
		const selection = {
			nodeIds: new Set(
				cy.$("node:selected[!__isGroupNode]").map((e) => e.id()),
			),
			edgeIds: new Set(cy.$("edge:selected").map((e) => e.id())),
			groupIds: new Set(
				cy.$("node:selected[?__isGroupNode]").map((e) => e.id()),
			),
		};
		for (const listener of this.#selectionListeners) listener(selection);
	}
}
