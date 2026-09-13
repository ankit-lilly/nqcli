import { type Accessor, type Setter, Show, lazy, Suspense } from "solid-js";
import type { GraphElement, GraphSelection } from "../domain/graph";
const GraphView = lazy(() => import("./GraphView"));
const JsonView = lazy(() => import("./JsonView"));

interface Props {
	viewMode: Accessor<"json" | "graph">;
	jsonResult: Accessor<string>;
	graphElements: Accessor<GraphElement[]>;
	error: Accessor<string>;
	selectedElement: Accessor<GraphSelection>;
	setSelectedElement: Setter<GraphSelection>;
	loading: Accessor<boolean>;
	onExpand: (id: string) => void;
	onClear: () => void;
}

export default function ResultPanel(props: Props) {
	return (
		<div class="h-full flex flex-col relative">
			<Show when={props.error()}>
				<div class="px-4 py-2 text-xs font-medium bg-error/10 text-error border-b border-error/20">
					{props.error()}
				</div>
			</Show>

			<div class="flex-1 min-h-0 overflow-hidden">
				<Show
					when={props.viewMode() === "graph"}
					fallback={
						<Suspense fallback={<div class="p-4">Loading JSON…</div>}>
							<JsonView result={props.jsonResult} />
						</Suspense>
					}
				>
					<Suspense fallback={<div class="p-4">Loading graph…</div>}>
						<GraphView
							elements={props.graphElements}
							selectedElement={props.selectedElement}
							setSelectedElement={props.setSelectedElement}
							onExpand={props.onExpand}
							onClear={props.onClear}
						/>
					</Suspense>
				</Show>
			</div>

			<Show when={props.loading()}>
				<div class="loading-overlay absolute inset-0 flex items-center justify-center bg-base-100/60">
					<div class="flex items-center gap-2.5 px-4 py-2.5 rounded-box border border-base-300 bg-base-100 shadow-lg">
						<span class="loading loading-spinner loading-sm" />
						<span class="text-xs text-base-content/60">
							Working with Neptune...
						</span>
					</div>
				</div>
			</Show>
		</div>
	);
}
