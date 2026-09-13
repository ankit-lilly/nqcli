import { type Accessor, Show } from "solid-js";
import type { GraphElement } from "../domain/graph";

interface Props {
	queryTime: Accessor<number>;
	graphElements: Accessor<GraphElement[]>;
	error: Accessor<string>;
	viewMode: Accessor<"json" | "graph">;
	env: Accessor<string>;
}

const ENV_COLORS: Record<string, string> = {
	dev: "oklch(0.72 0.19 142)",
	qa: "oklch(0.75 0.18 55)",
	prod: "oklch(0.63 0.24 25)",
};

export default function StatusBar(props: Props) {
	const nodeCount = () =>
		props.graphElements().filter((e) => e.group === "nodes").length;
	const edgeCount = () =>
		props.graphElements().filter((e) => e.group === "edges").length;
	const envColor = () => ENV_COLORS[props.env()] || ENV_COLORS.dev;

	return (
		<div class="flex items-center gap-4 px-5 py-1 border-t border-base-300/30 bg-base-200/40 backdrop-blur-xl text-[11px] text-base-content/50">
			<div class="flex items-center gap-1.5">
				<span
					class="w-1.5 h-1.5 rounded-full"
					style={{ "background-color": envColor() }}
				/>
				<span class="uppercase font-semibold tracking-wider">
					{props.env()}
				</span>
			</div>
			<Show
				when={props.viewMode() === "graph" && props.graphElements().length > 0}
			>
				<span>
					<span class="font-medium text-base-content/70">{nodeCount()}</span>{" "}
					nodes
					<span class="opacity-30 mx-0.5">/</span>
					<span class="font-medium text-base-content/70">{edgeCount()}</span>{" "}
					edges
				</span>
			</Show>
			<Show when={props.queryTime() > 0}>
				<span>{Math.round(props.queryTime())}ms</span>
			</Show>
			<div class="ml-auto" />
			<Show when={props.error()}>
				<span class="text-error font-medium">Error</span>
			</Show>
			<span class="opacity-30">dGrapher</span>
		</div>
	);
}
