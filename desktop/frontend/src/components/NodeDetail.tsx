import {
	type Accessor,
	For,
	Show,
	createEffect,
	createMemo,
	createResource,
	createSignal,
	on,
	onCleanup,
} from "solid-js";
import type {
	ExpandVertexCommand,
	ExpansionDirection,
	NeighborOption,
	QueryType,
} from "../application/graph-explorer";
import { graphExplorer } from "../bootstrap/graph";
import type { GraphSelection } from "../domain/graph";

interface Props {
	element: Accessor<GraphSelection>;
	onClose: () => void;
	onRemove: () => void;
	onExpand: (options: Omit<ExpandVertexCommand, "id" | "type">) => void;
	expanding: Accessor<boolean>;
	queryType: Accessor<QueryType>;
}

export default function NodeDetail(props: Props) {
	const [fullData, setFullData] = createSignal<Record<string, unknown> | null>(
		null,
	);
	const [loading, setLoading] = createSignal(false);
	const [error, setError] = createSignal("");
	const [direction, setDirection] = createSignal<ExpansionDirection>("both");
	const [relationship, setRelationship] = createSignal("");
	const [neighborLabel, setNeighborLabel] = createSignal("");
	const [limit, setLimit] = createSignal(10);
	const cache = new Map<string, Record<string, unknown>>();
	const [neighborSummary] = createResource(
		() => {
			const selected = props.element();
			if (!selected || selected.group !== "nodes") return null;
			return {
				id: selected.data.id,
				type: props.queryType(),
				direction: direction(),
			};
		},
		(command) => graphExplorer.neighborSummary(command),
	);

	createEffect(
		on(props.element, (selected) => {
			let active = true;
			onCleanup(() => {
				active = false;
			});
			setLoading(false);
			setError("");
			setRelationship("");
			setNeighborLabel("");
			setLimit(10);
			setFullData(selected?.data ?? null);
			if (!selected || selected.group !== "nodes") return;

			const cached = cache.get(selected.data.id);
			if (cached) {
				setFullData({ ...selected.data, ...cached });
				return;
			}

			setLoading(true);
			graphExplorer
				.vertexProperties(selected.data.id)
				.then((result) => {
					if (!active) return;
					setLoading(false);
					if (!result) return;
					if (JSON.stringify(result).length <= 32_000) {
						if (cache.size >= 20) cache.delete(cache.keys().next().value!);
						cache.set(selected.data.id, result);
					}
					setFullData({ ...selected.data, ...result });
				})
				.catch((cause) => {
					if (!active) return;
					setLoading(false);
					setError(cause instanceof Error ? cause.message : String(cause));
				});
		}),
	);

	const entries = createMemo(() => {
		const data = fullData();
		if (!data) return [];
		return Object.entries(data).filter(
			([key]) =>
				key !== "id" &&
				key !== "label" &&
				key !== "displayName" &&
				key !== "source" &&
				key !== "target",
		);
	});

	const title = () => String(props.element()?.data.label ?? "Graph element");
	const isNode = () => props.element()?.group === "nodes";
	const totalNeighbors = () =>
		neighborSummary()?.nodes.reduce(
			(total, option) => total + option.count,
			0,
		) ?? 0;
	const availableForSelection = createMemo(() => {
		const summary = neighborSummary();
		if (!summary) return 0;
		const counts: number[] = [];
		if (relationship()) {
			counts.push(
				summary.relationships.find((option) => option.label === relationship())
					?.count ?? 0,
			);
		}
		if (neighborLabel()) {
			counts.push(
				summary.nodes.find((option) => option.label === neighborLabel())
					?.count ?? 0,
			);
		}
		return counts.length > 0 ? Math.min(...counts) : totalNeighbors();
	});

	function selectRoute(
		value: string,
		options: NeighborOption[],
		setter: (value: string) => void,
	) {
		setter(value);
		const count = options.find((option) => option.label === value)?.count;
		if (count) setLimit(Math.min(100, count));
	}

	return (
		<div class="p-4 text-xs">
			<div class="flex items-start justify-between mb-4">
				<div class="min-w-0 flex-1">
					<div class="text-[10px] font-semibold uppercase text-base-content/40 mb-1">
						{isNode() ? "Node" : "Relationship"}
					</div>
					<div class="font-semibold text-sm">{title()}</div>
					<div class="text-base-content/50 font-mono text-[10px] mt-1 break-all leading-relaxed">
						{props.element()?.data.id}
					</div>
				</div>
				<button
					onClick={props.onClose}
					class="btn btn-ghost btn-xs btn-circle ml-2"
					title="Close details"
					aria-label="Close details"
				>
					<svg
						width="13"
						height="13"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
					>
						<path d="M18 6 6 18" />
						<path d="m6 6 12 12" />
					</svg>
				</button>
			</div>
			<Show when={!isNode()}>
				<dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-2 mb-4 rounded border border-base-300 p-3">
					<dt class="text-base-content/50">Source</dt>
					<dd class="font-mono text-[10px] break-all">
						{String(props.element()?.data.source ?? "")}
					</dd>
					<dt class="text-base-content/50">Target</dt>
					<dd class="font-mono text-[10px] break-all">
						{String(props.element()?.data.target ?? "")}
					</dd>
				</dl>
			</Show>

			<Show when={isNode()}>
				<section class="pb-4 mb-4 border-b border-base-300">
					<div class="mb-2 flex items-center justify-between gap-2">
						<div class="text-[10px] font-semibold uppercase text-base-content/40">
							Expand relationships
						</div>
						<Show when={neighborSummary.loading}>
							<span class="loading loading-spinner loading-xs" />
						</Show>
					</div>
					<Show when={neighborSummary.error}>
						<p class="mb-2 text-[10px] text-error">
							Could not inspect available relationships.
						</p>
					</Show>
					<Show when={!neighborSummary.loading && neighborSummary()}>
						<div class="mb-2 text-[10px] text-base-content/50">
							{totalNeighbors().toLocaleString()} connected nodes in this
							direction
						</div>
					</Show>
					<div class="grid grid-cols-2 gap-2">
						<label class="form-control">
							<span class="label-text text-[10px] mb-1">Direction</span>
							<select
								class="select select-bordered select-xs w-full"
								value={direction()}
								onChange={(event) =>
									setDirection(event.currentTarget.value as ExpansionDirection)
								}
							>
								<option value="both">Both</option>
								<option value="out">Outgoing</option>
								<option value="in">Incoming</option>
							</select>
						</label>
						<label class="form-control">
							<span class="label-text text-[10px] mb-1">Limit</span>
							<input
								class="input input-bordered input-xs w-full"
								type="number"
								min="1"
								max="100"
								value={limit()}
								onInput={(event) =>
									setLimit(Number.parseInt(event.currentTarget.value, 10) || 10)
								}
							/>
						</label>
					</div>
					<label class="form-control mt-2">
						<span class="label-text text-[10px] mb-1">Relationship type</span>
						<select
							class="select select-bordered select-xs w-full"
							value={relationship()}
							disabled={neighborSummary.loading}
							onChange={(event) =>
								selectRoute(
									event.currentTarget.value,
									neighborSummary()?.relationships ?? [],
									setRelationship,
								)
							}
						>
							<option value="">Any relationship</option>
							<For each={neighborSummary()?.relationships ?? []}>
								{(option) => (
									<option value={option.label}>
										{option.label} ({option.count.toLocaleString()})
									</option>
								)}
							</For>
						</select>
					</label>
					<label class="form-control mt-2">
						<span class="label-text text-[10px] mb-1">Connected node type</span>
						<select
							class="select select-bordered select-xs w-full"
							value={neighborLabel()}
							disabled={neighborSummary.loading}
							onChange={(event) =>
								selectRoute(
									event.currentTarget.value,
									neighborSummary()?.nodes ?? [],
									setNeighborLabel,
								)
							}
						>
							<option value="">Any node type</option>
							<For each={neighborSummary()?.nodes ?? []}>
								{(option) => (
									<option value={option.label}>
										{option.label} ({option.count.toLocaleString()})
									</option>
								)}
							</For>
						</select>
					</label>
					<Show when={availableForSelection() > 0}>
						<div class="mt-2 text-[10px] text-base-content/50">
							Up to {availableForSelection().toLocaleString()} connected nodes
							match
						</div>
					</Show>
					<button
						class="btn btn-primary btn-sm w-full mt-3"
						disabled={props.expanding()}
						onClick={() =>
							props.onExpand({
								direction: direction(),
								relationship: relationship().trim(),
								neighborLabel: neighborLabel().trim(),
								limit: Math.min(100, Math.max(1, limit())),
							})
						}
					>
						<Show when={props.expanding()} fallback="Expand connected nodes">
							<span class="loading loading-spinner loading-xs" /> Expanding
						</Show>
					</button>
				</section>
			</Show>

			<Show when={error()}>
				<p role="alert" class="text-error mb-3">
					{error()}
				</p>
			</Show>
			<Show when={loading()}>
				<div class="flex items-center gap-2 text-base-content/50 py-2">
					<span class="loading loading-spinner loading-xs" /> Loading
					properties...
				</div>
			</Show>

			<Show when={entries().length > 0}>
				<div class="text-[10px] font-semibold uppercase text-base-content/40 mb-2">
					Properties
				</div>
				<table class="w-full">
					<tbody>
						<For each={entries()}>
							{([key, value]) => (
								<tr class="border-t border-base-300/50">
									<td class="py-2 pr-3 font-medium text-base-content/50 whitespace-nowrap align-top">
										{key}
									</td>
									<td class="py-2 break-all font-mono text-[11px] text-base-content/80">
										{typeof value === "object"
											? JSON.stringify(value, null, 2)
											: String(value)}
									</td>
								</tr>
							)}
						</For>
					</tbody>
				</table>
			</Show>

			<button
				class="btn btn-error btn-outline btn-sm w-full mt-4"
				onClick={props.onRemove}
			>
				Remove from graph
			</button>
		</div>
	);
}
