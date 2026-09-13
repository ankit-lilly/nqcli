import {
	type Accessor,
	type Setter,
	onMount,
	onCleanup,
	createEffect,
	createSignal,
	untrack,
} from "solid-js";
import { CytoscapeSurface } from "@nq/graph-surface";
import cytoscape, { type Core } from "cytoscape";
import dagre from "cytoscape-dagre";
import type {
	GraphElement,
	GraphElementData,
	GraphSelection,
} from "../domain/graph";

cytoscape.use(dagre);

// Weak keys let replaced result arrays and their positions be collected.
const layoutsByResult = new WeakMap<
	GraphElement[],
	{
		positions: Record<string, { x: number; y: number }>;
		zoom: number;
		pan: { x: number; y: number };
		layout: string;
	}
>();

const LABEL_COLORS: Record<string, string> = {
	Study: "#4A90D9",
	StudyVersion: "#7B68EE",
	InterventionalStudyDesign: "#E67E22",
	ObservationalStudyDesign: "#E67E22",
	StudyEpoch: "#27AE60",
	Encounter: "#E74C3C",
	Activity: "#F39C12",
	ScheduleTimeline: "#1ABC9C",
	BiomedicalConcept: "#9B59B6",
	Code: "#95A5A6",
	AliasCode: "#7F8C8D",
	StudyIdentifier: "#2980B9",
	StudyTitle: "#8E44AD",
	Organization: "#16A085",
	Amendment: "#C0392B",
	SubjectEnrollment: "#D35400",
	ScheduledActivityInstance: "#2C3E50",
	StudyArm: "#E74C3C",
	StudyCell: "#3498DB",
	StudyElement: "#1ABC9C",
	StudyDesignPopulation: "#9B59B6",
	EligibilityCriterion: "#F1C40F",
	Timing: "#34495E",
	Condition: "#7F8C8D",
	StudySite: "#27AE60",
};

const EDGE_LABEL_CLASS = "show-edge-labels";
const LATEST_EDGE_CLASS = "latest-version-edge";

interface Props {
	elements: Accessor<GraphElement[]>;
	selectedElement: Accessor<GraphSelection>;
	setSelectedElement: Setter<GraphSelection>;
	onExpand: (id: string) => void;
	onClear: () => void;
}

function resolveHex(varName: string, fallback: string): string {
	const el = document.createElement("div");
	el.style.color = `var(${varName})`;
	el.style.display = "none";
	document.body.appendChild(el);
	const computed = getComputedStyle(el).color;
	document.body.removeChild(el);
	if (!computed || computed === "") return fallback;
	const match = computed.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
	if (!match) return fallback;
	const [, r, g, b] = match;
	return `#${Number(r).toString(16).padStart(2, "0")}${Number(g).toString(16).padStart(2, "0")}${Number(b).toString(16).padStart(2, "0")}`;
}

function mixHex(hex: string, alpha: number, bgHex: string): string {
	const parse = (h: string) => [
		parseInt(h.slice(1, 3), 16),
		parseInt(h.slice(3, 5), 16),
		parseInt(h.slice(5, 7), 16),
	];
	const fg = parse(hex);
	const bg = parse(bgHex);
	const mix = fg.map((f, i) => Math.round(f * alpha + bg[i] * (1 - alpha)));
	return `#${mix.map((c) => c.toString(16).padStart(2, "0")).join("")}`;
}

function themeColors() {
	const bg = resolveHex("--color-base-100", "#ffffff");
	const content = resolveHex("--color-base-content", "#333333");
	return {
		bg,
		content,
		edgeLine: mixHex(content, 0.2, bg),
		edgeText: mixHex(content, 0.55, bg),
		edgeOutline: bg,
		nodeText: content,
		nodeOutline: bg,
		nodeBorder: mixHex(content, 0.1, bg),
	};
}

export default function GraphView(props: Props) {
	let containerRef!: HTMLDivElement;
	let surface: CytoscapeSurface | undefined;
	let cy: Core | undefined;
	const [activeLayout, setActiveLayout] = createSignal("dagre-tb");
	const [edgeLabels, setEdgeLabels] = createSignal(false);
	const [themeKey, setThemeKey] = createSignal(0);
	let currentElements: GraphElement[] | undefined;
	let resizeFrame = 0;
	let resizeObserver: ResizeObserver | undefined;
	let activeCyLayout: ReturnType<Core["layout"]> | undefined;
	function savePositions() {
		if (!cy || !currentElements) return;
		const positions: Record<string, { x: number; y: number }> = {};
		cy.nodes().forEach((n) => {
			positions[n.id()] = { ...n.position() };
		});
		layoutsByResult.set(currentElements, {
			positions,
			zoom: cy.zoom(),
			pan: { ...cy.pan() },
			layout: untrack(activeLayout),
		});
	}

	onMount(() => {
		const observer = new MutationObserver(() => setThemeKey((k) => k + 1));
		observer.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ["data-theme"],
		});
		onCleanup(() => observer.disconnect());
	});

	const cyStyle: any[] = [
		{
			selector: "node",
			style: {
				label: "data(displayName)",
				"text-valign": "bottom",
				"text-halign": "center",
				"text-margin-y": 6,
				"font-size": "13px",
				"font-weight": 600,
				"font-family":
					"-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
				"min-zoomed-font-size": 8,
				width: 50,
				height: 50,
				"background-color": "#7f8c8d",
				color: "#333",
				"text-outline-width": 2,
				"text-outline-color": "#fff",
				"border-width": 2,
				"border-color": "#eee",
				"overlay-padding": 6,
			},
		},
		{
			selector: "edge",
			style: {
				label: "",
				"curve-style": "bezier",
				"control-point-step-size": 40,
				"target-arrow-shape": "triangle",
				"font-size": "11px",
				"font-weight": 500,
				"font-family":
					"-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
				"min-zoomed-font-size": 8,
				"text-rotation": "autorotate",
				"text-margin-y": -8,
				"text-background-opacity": 0,
				"text-background-padding": 2,
				"text-background-shape": "round-rectangle",
				"line-color": "#ccc",
				"target-arrow-color": "#ccc",
				width: 1.5,
				color: "#888",
				"text-outline-width": 3,
				"text-outline-color": "#fff",
				"text-outline-opacity": 1,
				opacity: 0.9,
			},
		},
		{
			selector: `edge.${EDGE_LABEL_CLASS}`,
			style: {
				label: "data(label)",
				"text-background-opacity": 1,
			},
		},
		{
			selector: `edge.${LATEST_EDGE_CLASS}`,
			style: {
				width: 2.75,
				"line-style": "dashed",
				"arrow-scale": 1.1,
			},
		},
		{
			selector: "node:selected",
			style: {
				"border-width": 3,
				"border-color": "#FFD700",
				"overlay-color": "#FFD700",
				"overlay-opacity": 0.08,
			},
		},
		{
			selector: "edge:selected",
			style: {
				width: 3,
				"line-color": "#E6B84A",
				"target-arrow-color": "#E6B84A",
				"overlay-color": "#E6B84A",
				"overlay-opacity": 0.08,
			},
		},
		{
			selector: "node:active",
			style: {
				"overlay-opacity": 0.04,
			},
		},
	];

	onMount(() => {
		surface = new CytoscapeSurface(
			{
				layout: { name: "preset" },
				wheelSensitivity: 0.25,
				minZoom: 0.05,
				maxZoom: 5,
				pixelRatio: "auto",
				textureOnViewport: true,
				hideEdgesOnViewport: true,
			} as any,
			{
				onNodeClick: (data) =>
					props.setSelectedElement({
						group: "nodes",
						data: data as GraphElementData,
					}),
				onEdgeClick: (data) =>
					props.setSelectedElement({
						group: "edges",
						data: data as GraphElementData,
					}),
				onCanvasClick: () => props.setSelectedElement(null),
				onNodeDoubleClick: (data) => {
					if (typeof data.id === "string") props.onExpand(data.id);
				},
			},
		);
		surface.mount(containerRef);
		cy = surface.cytoscape;
		cy?.style(cyStyle);

		resizeObserver = new ResizeObserver(() => {
			cancelAnimationFrame(resizeFrame);
			resizeFrame = requestAnimationFrame(() => cy?.resize());
		});
		resizeObserver.observe(containerRef);
	});

	onCleanup(() => {
		savePositions();
		activeCyLayout?.stop();
		resizeObserver?.disconnect();
		cancelAnimationFrame(resizeFrame);
		surface?.destroy();
		cy = undefined;
	});

	function applyThemeColors() {
		if (!cy) return;
		const tc = themeColors();
		cy.batch(() => {
			cy!.nodes().style({
				color: tc.nodeText,
				"text-outline-color": tc.nodeOutline,
				"border-color": tc.nodeBorder,
			});
			cy!.edges().style({
				"line-color": tc.edgeLine,
				"target-arrow-color": tc.edgeLine,
				color: tc.edgeText,
				"text-background-color": tc.edgeOutline,
				"text-outline-color": tc.edgeOutline,
			});
		});
	}

	function applyEdgeRendering(showLabels: boolean) {
		if (!cy) return;
		const edges = cy!.edges();
		edges.toggleClass(EDGE_LABEL_CLASS, showLabels);
		edges.forEach((edge) => {
			edge.toggleClass(
				LATEST_EDGE_CLASS,
				edge.data("label") === "has_latest_version",
			);
		});
	}

	// Theme change — only recolor, preserve layout
	createEffect(() => {
		const _theme = themeKey();
		applyThemeColors();
	});

	// Reconcile by ID and restore positions on view remount.
	createEffect(() => {
		const elems = props.elements();
		if (!cy) return;

		savePositions();
		currentElements = elems;
		activeCyLayout?.stop();
		if (!elems || elems.length === 0) {
			cy.elements().remove();
			return;
		}
		const saved = layoutsByResult.get(elems);

		const tc = themeColors();
		const showEdgeLabels = untrack(edgeLabels);

		cy.batch(() => {
			const incoming = new Set(elems.map((el) => String(el.data.id)));
			cy!
				.elements()
				.filter((el) => !incoming.has(el.id()))
				.remove();
			for (const el of elems) {
				const data = {
					...el.data,
					displayName: String(
						el.data.name || el.data.decode || el.data.label || "",
					).slice(0, 60),
				};
				const existing = cy!.getElementById(String(el.data.id));
				if (existing.length) existing.data(data);
				else cy!.add({ group: el.group, data } as any);
			}

			// All styling in one batch: label colors, topology, theme
			const labelColorMap = new Map(Object.entries(LABEL_COLORS));
			cy!.nodes().forEach((node) => {
				const lbl = node.data("label") as string;
				const bg = labelColorMap.get(lbl) || "#7f8c8d";

				const inDeg = node.indegree(false);
				const outDeg = node.outdegree(false);
				const total = inDeg + outDeg;

				let shape = "ellipse";
				let size = Math.min(65, 38 + total * 3);
				let fontSize = "13px";
				let fontWeight = 600;
				let borderWidth = 2;
				let textMarginY = 6;

				if (inDeg === 0 && outDeg > 0) {
					shape = "diamond";
					size = Math.min(80, 55 + outDeg * 4);
					fontSize = "15px";
					fontWeight = 700;
					borderWidth = 3;
					textMarginY = 8;
				} else if (outDeg === 0) {
					size = 30;
					fontSize = "11px";
					fontWeight = 400;
					borderWidth = 1;
					textMarginY = 4;
				}

				node.style({
					"background-color": bg,
					shape,
					width: size,
					height: size,
					"font-size": fontSize,
					"font-weight": fontWeight,
					"border-width": borderWidth,
					"text-margin-y": textMarginY,
					color: tc.nodeText,
					"text-outline-color": tc.nodeOutline,
					"border-color": tc.nodeBorder,
				});
			});

			cy!.edges().style({
				"line-color": tc.edgeLine,
				"target-arrow-color": tc.edgeLine,
				color: tc.edgeText,
				"text-background-color": tc.edgeOutline,
				"text-outline-color": tc.edgeOutline,
			});
			applyEdgeRendering(showEdgeLabels);
		});

		if (saved) {
			cy.nodes().positions(
				(node) => saved.positions[node.id()] || { x: 0, y: 0 },
			);
			cy.zoom(saved.zoom);
			cy.pan(saved.pan);
			setActiveLayout(saved.layout);
		} else {
			const small = cy.nodes().length <= 150;
			activeCyLayout = cy.layout({
				name: small ? "dagre" : "grid",
				rankDir: "TB",
				nodeSep: 60,
				rankSep: 80,
				fit: true,
				padding: 40,
				animate: false,
			} as any);
			activeCyLayout.run();
			setActiveLayout(small ? "dagre-tb" : "grid");
		}
	});

	function runLayout(
		key: string,
		name: string,
		opts: Record<string, any> = {},
	) {
		if (!cy) return;
		setActiveLayout(key);
		const count = cy.elements().length;
		activeCyLayout?.stop();
		activeCyLayout = cy.layout({
			name,
			fit: true,
			padding: 40,
			animate: count < 100,
			animationDuration: 250,
			animationEasing: "ease-out",
			...opts,
		} as any);
		activeCyLayout.run();
	}

	createEffect(() => {
		applyEdgeRendering(edgeLabels());
	});

	createEffect(() => {
		const selected = props.selectedElement();
		surface?.setSelection({
			nodeIds: new Set(selected?.group === "nodes" ? [selected.data.id] : []),
			edgeIds: new Set(selected?.group === "edges" ? [selected.data.id] : []),
			groupIds: new Set(),
		});
	});
	const layouts = [
		{
			key: "dagre-tb",
			label: "Top-Down",
			name: "dagre",
			opts: { rankDir: "TB" },
		},
		{
			key: "dagre-lr",
			label: "Left-Right",
			name: "dagre",
			opts: { rankDir: "LR" },
		},
		{
			key: "cose",
			label: "Force",
			name: "cose",
			opts: { nodeRepulsion: () => 4500, idealEdgeLength: () => 80 },
		},
		{ key: "grid", label: "Grid", name: "grid", opts: {} },
		{ key: "circle", label: "Circle", name: "circle", opts: {} },
	];

	return (
		<div class="graph-container">
			<div class="flex items-center gap-1.5 px-3 py-2 border-b border-base-300">
				{layouts.map((l) => (
					<button
						onClick={() => runLayout(l.key, l.name, l.opts)}
						class={`btn btn-xs ${activeLayout() === l.key ? "btn-primary" : "btn-ghost"}`}
					>
						{l.label}
					</button>
				))}
				<label class="text-xs flex gap-1 items-center">
					<input
						type="checkbox"
						checked={edgeLabels()}
						onChange={(e) => setEdgeLabels(e.currentTarget.checked)}
					/>
					Edge labels
				</label>
				<div class="ml-auto" />
				<button
					onClick={() => surface?.zoomOut()}
					class="btn btn-xs btn-ghost btn-square"
					title="Zoom out"
					aria-label="Zoom out"
				>
					−
				</button>
				<button
					onClick={() => surface?.zoomIn()}
					class="btn btn-xs btn-ghost btn-square"
					title="Zoom in"
					aria-label="Zoom in"
				>
					+
				</button>
				<button
					onClick={() => surface?.fit(40)}
					class="btn btn-xs btn-ghost"
					title="Zoom to fit all elements"
				>
					Fit
				</button>
				<button
					onClick={() => {
						const blob = surface?.exportPNG();
						if (!blob) return;
						const url = URL.createObjectURL(blob);
						const anchor = document.createElement("a");
						anchor.href = url;
						anchor.download = "nq-graph.png";
						anchor.click();
						URL.revokeObjectURL(url);
					}}
					class="btn btn-xs btn-ghost"
					title="Export graph as PNG"
				>
					Export
				</button>
				<button
					onClick={props.onClear}
					class="btn btn-xs btn-ghost text-error"
					title="Clear graph"
				>
					Clear
				</button>
			</div>
			<div ref={containerRef!} class="cytoscape-canvas" />
		</div>
	);
}
