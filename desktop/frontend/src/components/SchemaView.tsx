import { RefreshCw } from "lucide-solid";
import {
	For,
	Show,
	type Accessor,
	createEffect,
	createMemo,
	createResource,
	createSelector,
	createSignal,
	onCleanup,
} from "solid-js";
import { schemaExplorer } from "../bootstrap/graph";
import type { GraphSelection } from "../domain/graph";
import { DEFAULT_NODE_COLOR, NODE_LABEL_COLORS } from "../domain/graph-style";
import {
	formatCount,
	schemaConnectionID,
	schemaGraphElements,
	schemaSelection,
	type SchemaSelectionDetails,
	type SchemaSnapshot,
} from "../domain/schema";
import GraphView from "./GraphView";
import SplitPane from "./SplitPane";

const emptySchema: SchemaSnapshot = {
	status: "empty",
	vertices: [],
	edges: [],
	edgeConnections: [],
};

export default function SchemaView(props: { profile: Accessor<string> }) {
	const [selection, setSelection] = createSignal<GraphSelection>(null);
	const [refreshing, setRefreshing] = createSignal(false);
	const [refreshError, setRefreshError] = createSignal("");
	const [snapshot, { mutate, refetch }] = createResource(
		() => props.profile(),
		() => schemaExplorer.load(),
		{ initialValue: emptySchema },
	);
	// Keep the previous snapshot visible while the completed schema is refetched.
	const currentSnapshot = () => snapshot.latest ?? emptySchema;
	const elements = createMemo(() => schemaGraphElements(currentSnapshot()));
	const error = () => {
		if (refreshError()) return refreshError();
		if (snapshot.error)
			return snapshot.error instanceof Error
				? snapshot.error.message
				: String(snapshot.error);
		return currentSnapshot().error ?? "";
	};

	createEffect(() => {
		props.profile();
		setSelection(null);
	});

	const unsubscribe = schemaExplorer.subscribe((event) => {
		const profile = props.profile() || "default";
		if (event.key !== profile) return;
		mutate((current) => ({
			...(current ?? emptySchema),
			status: event.status,
			phase: event.phase,
			completed: event.completed,
			total: event.total,
			error: event.error,
		}));
		if (event.status === "ready") void refetch();
	});
	onCleanup(unsubscribe);

	async function refresh() {
		if (refreshing()) return;
		setRefreshing(true);
		setRefreshError("");
		try {
			mutate(await schemaExplorer.refresh());
		} catch (cause) {
			setRefreshError(cause instanceof Error ? cause.message : String(cause));
		} finally {
			setRefreshing(false);
		}
	}

	return (
		<div class="flex h-full min-h-0 flex-col">
			<div class="flex min-h-11 items-center gap-3 border-b border-base-300 px-4">
				<div>
					<div class="text-xs font-semibold">Schema discovery</div>
					<div class="text-[10px] text-base-content/45">
						{formatCount(currentSnapshot().totalVertices)} vertices ·{" "}
						{formatCount(currentSnapshot().totalEdges)} edges
					</div>
				</div>
				<Show when={currentSnapshot().status === "running"}>
					<div class="flex items-center gap-2 text-xs text-base-content/55">
						<span class="loading loading-spinner loading-xs" />
						<span>
							{currentSnapshot().phase || "Discovering"}
							{currentSnapshot().total
								? ` ${currentSnapshot().completed ?? 0}/${currentSnapshot().total}`
								: ""}
						</span>
					</div>
				</Show>
				<div class="flex-1" />
				<Show when={currentSnapshot().lastUpdate}>
					<span class="text-[10px] text-base-content/40">
						Updated{" "}
						{new Date(currentSnapshot().lastUpdate!).toLocaleTimeString()}
					</span>
				</Show>
				<button
					class="btn btn-xs btn-outline"
					disabled={refreshing() || currentSnapshot().status === "running"}
					onClick={() => void refresh()}
				>
					<RefreshCw size={13} /> Refresh
				</button>
			</div>
			<Show when={error()}>
				<p
					role="alert"
					class="border-b border-error/20 bg-error/10 px-4 py-2 text-xs text-error"
				>
					{error()}
				</p>
			</Show>
			<SplitPane
				class="flex-1"
				direction="horizontal"
				initialSize={0.24}
				minSize={240}
				minSecond={420}
				storageKey="nq-schema-sidebar-width"
				first={
					<SchemaSidebar
						snapshot={currentSnapshot}
						selection={selection}
						setSelection={setSelection}
					/>
				}
				second={
					<GraphView
						elements={elements}
						selectedElement={selection}
						setSelectedElement={setSelection}
						focusSelection
						emptyMessage={
							currentSnapshot().status === "running"
								? "Discovering graph schema..."
								: "No schema data is available."
						}
					/>
				}
			/>
		</div>
	);
}

function SchemaSidebar(props: {
	snapshot: Accessor<SchemaSnapshot>;
	selection: Accessor<GraphSelection>;
	setSelection: (selection: GraphSelection) => void;
}) {
	const [filter, setFilter] = createSignal("");
	const selected = createMemo(() =>
		schemaSelection(props.snapshot(), props.selection()),
	);
	const selectedKey = () => {
		const selection = props.selection();
		return selection ? `${selection.group}:${selection.data.id}` : "";
	};
	const isSelected = createSelector(selectedKey);
	const vertices = createMemo(() => {
		const query = filter().trim().toLowerCase();
		return props
			.snapshot()
			.vertices.filter((vertex) => vertex.type.toLowerCase().includes(query));
	});
	const connections = createMemo(() => {
		const query = filter().trim().toLowerCase();
		return props
			.snapshot()
			.edgeConnections.filter((connection) =>
				`${connection.sourceVertexType} ${connection.edgeType} ${connection.targetVertexType}`
					.toLowerCase()
					.includes(query),
			);
	});

	return (
		<aside class="flex h-full flex-col bg-base-200/30">
			<div class="border-b border-base-300 p-3">
				<input
					class="input input-bordered input-sm w-full"
					type="search"
					placeholder="Filter node or relationship types"
					value={filter()}
					onInput={(event) => setFilter(event.currentTarget.value)}
				/>
			</div>
			<div class="flex-1 overflow-auto">
				<Show when={selected()}>
					{(item) => (
						<SchemaDetails
							item={item()}
							onClose={() => props.setSelection(null)}
						/>
					)}
				</Show>
				<section class="border-b border-base-300 py-2">
					<div class="flex items-center justify-between px-3 py-1 text-[10px] font-semibold uppercase text-base-content/45">
						<span>Node types</span>
						<span>{vertices().length}</span>
					</div>
					<For each={vertices()}>
						{(vertex) => {
							const key = `nodes:${vertex.type}`;
							return (
								<button
									class={`flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-base-200 ${isSelected(key) ? "bg-primary/10 text-primary" : ""}`}
									onClick={() =>
										props.setSelection({
											group: "nodes",
											data: { id: vertex.type, label: vertex.type },
										})
									}
								>
									<span
										class="size-2.5 shrink-0 rounded-full"
										style={{
											"background-color":
												NODE_LABEL_COLORS[vertex.type] ?? DEFAULT_NODE_COLOR,
										}}
									/>
									<span class="min-w-0 flex-1 truncate">{vertex.type}</span>
									<span class="text-[10px] text-base-content/40">
										{formatCount(vertex.total)}
									</span>
								</button>
							);
						}}
					</For>
				</section>
				<section class="py-2">
					<div class="flex items-center justify-between px-3 py-1 text-[10px] font-semibold uppercase text-base-content/45">
						<span>Relationships</span>
						<span>{connections().length}</span>
					</div>
					<For each={connections()}>
						{(connection) => {
							const id = schemaConnectionID(connection);
							const key = `edges:${id}`;
							return (
								<button
									class={`w-full px-3 py-1.5 text-left hover:bg-base-200 ${isSelected(key) ? "bg-primary/10 text-primary" : ""}`}
									onClick={() =>
										props.setSelection({
											group: "edges",
											data: {
												id,
												label: connection.edgeType,
												source: connection.sourceVertexType,
												target: connection.targetVertexType,
											},
										})
									}
								>
									<div class="truncate text-xs font-medium">
										{connection.edgeType}
									</div>
									<div class="truncate text-[10px] text-base-content/45">
										{connection.sourceVertexType} →{" "}
										{connection.targetVertexType}
									</div>
								</button>
							);
						}}
					</For>
				</section>
			</div>
		</aside>
	);
}

function SchemaDetails(props: {
	item: SchemaSelectionDetails;
	onClose: () => void;
}) {
	const attributes = () =>
		props.item.kind === "vertex"
			? props.item.vertex.attributes
			: (props.item.edge?.attributes ?? []);
	return (
		<section class="border-b border-base-300 p-3">
			<div class="flex items-start gap-2">
				<div class="min-w-0 flex-1">
					<div class="text-[10px] font-semibold uppercase text-base-content/45">
						{props.item.kind === "vertex" ? "Node type" : "Relationship"}
					</div>
					<div class="truncate text-sm font-semibold">
						{props.item.kind === "vertex"
							? props.item.vertex.type
							: props.item.connection.edgeType}
					</div>
				</div>
				<button
					class="btn btn-ghost btn-xs btn-square"
					aria-label="Close details"
					onClick={props.onClose}
				>
					×
				</button>
			</div>
			<Show
				when={props.item.kind === "vertex"}
				fallback={
					<div class="mt-2 text-xs text-base-content/60">
						{props.item.kind === "connection"
							? props.item.connection.sourceVertexType
							: ""}{" "}
						→{" "}
						{props.item.kind === "connection"
							? props.item.connection.targetVertexType
							: ""}
					</div>
				}
			>
				<div class="mt-2 text-xs text-base-content/55">
					{formatCount(
						props.item.kind === "vertex" ? props.item.vertex.total : undefined,
					)}{" "}
					total
				</div>
			</Show>
			<div class="mt-3 text-[10px] font-semibold uppercase text-base-content/45">
				Properties
			</div>
			<Show
				when={attributes().length > 0}
				fallback={
					<p class="mt-1 text-xs text-base-content/45">
						No properties discovered
					</p>
				}
			>
				<For each={attributes()}>
					{(attribute) => (
						<div class="flex justify-between gap-2 border-t border-base-300/60 py-1.5 text-xs">
							<span class="truncate">{attribute.name}</span>
							<span class="text-[10px] text-base-content/40">
								{attribute.dataType}
							</span>
						</div>
					)}
				</For>
			</Show>
		</section>
	);
}
