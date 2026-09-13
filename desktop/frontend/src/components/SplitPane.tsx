import Resizable, { type Size } from "@corvu/resizable";
import { makePersisted } from "@solid-primitives/storage";
import { type JSX, createSignal } from "solid-js";

interface Props {
	direction: "horizontal" | "vertical";
	first: JSX.Element;
	second: JSX.Element;
	initialSize: number;
	minSize: Size;
	minSecond?: Size;
	storageKey?: string;
	sizedPane?: "first" | "second";
	class?: string;
}

export default function SplitPane(props: Props) {
	const sizeState = createSignal<number[]>([]);
	const [sizes, setSizes] = props.storageKey
		? makePersisted<number[], typeof sizeState>(sizeState, {
				name: `${props.storageKey}-ratios`,
			})
		: sizeState;
	const firstSize =
		props.sizedPane === "second" ? 1 - props.initialSize : props.initialSize;
	const secondSize = 1 - firstSize;
	const horizontal = props.direction === "horizontal";

	return (
		<Resizable
			orientation={props.direction}
			sizes={sizes()}
			onSizesChange={setSizes}
			initialSizes={[firstSize, secondSize]}
			class={`min-h-0 min-w-0 overflow-hidden ${props.class ?? ""}`}
		>
			<Resizable.Panel
				minSize={
					props.sizedPane === "second"
						? (props.minSecond ?? 220)
						: props.minSize
				}
				class="min-h-0 min-w-0 overflow-hidden"
			>
				{props.first}
			</Resizable.Panel>
			<Resizable.Handle
				type="button"
				aria-label="Resize panels"
				class={`group relative z-10 shrink-0 border-0 bg-transparent p-0 outline-none ${horizontal ? "w-1 cursor-col-resize" : "h-1 cursor-row-resize"}`}
			>
				<span
					class={`absolute bg-base-300 transition-colors group-hover:bg-primary/60 group-focus-visible:bg-primary ${horizontal ? "inset-y-0 left-1/2 w-px -translate-x-1/2" : "inset-x-0 top-1/2 h-px -translate-y-1/2"}`}
				/>
			</Resizable.Handle>
			<Resizable.Panel
				minSize={
					props.sizedPane === "second"
						? props.minSize
						: (props.minSecond ?? 220)
				}
				class="min-h-0 min-w-0 overflow-hidden"
			>
				{props.second}
			</Resizable.Panel>
		</Resizable>
	);
}
