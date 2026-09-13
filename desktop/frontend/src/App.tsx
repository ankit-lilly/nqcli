import {
	batch,
	lazy,
	Show,
	Suspense,
	createEffect,
	createSignal,
} from "solid-js";
import { DesktopService } from "../bindings/github.com/ankit-lilly/nqcli/internal/desktop";
import type { ExpandVertexCommand } from "./application/graph-explorer";
import { graphExplorer } from "./bootstrap/graph";
import {
	mergeGraphElements,
	removeGraphElement,
	visibleVertexIDs,
	type GraphElement,
	type GraphSelection,
} from "./domain/graph";
import NodeDetail from "./components/NodeDetail";
import ProfileSwitcher from "./components/ProfileSwitcher";
import QueryToolbar from "./components/QueryToolbar";
import ResultPanel from "./components/ResultPanel";
import SavedQueries from "./components/SavedQueries";
import SplitPane from "./components/SplitPane";
import StatusBar from "./components/StatusBar";
import ThemeSwitcher from "./components/ThemeSwitcher";
import { applyTheme, detectInitialTheme, type Theme } from "./lib/theme";

const QueryEditor = lazy(() => import("./components/QueryEditor"));
const SchemaView = lazy(() => import("./components/SchemaView"));

type Workspace = "graph" | "schema";

export default function App() {
	const [theme, setTheme] = createSignal<Theme>(detectInitialTheme());
	const [workspace, setWorkspace] = createSignal<Workspace>("graph");
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
	const [currentProfile, setCurrentProfile] = createSignal("");
	let generation = 0;
	let expansionGeneration = 0;

	createEffect(() => applyTheme(theme()));

	function clearResults() {
		batch(() => {
			setJsonResult("");
			setGraphElements([]);
			setSelectedElement(null);
			setError("");
			setWarning("");
			setQueryTime(0);
			setExpanding(false);
			++expansionGeneration;
		});
	}

	async function cancel() {
		++generation;
		try {
			await DesktopService.CancelQueries();
		} catch (cause) {
			setError(String(cause));
		} finally {
			setLoading(false);
		}
	}

	async function changeWorkspace(next: Workspace) {
		if (next === workspace() || switching()) return;
		if (loading() || expanding()) await cancel();
		setWorkspace(next);
	}

	async function execute() {
		if (loading() || switching() || !query().trim()) return;
		const token = ++generation;
		setLoading(true);
		clearResults();
		const start = performance.now();
		try {
			const response = await graphExplorer.run({
				query: query(),
				type: queryType(),
			});
			if (token !== generation) return;
			batch(() => {
				setGraphElements(response.elements);
				setJsonResult(response.json);
				setWarning(response.warning ?? "");
			});
		} catch (cause) {
			if (token === generation)
				setError(cause instanceof Error ? cause.message : String(cause));
		} finally {
			if (token === generation) {
				setQueryTime(performance.now() - start);
				setLoading(false);
			}
		}
	}

	async function expand(
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
		} catch (cause) {
			if (token === expansionGeneration)
				setError(cause instanceof Error ? cause.message : String(cause));
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

	const queryResults = () => (
		<div class="flex h-full min-h-0 flex-col">
			<Show when={warning() && viewMode() === "graph"}>
				<p
					role="status"
					class="border-b border-base-300 px-5 py-2 text-xs text-warning"
				>
					{warning()}
				</p>
			</Show>
			<div class="min-h-0 flex-1">
				<Show
					when={selectedElement() && viewMode() === "graph"}
					fallback={
						<ResultPanel
							viewMode={viewMode}
							jsonResult={jsonResult}
							graphElements={graphElements}
							error={error}
							selectedElement={selectedElement}
							setSelectedElement={setSelectedElement}
							loading={() => loading() || expanding()}
							onExpand={(id) => void expand(id)}
							onClear={() => {
								setGraphElements([]);
								setSelectedElement(null);
							}}
						/>
					}
				>
					<SplitPane
						class="h-full"
						direction="horizontal"
						initialSize={0.27}
						minSize={270}
						minSecond={420}
						sizedPane="second"
						storageKey="nq-graph-details-width"
						first={
							<ResultPanel
								viewMode={viewMode}
								jsonResult={jsonResult}
								graphElements={graphElements}
								error={error}
								selectedElement={selectedElement}
								setSelectedElement={setSelectedElement}
								loading={() => loading() || expanding()}
								onExpand={(id) => void expand(id)}
								onClear={() => {
									setGraphElements([]);
									setSelectedElement(null);
								}}
							/>
						}
						second={
							<NodeDetail
								element={selectedElement}
								onClose={() => setSelectedElement(null)}
								onRemove={removeSelectedElement}
								onExpand={(options) => {
									const selected = selectedElement();
									if (selected?.group === "nodes")
										void expand(selected.data.id, options);
								}}
								expanding={expanding}
								queryType={queryType}
							/>
						}
					/>
				</Show>
			</div>
			<StatusBar
				queryTime={queryTime}
				graphElements={graphElements}
				error={error}
				viewMode={viewMode}
				env={currentEnv}
			/>
		</div>
	);

	return (
		<div class="flex h-full flex-col bg-base-100 text-sm text-base-content">
			<header class="titlebar flex h-10 items-center gap-3 border-b border-base-300/30 bg-base-200/60 pl-20 pr-5 backdrop-blur-xl">
				<h1 class="select-none text-sm font-semibold">
					<span class="text-base-content/60">d</span>
					<span class="font-bold text-primary">Graph</span>
					<span class="text-base-content/60">er</span>
				</h1>
				<div class="h-4 w-px bg-base-content/10" />
				<ProfileSwitcher
					onSwitchStart={async () => {
						setSwitching(true);
						await cancel();
						clearResults();
					}}
					onSwitchEnd={() => setSwitching(false)}
					onProfileChange={(profile, env) => {
						setCurrentProfile(profile || "default");
						setCurrentEnv(env);
					}}
				/>
				<div class="h-4 w-px bg-base-content/10" />
				<nav class="no-drag flex gap-0.5" aria-label="Workspace">
					<button
						disabled={switching()}
						class={`btn btn-xs ${workspace() === "graph" ? "btn-primary" : "btn-ghost"}`}
						onClick={() => void changeWorkspace("graph")}
					>
						Graph
					</button>
					<button
						disabled={switching()}
						class={`btn btn-xs ${workspace() === "schema" ? "btn-primary" : "btn-ghost"}`}
						onClick={() => void changeWorkspace("schema")}
					>
						Schema
					</button>
				</nav>
				<div class="flex-1" />
				<ThemeSwitcher theme={theme} setTheme={setTheme} />
			</header>
			<Show
				when={!switching()}
				fallback={
					<div role="status" class="grid flex-1 place-items-center">
						<span class="loading loading-spinner loading-sm" />
					</div>
				}
			>
				<Show
					when={workspace() === "graph"}
					fallback={
						<Suspense fallback={<div class="p-4">Loading schema...</div>}>
							<SchemaView profile={currentProfile} />
						</Suspense>
					}
				>
					<SplitPane
						class="flex-1"
						direction="vertical"
						initialSize={0.26}
						minSize={128}
						minSecond={300}
						storageKey="nq-query-editor-height"
						first={
							<div class="flex h-full flex-col gap-2 px-5 py-3">
								<div class="flex flex-wrap items-center gap-2">
									<SavedQueries
										onSelect={(sample) => {
											setQuery(sample.query);
											setQueryType("gremlin");
											setViewMode(sample.view);
											clearResults();
										}}
										disabled={loading()}
									/>
									<div class="h-4 w-px bg-base-content/10" />
									<QueryToolbar
										queryType={queryType}
										setQueryType={setQueryType}
										viewMode={viewMode}
										onViewModeChange={(mode) => setViewMode(mode)}
										query={query}
										setQuery={setQuery}
										loading={loading}
										onExecute={execute}
										onCancel={() => void cancel()}
									/>
								</div>
								<div class="min-h-0 flex-1">
									<Suspense fallback={<p>Loading editor...</p>}>
										<QueryEditor
											query={query}
											setQuery={setQuery}
											queryType={queryType}
											onExecute={execute}
										/>
									</Suspense>
								</div>
							</div>
						}
						second={queryResults()}
					/>
				</Show>
			</Show>
		</div>
	);
}
