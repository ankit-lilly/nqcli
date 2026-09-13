import { CytoscapeSurface, type SurfaceCallbacks } from "@nq/graph-surface";
import cytoscape from "cytoscape";
import cyCanvas from "cytoscape-canvas";
import d3Force from "cytoscape-d3-force";
import dagre from "cytoscape-dagre";
import fcose from "cytoscape-fcose";
import klay from "cytoscape-klay";

import type {
	GraphSurfaceElement,
	GraphSurfacePort,
	GraphSurfaceSelection,
} from "@/application";
import { runLayout } from "./layout";
import type { Config, CytoscapeType, LayoutName } from "./model";

cytoscape.use(klay);
cytoscape.use(dagre);
cytoscape.use(d3Force);
cytoscape.use(fcose);
cyCanvas(cytoscape);

export type CytoscapeSurfaceCallbacks = Pick<
	SurfaceCallbacks,
	"onLayoutRunningChanged" | "onZoomChanged" | "onPanChanged"
>;

/** NQ's Cytoscape plugins and layouts over the shared framework-neutral surface. */
export class CytoscapeSurfaceAdapter
	extends CytoscapeSurface
	implements GraphSurfacePort
{
	constructor(config: Config = {}, callbacks: CytoscapeSurfaceCallbacks = {}) {
		super(config, callbacks);
	}

	override get cytoscape(): CytoscapeType | undefined {
		return super.cytoscape as CytoscapeType | undefined;
	}

	override setElements(
		nodes: GraphSurfaceElement[],
		edges: GraphSurfaceElement[],
	): void {
		super.setElements(nodes, edges);
	}

	override setSelection(selection: GraphSurfaceSelection): void {
		super.setSelection({
			nodeIds: new Set(selection.nodeIds),
			edgeIds: new Set(selection.edgeIds),
			groupIds: new Set(selection.groupIds),
		});
	}

	runLayout(name: string, animate = true): void {
		if (this.cytoscape) {
			runLayout(this.cytoscape, name as LayoutName, {}, animate);
		}
	}
}
