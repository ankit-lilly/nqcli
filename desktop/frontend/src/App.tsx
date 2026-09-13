import { createSignal, createEffect, Show, lazy, Suspense } from "solid-js";
import { DesktopService } from "../bindings/github.com/ankit-lilly/nqcli/internal/desktop";
import { detectInitialTheme, applyTheme, type Theme } from "./lib/theme";
const QueryEditor = lazy(() => import("./components/QueryEditor"));
import QueryToolbar from "./components/QueryToolbar";
import SavedQueries from "./components/SavedQueries";
import ResultPanel from "./components/ResultPanel";
import NodeDetail from "./components/NodeDetail";
import StatusBar from "./components/StatusBar";
import ProfileSwitcher from "./components/ProfileSwitcher";
import ThemeSwitcher from "./components/ThemeSwitcher";
import { graphExplorer } from "./bootstrap/graph";
import type { ExpandVertexCommand } from "./application/graph-explorer";
import {
	graphLabels,
	mergeGraphElements,
	removeGraphElement,
	visibleVertexIDs,
	type GraphElement,
	type GraphSelection,
} from "./domain/graph";
const SoAMatrix = lazy(() => import("./components/SoAMatrix"));

export default function App() {
	const [theme, setTheme] = createSignal<Theme>(detectInitialTheme());
	const [workspace, setWorkspace] = createSignal<"soa" | "query">("soa");
	const [viewMode, setViewMode] = createSignal<"json" | "graph">("graph");
	const [query, setQuery] = createSignal("g.V().hasLabel('Study').limit(20)");
	const [queryType, setQueryType] = createSignal<"gremlin" | "cypher">(
		"gremlin",
	);
	const [loading, setLoading] = createSignal(false);
	const [expanding, setExpanding] = createSignal(false);
	const [switching, setSwitching] = createSignal(false);
	const [jsonResult, setJsonResult] = createSignal("");
	const [graphElements, setGraphElements] = createSignal<GraphElement[]>([]);
	const [error, setError] = createSignal("");
	const [warning, setWarning] = createSignal("");
	const [selectedElement, setSelectedElement] =
		createSignal<GraphSelection>(null);
	const [queryTime, setQueryTime] = createSignal(0);
	const [currentEnv, setCurrentEnv] = createSignal("dev");
	let generation = 0;
	let expansionGeneration = 0;
	createEffect(() => applyTheme(theme()));
	function clearResults() {
		setJsonResult("");
		setGraphElements([]);
		setSelectedElement(null);
		setError("");
		setWarning("");
		setQueryTime(0);
		setExpanding(false);
		++expansionGeneration;
	}
	async function cancel() {
		++generation;
		// Keep Run disabled until cancellation has reached Go, so a new query cannot
		// accidentally inherit the context that is about to be cancelled.
		try {
			await DesktopService.CancelQueries();
		} catch (e) {
			setError(String(e));
		} finally {
			setLoading(false);
		}
	}
	async function changeWorkspace(next: "soa" | "query") {
		if (next === workspace() || switching()) return;
		setSwitching(true);
		await cancel();
		clearResults();
		setWorkspace(next);
		setSwitching(false);
	}
	async function handleExecute() {
		if (loading() || switching() || !query().trim()) return;
		const token = ++generation,
			mode = viewMode();
		setLoading(true);
		clearResults();
		const start = performance.now();
		try {
			if (mode === "graph") {
				const response = await graphExplorer.run({
					query: query(),
					type: queryType(),
				});
				if (token !== generation) return;
				setGraphElements(response.elements);
				setWarning(response.warning ?? "");
			} else {
				const response = await DesktopService.ExecuteQuery({
					query: query(),
					type: queryType(),
					serializer: "",
				});
				if (token !== generation) return;
				if (response.error) throw new Error(response.error);
				setJsonResult(response.processed || "");
			}
		} catch (e) {
			if (token === generation)
				setError(e instanceof Error ? e.message : String(e));
		} finally {
			if (token === generation) {
				setQueryTime(performance.now() - start);
				setLoading(false);
			}
		}
	}

	async function handleExpand(
		id: string,
		options: Omit<ExpandVertexCommand, "id" | "type"> = {},
	) {
		if (expanding() || switching()) return;
		const token = ++expansionGeneration;
		setExpanding(true);
		setError("");
		try {
			const response = await graphExplorer.expand({
				id,
				type: queryType(),
				excludedVertexIds: visibleVertexIDs(graphElements()),
				...options,
			});
			if (token !== expansionGeneration) return;
			if (response.elements.length === 0) {
				setWarning("No more connected nodes matched this expansion.");
				return;
			}
			setGraphElements((current) =>
				mergeGraphElements(current, response.elements),
			);
			setWarning(response.warning ?? "");
		} catch (e) {
			if (token === expansionGeneration) {
				setError(e instanceof Error ? e.message : String(e));
			}
		} finally {
			if (token === expansionGeneration) setExpanding(false);
		}
	}

	function removeSelectedElement() {
		const selected = selectedElement();
		if (!selected) return;
		setGraphElements((elements) => removeGraphElement(elements, selected));
		setSelectedElement(null);
	}

	const labels = () => graphLabels(graphElements());
	return (
		<div class="flex flex-col h-full bg-base-100 text-base-content text-sm">
			<div class="titlebar flex items-center gap-3 pl-20 pr-5 h-10 bg-base-200/60 backdrop-blur-xl border-b border-base-300/30">
				<h1 class="text-sm font-semibold tracking-tight select-none">
					<span class="text-base-content/60">d</span>
					<span class="text-primary font-bold">Graph</span>
					<span class="text-base-content/60">er</span>
				</h1>
				<div class="w-px h-4 bg-base-content/10" />
				<ProfileSwitcher
					onSwitchStart={async () => {
						setSwitching(true);
						await cancel();
						clearResults();
					}}
					onSwitchEnd={() => setSwitching(false)}
					onProfileChange={(_profile, env) => setCurrentEnv(env)}
				/>
				<div class="w-px h-4 bg-base-content/10" />
				<nav class="flex gap-0.5 no-drag" aria-label="Workspace">
					<button
						disabled={switching()}
						class={`btn btn-xs ${workspace() === "soa" ? "btn-primary" : "btn-ghost"}`}
						onClick={() => void changeWorkspace("soa")}
					>
						Schedule of Activities
					</button>
					<button
						disabled={switching()}
						class={`btn btn-xs ${workspace() === "query" ? "btn-primary" : "btn-ghost"}`}
						onClick={() => void changeWorkspace("query")}
					>
						Query workspace
					</button>
				</nav>
				<div class="flex-1" />
				<ThemeSwitcher theme={theme} setTheme={setTheme} />
			</div>
			<Show
				when={!switching()}
				fallback={
					<div role="status" class="p-8">
						Switching…
					</div>
				}
			>
				<Show
					when={workspace() === "soa"}
					fallback={
						<>
							<div class="flex flex-col gap-2 px-5 py-3 border-b border-base-300">
								<div class="flex items-center gap-2 flex-wrap">
									<SavedQueries
										onSelect={(sample) => {
											setQuery(sample.query);
											setQueryType("gremlin");
											setViewMode(sample.view);
											clearResults();
										}}
										disabled={loading()}
									/>
									<div class="w-px h-4 bg-base-content/10" />
									<QueryToolbar
										queryType={queryType}
										setQueryType={setQueryType}
										viewMode={viewMode}
										setViewMode={setViewMode}
										query={query}
										setQuery={setQuery}
										loading={loading}
										onExecute={handleExecute}
										onCancel={() => void cancel()}
									/>
								</div>
								<Suspense fallback={<p>Loading editor…</p>}>
									<QueryEditor
										query={query}
										setQuery={setQuery}
										queryType={queryType}
										onExecute={handleExecute}
									/>
								</Suspense>
							</div>
							<Show when={warning()}>
								<p role="status" class="px-5 py-2 text-xs text-warning">
									{warning()}
								</p>
							</Show>
							<div class="flex flex-1 min-h-0 overflow-hidden">
								<div class="flex-1 min-w-0 overflow-hidden">
									<ResultPanel
										viewMode={viewMode}
										jsonResult={jsonResult}
										graphElements={graphElements}
										error={error}
										selectedElement={selectedElement}
										setSelectedElement={setSelectedElement}
										loading={() => loading() || expanding()}
										onExpand={(id) => void handleExpand(id)}
										onClear={() => {
											setGraphElements([]);
											setSelectedElement(null);
										}}
									/>
								</div>
								<Show when={selectedElement()}>
									<div class="w-80 shrink-0 border-l border-base-300 overflow-auto">
										<NodeDetail
											element={selectedElement}
											onClose={() => setSelectedElement(null)}
											onRemove={removeSelectedElement}
											onExpand={(options) => {
												const selected = selectedElement();
												if (selected?.group === "nodes") {
													void handleExpand(selected.data.id, options);
												}
											}}
											expanding={expanding}
											nodeLabels={() => labels().nodeLabels}
											relationshipLabels={() => labels().relationshipLabels}
										/>
									</div>
								</Show>
							</div>
							<StatusBar
								queryTime={queryTime}
								graphElements={graphElements}
								error={error}
								viewMode={viewMode}
								env={currentEnv}
							/>
						</>
					}
				>
					<div class="flex-1 min-h-0">
						<Suspense fallback={<p class="p-4">Loading schedule explorer…</p>}>
							<SoAMatrix />
						</Suspense>
					</div>
				</Show>
			</Show>
		</div>
	);
}
