import { CytoscapeSurface } from "@nq/graph-surface";
import cytoscape, { type Core } from "cytoscape";
import dagre from "cytoscape-dagre";
import {
	Download,
	Maximize2,
	Minimize2,
	RotateCcw,
	Scan,
	SlidersHorizontal,
	Trash2,
	ZoomIn,
	ZoomOut,
} from "lucide-solid";
import {
	For,
	Show,
	type Accessor,
	type Setter,
	createEffect,
	createMemo,
	createSignal,
	onCleanup,
	onMount,
	untrack,
} from "solid-js";
import { createStore } from "solid-js/store";
import type {
	GraphElement,
	GraphElementData,
	GraphSelection,
} from "../domain/graph";
import { DEFAULT_NODE_COLOR, NODE_LABEL_COLORS } from "../domain/graph-style";

cytoscape.use(dagre);

type LayoutState = {
	positions: Record<string, { x: number; y: number }>;
	zoom: number;
	pan: { x: number; y: number };
	layout: string;
};

const layoutsByResult = new WeakMap<GraphElement[], LayoutState>();

const NODE_LABEL_CLASS = "show-node-labels";
const EDGE_LABEL_CLASS = "show-edge-labels";
const LATEST_EDGE_CLASS = "latest-version-edge";

const layouts = [
	{
		key: "dagre-tb",
		label: "Top-down",
		name: "dagre",
		opts: { rankDir: "TB" },
	},
	{
		key: "dagre-lr",
		label: "Left-right",
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

interface Props {
	elements: Accessor<GraphElement[]>;
	selectedElement: Accessor<GraphSelection>;
	setSelectedElement: Setter<GraphSelection>;
	onExpand?: (id: string) => void;
	onClear?: () => void;
	emptyMessage?: string;
}

function resolveHex(varName: string, fallback: string): string {
	const element = document.createElement("div");
	element.style.color = `var(${varName})`;
	element.style.display = "none";
	document.body.appendChild(element);
	const computed = getComputedStyle(element).color;
	element.remove();
	const match = computed.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
	if (!match) return fallback;
	return `#${match
		.slice(1, 4)
		.map((part) => Number(part).toString(16).padStart(2, "0"))
		.join("")}`;
}

function mixHex(hex: string, alpha: number, background: string): string {
	const parse = (value: string) => [
		Number.parseInt(value.slice(1, 3), 16),
		Number.parseInt(value.slice(3, 5), 16),
		Number.parseInt(value.slice(5, 7), 16),
	];
	const foreground = parse(hex);
	const bg = parse(background);
	return `#${foreground
		.map((value, index) =>
			Math.round(value * alpha + bg[index] * (1 - alpha))
				.toString(16)
				.padStart(2, "0"),
		)
		.join("")}`;
}

function themeColors() {
	const background = resolveHex("--color-base-100", "#ffffff");
	const content = resolveHex("--color-base-content", "#333333");
	return {
		background,
		content,
		edgeLine: mixHex(content, 0.24, background),
		edgeText: mixHex(content, 0.62, background),
		nodeBorder: mixHex(content, 0.14, background),
	};
}

export default function GraphView(props: Props) {
	let containerRef!: HTMLDivElement;
	let surface: CytoscapeSurface | undefined;
	let cy: Core | undefined;
	let currentElements: GraphElement[] | undefined;
	let activeCyLayout: ReturnType<Core["layout"]> | undefined;
	let resizeObserver: ResizeObserver | undefined;
	let resizeFrame = 0;
	const [activeLayout, setActiveLayout] = createSignal("dagre-tb");
	const [fullscreen, setFullscreen] = createSignal(false);
	const [themeKey, setThemeKey] = createSignal(0);
	const [rendering, setRendering] = createStore({
		nodeLabels: true,
		edgeLabels: false,
		topologySizing: true,
		edgeCurve: "bezier" as "bezier" | "straight" | "taxi",
	});

	const legend = createMemo(() => {
		const labels = new Set<string>();
		for (const element of props.elements()) {
			if (element.group === "nodes" && typeof element.data.label === "string") {
				labels.add(element.data.label);
			}
		}
		return [...labels].sort();
	});

	function savePositions() {
		if (!cy || !currentElements) return;
		const positions: LayoutState["positions"] = {};
		cy.nodes().forEach((node) => {
			positions[node.id()] = { ...node.position() };
		});
		layoutsByResult.set(currentElements, {
			positions,
			zoom: cy.zoom(),
			pan: { ...cy.pan() },
			layout: untrack(activeLayout),
		});
	}

	const cyStyle: any[] = [
		{
			selector: "node",
			style: {
				label: "",
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
				"background-color": "#64748b",
				"text-outline-width": 2,
				"border-width": 2,
				"overlay-padding": 6,
			},
		},
		{
			selector: `node.${NODE_LABEL_CLASS}`,
			style: { label: "data(displayName)" },
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
				"text-outline-width": 3,
				width: 1.5,
				opacity: 0.9,
			},
		},
		{
			selector: `edge.${EDGE_LABEL_CLASS}`,
			style: { label: "data(label)", "text-background-opacity": 1 },
		},
		{
			selector: `edge.${LATEST_EDGE_CLASS}`,
			style: { width: 2.75, "line-style": "dashed", "arrow-scale": 1.1 },
		},
		{
			selector: "node:selected",
			style: {
				"border-width": 3,
				"border-color": "#f5b82e",
				"overlay-color": "#f5b82e",
				"overlay-opacity": 0.1,
			},
		},
		{
			selector: "edge:selected",
			style: {
				width: 3,
				"line-color": "#e6a817",
				"target-arrow-color": "#e6a817",
				"overlay-color": "#e6a817",
				"overlay-opacity": 0.1,
			},
		},
	];

	onMount(() => {
		const themeObserver = new MutationObserver(() =>
			setThemeKey((key) => key + 1),
		);
		themeObserver.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ["data-theme"],
		});
		const onKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") setFullscreen(false);
		};
		window.addEventListener("keydown", onKeyDown);

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
					if (props.onExpand && typeof data.id === "string")
						props.onExpand(data.id);
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

		onCleanup(() => {
			savePositions();
			activeCyLayout?.stop();
			resizeObserver?.disconnect();
			cancelAnimationFrame(resizeFrame);
			themeObserver.disconnect();
			window.removeEventListener("keydown", onKeyDown);
			surface?.destroy();
			cy = undefined;
		});
	});

	function applyRendering() {
		if (!cy) return;
		const colors = themeColors();
		cy.batch(() => {
			cy!.nodes().forEach((node) => {
				const inDegree = node.indegree(false);
				const outDegree = node.outdegree(false);
				const degree = inDegree + outDegree;
				const emphasizeRoot =
					rendering.topologySizing && inDegree === 0 && outDegree > 0;
				const emphasizeLeaf = rendering.topologySizing && outDegree === 0;
				const size = emphasizeRoot
					? Math.min(80, 55 + outDegree * 4)
					: emphasizeLeaf
						? 30
						: rendering.topologySizing
							? Math.min(65, 38 + degree * 3)
							: 48;
				node.toggleClass(NODE_LABEL_CLASS, rendering.nodeLabels);
				node.style({
					"background-color":
						NODE_LABEL_COLORS[String(node.data("label"))] ?? DEFAULT_NODE_COLOR,
					shape: emphasizeRoot ? "diamond" : "ellipse",
					width: size,
					height: size,
					"font-size": emphasizeRoot ? "15px" : emphasizeLeaf ? "11px" : "13px",
					"font-weight": emphasizeRoot ? 700 : emphasizeLeaf ? 400 : 600,
					"border-width": node.selected()
						? 3
						: emphasizeRoot
							? 3
							: emphasizeLeaf
								? 1
								: 2,
					color: colors.content,
					"text-outline-color": colors.background,
					"border-color": node.selected() ? "#f5b82e" : colors.nodeBorder,
				});
			});
			cy!.edges().forEach((edge) => {
				edge.toggleClass(EDGE_LABEL_CLASS, rendering.edgeLabels);
				edge.toggleClass(
					LATEST_EDGE_CLASS,
					edge.data("label") === "has_latest_version",
				);
				const lineColor = edge.selected() ? "#e6a817" : colors.edgeLine;
				edge.style({
					"curve-style": rendering.edgeCurve,
					"line-color": lineColor,
					"target-arrow-color": lineColor,
					width: edge.selected()
						? 3
						: edge.data("label") === "has_latest_version"
							? 2.75
							: 1.5,
					color: colors.edgeText,
					"text-background-color": colors.background,
					"text-outline-color": colors.background,
				});
			});
		});
	}

	createEffect(() => {
		themeKey();
		applyRendering();
	});

	createEffect(() => {
		const elements = props.elements();
		if (!cy) return;
		savePositions();
		currentElements = elements;
		activeCyLayout?.stop();
		if (elements.length === 0) {
			cy.elements().remove();
			return;
		}
		const saved = layoutsByResult.get(elements);
		cy.batch(() => {
			const incoming = new Set(elements.map((element) => element.data.id));
			cy!
				.elements()
				.filter((element) => !incoming.has(element.id()))
				.remove();
			for (const element of elements) {
				const data = {
					...element.data,
					displayName: String(
						element.data.name ||
							element.data.decode ||
							element.data.label ||
							"",
					).slice(0, 60),
				};
				const existing = cy!.getElementById(element.data.id);
				if (existing.length) existing.data(data);
				else cy!.add({ group: element.group, data } as any);
			}
		});
		applyRendering();
		if (saved) {
			cy.nodes().positions(
				(node) => saved.positions[node.id()] ?? { x: 0, y: 0 },
			);
			cy.zoom(saved.zoom);
			cy.pan(saved.pan);
			setActiveLayout(saved.layout);
		} else {
			runLayout(cy.nodes().length <= 150 ? "dagre-tb" : "grid", false);
		}
	});

	createEffect(() => {
		const selected = props.selectedElement();
		surface?.setSelection({
			nodeIds: new Set(selected?.group === "nodes" ? [selected.data.id] : []),
			edgeIds: new Set(selected?.group === "edges" ? [selected.data.id] : []),
			groupIds: new Set(),
		});
		applyRendering();
	});

	function runLayout(key = activeLayout(), animate = true) {
		if (!cy) return;
		const layout =
			layouts.find((candidate) => candidate.key === key) ?? layouts[0];
		setActiveLayout(layout.key);
		activeCyLayout?.stop();
		activeCyLayout = cy.layout({
			name: layout.name,
			fit: true,
			padding: 40,
			animate: animate && cy.elements().length < 100,
			animationDuration: 250,
			animationEasing: "ease-out",
			...layout.opts,
		} as any);
		activeCyLayout.run();
	}

	function exportPNG() {
		const blob = surface?.exportPNG();
		if (!blob) return;
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement("a");
		anchor.href = url;
		anchor.download = "nq-graph.png";
		anchor.click();
		URL.revokeObjectURL(url);
	}

	const iconButton = "btn btn-xs btn-ghost btn-square";
	return (
		<div
			class={`graph-container bg-base-100 ${fullscreen() ? "fixed inset-0 z-50" : ""}`}
		>
			<div class="flex items-center gap-1 px-3 py-2 border-b border-base-300 min-h-11">
				<select
					class="select select-bordered select-xs w-32"
					value={activeLayout()}
					onChange={(event) => runLayout(event.currentTarget.value)}
					aria-label="Graph layout"
				>
					<For each={layouts}>
						{(layout) => <option value={layout.key}>{layout.label}</option>}
					</For>
				</select>
				<button
					class={iconButton}
					title="Re-run layout"
					aria-label="Re-run layout"
					onClick={() => runLayout()}
				>
					<RotateCcw size={15} />
				</button>
				<details class="dropdown">
					<summary
						class={iconButton}
						title="Rendering options"
						aria-label="Rendering options"
					>
						<SlidersHorizontal size={15} />
					</summary>
					<div class="dropdown-content z-30 mt-2 w-56 border border-base-300 bg-base-100 p-3 shadow-lg">
						<div class="text-xs font-semibold mb-2">Rendering</div>
						<label class="flex items-center justify-between py-1 text-xs">
							<span>Node labels</span>
							<input
								type="checkbox"
								class="toggle toggle-xs"
								checked={rendering.nodeLabels}
								onChange={(event) =>
									setRendering("nodeLabels", event.currentTarget.checked)
								}
							/>
						</label>
						<label class="flex items-center justify-between py-1 text-xs">
							<span>Edge labels</span>
							<input
								type="checkbox"
								class="toggle toggle-xs"
								checked={rendering.edgeLabels}
								onChange={(event) =>
									setRendering("edgeLabels", event.currentTarget.checked)
								}
							/>
						</label>
						<label class="flex items-center justify-between py-1 text-xs">
							<span>Size by topology</span>
							<input
								type="checkbox"
								class="toggle toggle-xs"
								checked={rendering.topologySizing}
								onChange={(event) =>
									setRendering("topologySizing", event.currentTarget.checked)
								}
							/>
						</label>
						<label class="block pt-2 text-xs">
							<span class="block mb-1 text-base-content/60">Edge routing</span>
							<select
								class="select select-bordered select-xs w-full"
								value={rendering.edgeCurve}
								onChange={(event) =>
									setRendering(
										"edgeCurve",
										event.currentTarget.value as typeof rendering.edgeCurve,
									)
								}
							>
								<option value="bezier">Curved</option>
								<option value="straight">Straight</option>
								<option value="taxi">Orthogonal</option>
							</select>
						</label>
					</div>
				</details>
				<details class="dropdown">
					<summary class="btn btn-xs btn-ghost">Legend</summary>
					<div class="dropdown-content z-30 mt-2 max-h-72 w-64 overflow-auto border border-base-300 bg-base-100 p-3 shadow-lg">
						<For each={legend()}>
							{(label) => (
								<div class="flex items-center gap-2 py-1 text-xs">
									<span
										class="size-3 shrink-0 rounded-full"
										style={{
											"background-color":
												NODE_LABEL_COLORS[label] ?? DEFAULT_NODE_COLOR,
										}}
									/>
									<span class="truncate">{label}</span>
								</div>
							)}
						</For>
					</div>
				</details>
				<div class="flex-1" />
				<button
					class={iconButton}
					title="Zoom out"
					aria-label="Zoom out"
					onClick={() => surface?.zoomOut()}
				>
					<ZoomOut size={15} />
				</button>
				<button
					class={iconButton}
					title="Zoom in"
					aria-label="Zoom in"
					onClick={() => surface?.zoomIn()}
				>
					<ZoomIn size={15} />
				</button>
				<button
					class={iconButton}
					title="Zoom to fit"
					aria-label="Zoom to fit"
					onClick={() => surface?.fit(40)}
				>
					<Scan size={15} />
				</button>
				<button
					class={iconButton}
					title="Export PNG"
					aria-label="Export PNG"
					onClick={exportPNG}
				>
					<Download size={15} />
				</button>
				<Show when={props.onClear}>
					<button
						class={`${iconButton} text-error`}
						title="Clear graph"
						aria-label="Clear graph"
						onClick={props.onClear}
					>
						<Trash2 size={15} />
					</button>
				</Show>
				<button
					class={iconButton}
					title={fullscreen() ? "Exit full screen" : "Full screen"}
					aria-label={fullscreen() ? "Exit full screen" : "Full screen"}
					onClick={() => setFullscreen((value) => !value)}
				>
					<Show when={fullscreen()} fallback={<Maximize2 size={15} />}>
						<Minimize2 size={15} />
					</Show>
				</button>
			</div>
			<div class="relative flex-1 min-h-0">
				<div ref={containerRef!} class="cytoscape-canvas absolute inset-0" />
				<Show when={props.elements().length === 0}>
					<div class="pointer-events-none absolute inset-0 grid place-items-center text-sm text-base-content/45">
						{props.emptyMessage ?? "Run a graph query to begin."}
					</div>
				</Show>
			</div>
		</div>
	);
}
