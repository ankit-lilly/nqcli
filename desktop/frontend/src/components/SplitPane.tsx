import {
	type JSX,
	createEffect,
	createSignal,
	onCleanup,
	onMount,
} from "solid-js";

interface Props {
	direction: "horizontal" | "vertical";
	first: JSX.Element;
	second: JSX.Element;
	initialSize: number;
	minSize: number;
	minSecond?: number;
	storageKey?: string;
	sizedPane?: "first" | "second";
	class?: string;
}

export default function SplitPane(props: Props) {
	let root!: HTMLDivElement;
	let stopResize: (() => void) | undefined;
	const [size, setSize] = createSignal(props.initialSize);
	const minSecond = () => props.minSecond ?? 220;

	function clamp(value: number) {
		const available =
			props.direction === "horizontal"
				? root.getBoundingClientRect().width
				: root.getBoundingClientRect().height;
		return Math.min(
			Math.max(props.minSize, available - minSecond()),
			Math.max(props.minSize, value),
		);
	}

	onMount(() => {
		if (!props.storageKey) return;
		const saved = Number.parseFloat(
			localStorage.getItem(props.storageKey) ?? "",
		);
		if (Number.isFinite(saved)) setSize(clamp(saved));
	});

	createEffect(() => {
		const current = size();
		if (props.storageKey)
			localStorage.setItem(props.storageKey, String(current));
	});

	function startResize(event: PointerEvent) {
		event.preventDefault();
		stopResize?.();
		const origin =
			props.direction === "horizontal" ? event.clientX : event.clientY;
		const initial = size();
		const move = (next: PointerEvent) => {
			const point =
				props.direction === "horizontal" ? next.clientX : next.clientY;
			const delta = point - origin;
			setSize(clamp(initial + (props.sizedPane === "second" ? -delta : delta)));
		};
		stopResize = () => {
			window.removeEventListener("pointermove", move);
			window.removeEventListener("pointerup", stopResize!);
			document.body.classList.remove("resizing-panels");
			stopResize = undefined;
		};
		document.body.classList.add("resizing-panels");
		window.addEventListener("pointermove", move);
		window.addEventListener("pointerup", stopResize);
	}

	onCleanup(() => stopResize?.());

	function resizeWithKeyboard(event: KeyboardEvent) {
		const previous = props.direction === "horizontal" ? "ArrowLeft" : "ArrowUp";
		const next = props.direction === "horizontal" ? "ArrowRight" : "ArrowDown";
		if (event.key !== previous && event.key !== next) return;
		event.preventDefault();
		setSize((current) => clamp(current + (event.key === next ? 16 : -16)));
	}

	const sizedStyle = () =>
		props.direction === "horizontal"
			? { width: `${size()}px` }
			: { height: `${size()}px` };
	const sizedClass = "min-h-0 min-w-0 shrink-0 overflow-hidden";
	const flexibleClass = "min-h-0 min-w-0 flex-1 overflow-hidden";

	return (
		<div
			ref={root!}
			class={`flex min-h-0 min-w-0 overflow-hidden ${props.direction === "vertical" ? "flex-col" : "flex-row"} ${props.class ?? ""}`}
		>
			<div
				class={props.sizedPane === "second" ? flexibleClass : sizedClass}
				style={props.sizedPane === "second" ? undefined : sizedStyle()}
			>
				{props.first}
			</div>
			<div
				role="separator"
				tabIndex={0}
				aria-orientation={
					props.direction === "horizontal" ? "vertical" : "horizontal"
				}
				class={`group relative z-10 shrink-0 bg-base-300/70 outline-none hover:bg-primary/50 focus:bg-primary/60 ${props.direction === "horizontal" ? "w-px cursor-col-resize" : "h-px cursor-row-resize"}`}
				onPointerDown={startResize}
				onKeyDown={resizeWithKeyboard}
			>
				<span
					class={`absolute bg-base-content/20 group-hover:bg-primary ${props.direction === "horizontal" ? "left-[-2px] top-1/2 h-10 w-[5px] -translate-y-1/2" : "left-1/2 top-[-2px] h-[5px] w-10 -translate-x-1/2"}`}
				/>
			</div>
			<div
				class={props.sizedPane === "second" ? sizedClass : flexibleClass}
				style={props.sizedPane === "second" ? sizedStyle() : undefined}
			>
				{props.second}
			</div>
		</div>
	);
}
