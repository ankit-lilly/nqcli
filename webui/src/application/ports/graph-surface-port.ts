export type GraphSurfaceElement = {
	data: { id: string } & Record<string, unknown>;
};

export type GraphSurfaceSelection = {
	nodeIds: ReadonlySet<string>;
	edgeIds: ReadonlySet<string>;
	groupIds: ReadonlySet<string>;
};

/** Rendering boundary implemented by Cytoscape and driven by presentation code. */
export interface GraphSurfacePort {
	mount(container: HTMLElement): void;
	destroy(): void;
	setElements(nodes: GraphSurfaceElement[], edges: GraphSurfaceElement[]): void;
	setSelection(selection: GraphSurfaceSelection): void;
	runLayout(name: string, animate?: boolean): void;
	zoomIn(): void;
	zoomOut(): void;
	fit(): void;
	onSelectionChanged(
		listener: (selection: GraphSurfaceSelection) => void,
	): () => void;
}
