import { createSignal, For, Show } from "solid-js";
import {
	SAMPLE_QUERIES,
	buildSampleQuery,
	type SampleQuery,
} from "../lib/sampleQueries";

export default function SavedQueries(props: {
	onSelect: (sample: SampleQuery) => void;
	disabled: boolean;
}) {
	const [selected, setSelected] = createSignal("");
	let menu!: HTMLDetailsElement;

	function select(sample: SampleQuery) {
		if (props.disabled) return;
		props.onSelect({ ...sample, query: buildSampleQuery(sample) });
		setSelected(sample.name);
		menu.open = false;
	}

	return (
		<div class="flex items-center gap-3 flex-wrap text-xs">
			<details ref={menu!} class="dropdown">
				<summary
					class="btn btn-xs gap-1.5"
					aria-disabled={props.disabled}
					onClick={(event) => {
						if (props.disabled) event.preventDefault();
					}}
				>
					Sample queries{" "}
					<span class="badge badge-xs">{SAMPLE_QUERIES.length}</span>
				</summary>
				<div class="dropdown-content z-50 mt-2 w-96 max-w-[85vw] max-h-[65vh] overflow-y-auto rounded-box border border-base-300 bg-base-100 shadow-xl p-2">
					<ul class="menu w-full p-0">
						<For each={SAMPLE_QUERIES}>
							{(sample) => (
								<li>
									<button
										disabled={props.disabled}
										class="flex flex-col items-start gap-1"
										onClick={() => select(sample)}
									>
										<span class="flex w-full items-center justify-between gap-2">
											<span class="font-medium">{sample.name}</span>
											<span class="text-[10px] opacity-50 uppercase">
												{sample.view}
											</span>
										</span>
										<span class="text-[11px] opacity-60 text-left leading-relaxed">
											{sample.description}
										</span>
									</button>
								</li>
							)}
						</For>
					</ul>
				</div>
			</details>
			<Show when={selected()}>
				<span role="status" class="text-[11px] opacity-50">
					Loaded: {selected()}
				</span>
			</Show>
		</div>
	);
}
