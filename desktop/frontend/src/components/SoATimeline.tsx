import {
	createSignal,
	createMemo,
	onMount,
	onCleanup,
	For,
	Show,
} from "solid-js";
import type { TimelineResult } from "../lib/timeline";

const MARGIN = { left: 60, right: 20, top: 50, bottom: 30 };
const HEIGHT = 140;
const AXIS_Y = 90;
const MARKER_R = 5;
const EPOCH_HUES = [220, 150, 40, 330, 180];

export default function SoATimeline(props: { timeline: TimelineResult }) {
	let container!: HTMLDivElement;
	const [width, setWidth] = createSignal(800);

	onMount(() => {
		const ro = new ResizeObserver((entries) => {
			const w = entries[0]?.contentRect.width;
			if (w && w > 0) setWidth(Math.max(600, w));
		});
		ro.observe(container);
		onCleanup(() => ro.disconnect());
	});

	const plotWidth = () => width() - MARGIN.left - MARGIN.right;
	const minDay = createMemo(() => {
		const days = props.timeline.visits.map((v) => v.day + v.windowLowerDays);
		return days.length ? Math.min(...days) : 0;
	});
	const maxDay = createMemo(() => {
		const days = props.timeline.visits.map((v) => v.day + v.windowUpperDays);
		return days.length ? Math.max(...days) : 1;
	});
	const dayToX = (day: number) => {
		const range = maxDay() - minDay() || 1;
		return MARGIN.left + ((day - minDay()) / range) * plotWidth();
	};

	const tickInterval = createMemo(() => {
		const span = maxDay() - minDay();
		if (span <= 60) return 7;
		if (span <= 365) return 28;
		return 90;
	});
	const ticks = createMemo(() => {
		const result: number[] = [];
		const interval = tickInterval();
		const start = Math.ceil(minDay() / interval) * interval;
		for (let d = start; d <= maxDay(); d += interval) result.push(d);
		return result;
	});

	return (
		<div ref={container!} class="soa-timeline w-full">
			<svg
				width={width()}
				height={HEIGHT}
				viewBox={`0 0 ${width()} ${HEIGHT}`}
				class="block"
				role="img"
				aria-label="Study visit timeline"
			>
				{/* Epoch background bands */}
				<For each={props.timeline.epochBands}>
					{(band, i) => {
						const x1 = () => dayToX(band.startDay) - MARKER_R - 4;
						const x2 = () => dayToX(band.endDay) + MARKER_R + 4;
						const hue = EPOCH_HUES[i() % EPOCH_HUES.length];
						return (
							<g>
								<rect
									x={x1()}
									y={MARGIN.top - 10}
									width={Math.max(x2() - x1(), 8)}
									height={AXIS_Y - MARGIN.top + 20}
									rx="4"
									fill={`oklch(0.7 0.08 ${hue})`}
									opacity="0.12"
								/>
								<text
									x={(x1() + x2()) / 2}
									y={MARGIN.top - 14}
									text-anchor="middle"
									font-size="10"
									fill="currentColor"
									opacity="0.6"
								>
									{band.name}
								</text>
							</g>
						);
					}}
				</For>

				{/* Axis line */}
				<line
					x1={MARGIN.left}
					y1={AXIS_Y}
					x2={width() - MARGIN.right}
					y2={AXIS_Y}
					stroke="currentColor"
					opacity="0.2"
					stroke-width="1"
				/>

				{/* Tick marks and labels */}
				<Show when={props.timeline.valid}>
					<For each={ticks()}>
						{(day) => (
							<g>
								<line
									x1={dayToX(day)}
									y1={AXIS_Y - 3}
									x2={dayToX(day)}
									y2={AXIS_Y + 3}
									stroke="currentColor"
									opacity="0.3"
									stroke-width="1"
								/>
								<text
									x={dayToX(day)}
									y={AXIS_Y + 16}
									text-anchor="middle"
									font-size="9"
									fill="currentColor"
									opacity="0.4"
								>
									Day {day}
								</text>
							</g>
						)}
					</For>
				</Show>

				{/* Window range bars */}
				<For each={props.timeline.visits}>
					{(v) => {
						const hasWindow = () =>
							v.windowLowerDays !== 0 || v.windowUpperDays !== 0;
						return (
							<Show when={hasWindow()}>
								<rect
									x={dayToX(v.day + v.windowLowerDays)}
									y={AXIS_Y - 3}
									width={Math.max(
										dayToX(v.day + v.windowUpperDays) -
											dayToX(v.day + v.windowLowerDays),
										2,
									)}
									height="6"
									rx="3"
									fill="var(--color-info, oklch(0.7 0.15 230))"
									opacity="0.25"
								/>
							</Show>
						);
					}}
				</For>

				{/* Visit markers */}
				<For each={props.timeline.visits}>
					{(v) => (
						<circle
							cx={dayToX(v.day)}
							cy={AXIS_Y}
							r={MARKER_R}
							fill="var(--color-primary, oklch(0.6 0.2 260))"
						>
							<title>
								{v.name}
								{props.timeline.valid ? ` (Day ${v.day})` : ""}
							</title>
						</circle>
					)}
				</For>

				{/* Visit name labels */}
				<For each={props.timeline.visits}>
					{(v) => (
						<text
							x={dayToX(v.day)}
							y={AXIS_Y - 12}
							text-anchor="end"
							font-size="10"
							fill="currentColor"
							opacity="0.8"
							transform={`rotate(-45 ${dayToX(v.day)} ${AXIS_Y - 12})`}
						>
							{v.name.length > 20 ? v.name.slice(0, 18) + "…" : v.name}
						</text>
					)}
				</For>

				{/* Approximate notice */}
				<Show when={!props.timeline.valid}>
					<text
						x={width() - MARGIN.right}
						y={HEIGHT - 6}
						text-anchor="end"
						font-size="9"
						fill="currentColor"
						opacity="0.4"
						font-style="italic"
					>
						Positions are approximate — no parsed timing offsets in this
						schedule
					</text>
				</Show>
			</svg>
		</div>
	);
}
