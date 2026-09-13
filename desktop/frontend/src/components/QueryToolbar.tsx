import { type Accessor, type Setter, Show } from "solid-js";
import { formatQuery } from "../lib/formatQuery";

interface Props {
	queryType: Accessor<"gremlin" | "cypher">;
	setQueryType: Setter<"gremlin" | "cypher">;
	viewMode: Accessor<"json" | "graph">;
	setViewMode: Setter<"json" | "graph">;
	query: Accessor<string>;
	setQuery: Setter<string>;
	loading: Accessor<boolean>;
	onExecute: () => void;
	onCancel: () => void;
}

export default function QueryToolbar(props: Props) {
	return (
		<>
			<select
				disabled={props.loading()}
				value={props.queryType()}
				onChange={(e) => {
					const type = e.currentTarget.value as "gremlin" | "cypher";
					props.setQueryType(type);
				}}
				class="select select-bordered select-xs"
			>
				<option value="gremlin">Gremlin</option>
				<option value="cypher">Cypher</option>
			</select>

			<div class="join">
				<button
					disabled={props.loading()}
					onClick={() => props.setViewMode("graph")}
					class={`join-item btn btn-xs ${props.viewMode() === "graph" ? "btn-primary" : "btn-ghost"}`}
				>
					Graph
				</button>
				<button
					disabled={props.loading()}
					onClick={() => props.setViewMode("json")}
					class={`join-item btn btn-xs ${props.viewMode() === "json" ? "btn-primary" : "btn-ghost"}`}
				>
					JSON
				</button>
			</div>

			<button
				disabled={props.loading()}
				onClick={() =>
					props.setQuery(formatQuery(props.query(), props.queryType()))
				}
				class="btn btn-ghost btn-xs gap-1"
				title="Format query (Shift+Alt+F)"
			>
				<svg
					width="12"
					height="12"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<line x1="21" y1="10" x2="7" y2="10" />
					<line x1="21" y1="6" x2="3" y2="6" />
					<line x1="21" y1="14" x2="3" y2="14" />
					<line x1="21" y1="18" x2="7" y2="18" />
				</svg>
				Format
			</button>

			<div class="flex-1" />

			<Show when={props.loading()}>
				<button class="btn btn-xs" onClick={props.onCancel}>
					Cancel
				</button>
			</Show>
			<button
				onClick={props.onExecute}
				disabled={props.loading()}
				class="btn btn-primary btn-xs gap-1"
			>
				<Show
					when={props.loading()}
					fallback={
						<>
							<svg
								width="10"
								height="12"
								viewBox="0 0 10 12"
								fill="currentColor"
							>
								<path d="M0 0L10 6L0 12z" />
							</svg>
							Run
						</>
					}
				>
					<span class="loading loading-spinner loading-xs" />
					Running
				</Show>
			</button>
		</>
	);
}
