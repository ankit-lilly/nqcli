import {
	createSignal,
	createMemo,
	onMount,
	onCleanup,
	For,
	Show,
} from "solid-js";
import { DesktopService } from "../../bindings/github.com/ankit-lilly/nqcli/internal/desktop";
import {
	prepareSoA,
	cellDetails,
	type Named,
	type Versions,
	type Version,
	type SoAData,
	type Row,
	type Visit,
} from "../lib/soa";
import { buildTimeline } from "../lib/timeline";
import SoATimeline from "./SoATimeline";

const ROW_PAGE = 30,
	VISIT_PAGE = 10;

export default function SoAMatrix() {
	const [search, setSearch] = createSignal("");
	const [studies, setStudies] = createSignal<Named[]>([]);
	const [study, setStudy] = createSignal("");
	const [versions, setVersions] = createSignal<Version[]>([]);
	const [version, setVersion] = createSignal("");
	const [design, setDesign] = createSignal("");
	const [matrix, setMatrix] = createSignal<ReturnType<
		typeof prepareSoA
	> | null>(null);
	const [loading, setLoading] = createSignal(false);
	const [error, setError] = createSignal("");
	const [notice, setNotice] = createSignal("");
	const [versionNotice, setVersionNotice] = createSignal("");
	const [catalogNotice, setCatalogNotice] = createSignal("");
	const [filter, setFilter] = createSignal("");
	const [labsOnly, setLabsOnly] = createSignal(false);
	const [rowPage, setRowPage] = createSignal(0);
	const [visitPage, setVisitPage] = createSignal(0);
	const [selected, setSelected] = createSignal<{
		row: Row;
		visit: Visit;
	} | null>(null);
	const [timelineOpen, setTimelineOpen] = createSignal(true);
	let generation = 0;
	const currentVersion = createMemo(() =>
		versions().find((v) => v.id === version()),
	);
	const rows = createMemo(() =>
		(matrix()?.rows || []).filter(
			(r) =>
				(!labsOnly() || r.kind === "Lab concept") &&
				r.name.toLowerCase().includes(filter().toLowerCase()),
		),
	);
	const shownRows = createMemo(() =>
		rows().slice(rowPage() * ROW_PAGE, (rowPage() + 1) * ROW_PAGE),
	);
	const visits = createMemo(() => matrix()?.data.visits || []);
	const shownVisits = createMemo(() =>
		visits().slice(visitPage() * VISIT_PAGE, (visitPage() + 1) * VISIT_PAGE),
	);
	const selectedInstances = createMemo(() => {
		const s = selected();
		return s ? matrix()?.cells.get(s.visit.id)?.get(s.row.id) || [] : [];
	});
	const details = createMemo(() => {
		const s = selected(),
			m = matrix();
		return s && m
			? cellDetails(m.data, selectedInstances(), s.row.id)
			: { timings: [], conditions: [] };
	});
	const timeline = createMemo(() => {
		const m = matrix();
		return m ? buildTimeline(m.data, m.data.visits) : null;
	});
	const timingVisitNames = (ids: string[]) => {
		const m = matrix();
		if (!m) return "Unknown";
		return (
			ids
				.map((id) => {
					const instance = m.data.instances.find((i) => i.id === id);
					return (
						instance?.visits
							.map(
								(v) =>
									m.data.visits.find((item) => item.id === v)?.name ||
									"Unloaded visit",
							)
							.join(", ") || "Unloaded instance"
					);
				})
				.join(", ") || "Not specified"
		);
	};
	function clearMatrix() {
		setMatrix(null);
		setSelected(null);
		setRowPage(0);
		setVisitPage(0);
		setNotice("");
	}
	async function request(action: (token: number) => Promise<void>) {
		const token = ++generation;
		setLoading(true);
		setError("");
		try {
			await action(token);
		} catch (e) {
			if (token === generation)
				setError(e instanceof Error ? e.message : String(e));
		} finally {
			if (token === generation) setLoading(false);
		}
	}
	const ensure = <T extends { error?: string }>(response: T): T => {
		if (response.error) throw new Error(response.error);
		return response;
	};
	async function findStudies() {
		if (loading()) return;
		setStudy("");
		setVersions([]);
		setVersion("");
		setDesign("");
		setVersionNotice("");
		clearMatrix();
		await request(async (token) => {
			const r = ensure(await DesktopService.ListStudies({ search: search() }));
			if (token !== generation) return;
			setStudies((r.data || []) as Named[]);
			setCatalogNotice(r.warning || "");
		});
	}
	async function chooseStudy(id: string) {
		setVersionNotice("");
		setStudy(id);
		setVersions([]);
		setVersion("");
		setDesign("");
		clearMatrix();
		if (!id) return;
		await request(async (token) => {
			const r = ensure(await DesktopService.ListVersions({ studyId: id }));
			if (token !== generation) return;
			const result = (r.data as Versions[])?.[0];
			const list = result?.versions || [];
			if (list.length > 200 || list.some((v) => v.designs.length > 20))
				setVersionNotice(
					"Version/design list is limited. Use the query workspace to inspect older versions or additional designs.",
				);
			setVersions(
				list
					.slice(0, 200)
					.map((v) => ({ ...v, designs: v.designs.slice(0, 20) })),
			);
			const chosen = list.find((v) => result.latest.includes(v.id)) || list[0];
			if (chosen) {
				setVersion(chosen.id);
				setDesign(chosen.designs[0]?.id || "");
			}
		});
	}
	async function loadMatrix() {
		clearMatrix();
		await request(async (token) => {
			const r = ensure(
				await DesktopService.GetSoA({
					versionId: version(),
					designId: design(),
				}),
			);
			if (token !== generation) return;
			const raw = (r.data as SoAData[])?.[0];
			if (!raw)
				throw new Error(
					"This design is no longer available for the selected version. Reload the study.",
				);
			setMatrix(prepareSoA(raw));
		});
	}
	async function cancel() {
		++generation;
		await DesktopService.CancelQueries().catch(() => {});
		setLoading(false);
	}
	onMount(() => {
		void findStudies();
	});
	onCleanup(() => {
		++generation;
	});

	return (
		<section
			class="flex flex-col h-full min-h-0"
			aria-label="Schedule of Activities"
		>
			<div class="p-4 border-b border-base-300 space-y-3">
				<div>
					<h2 class="font-semibold">Schedule of Activities</h2>
					<p class="text-xs opacity-60">
						Explore planned assessments by visit, then inspect their timing and
						conditions.
					</p>
				</div>
				<form
					class="flex gap-2 flex-wrap"
					onSubmit={(e) => {
						e.preventDefault();
						void findStudies();
					}}
				>
					<input
						class="input input-sm w-64"
						aria-label="Search studies by trial alias"
						placeholder="Search trial alias (case-sensitive)"
						value={search()}
						onInput={(e) => setSearch(e.currentTarget.value)}
					/>
					<button class="btn btn-sm" disabled={loading()}>
						Search studies
					</button>
					<Show when={loading()}>
						<span role="status" class="text-sm self-center">
							Loading…
						</span>
						<button
							type="button"
							class="btn btn-sm btn-ghost"
							onClick={() => void cancel()}
						>
							Cancel
						</button>
					</Show>
				</form>
				<Show when={catalogNotice()}>
					<p class="text-xs text-warning">{catalogNotice()}</p>
				</Show>
				<div class="flex gap-3 flex-wrap items-end">
					<label class="text-xs flex flex-col gap-1">
						Study
						<select
							class="select select-sm w-60"
							value={study()}
							disabled={loading()}
							onChange={(e) => void chooseStudy(e.currentTarget.value)}
						>
							<option value="">Select a study</option>
							<For each={studies()}>
								{(s) => <option value={s.id}>{s.name}</option>}
							</For>
						</select>
					</label>
					<label class="text-xs flex flex-col gap-1">
						Version
						<select
							class="select select-sm w-64"
							value={version()}
							disabled={loading()}
							onChange={(e) => {
								setVersion(e.currentTarget.value);
								setDesign(currentVersion()?.designs[0]?.id || "");
								clearMatrix();
							}}
						>
							<option value="">Select a version</option>
							<For each={versions()}>
								{(v) => (
									<option value={v.id}>
										{v.name} · {v.status} · {v.amendment.join(", ")}
									</option>
								)}
							</For>
						</select>
					</label>
					<label class="text-xs flex flex-col gap-1">
						Design
						<select
							class="select select-sm w-64"
							value={design()}
							disabled={loading()}
							onChange={(e) => {
								setDesign(e.currentTarget.value);
								clearMatrix();
							}}
						>
							<option value="">Select a design</option>
							<For each={currentVersion()?.designs || []}>
								{(d) => <option value={d.id}>{d.name}</option>}
							</For>
						</select>
					</label>
					<button
						class="btn btn-primary btn-sm"
						disabled={loading() || !design()}
						onClick={() => void loadMatrix()}
					>
						Load schedule
					</button>
				</div>
				<Show when={error()}>
					<p role="alert" class="text-sm text-error">
						{error()}
					</p>
				</Show>
				<Show when={versionNotice()}>
					<p class="text-xs text-warning">{versionNotice()}</p>
				</Show>
				<Show when={notice()}>
					<p class="text-xs text-warning">{notice()}</p>
				</Show>
			</div>
			<Show
				when={matrix()}
				fallback={
					<div class="p-8 text-sm opacity-60">
						Select a study, version, and design to load its schedule.
					</div>
				}
			>
				<div class="px-4 py-2 border-b border-base-300 flex gap-4 flex-wrap items-center text-xs">
					<span>
						{visits().length} visits · {rows().length} assessments ·{" "}
						{matrix()?.data.instances.length} scheduled instances
					</span>
					<input
						aria-label="Filter assessments"
						class="input input-xs w-56"
						placeholder="Filter assessments"
						value={filter()}
						onInput={(e) => {
							setFilter(e.currentTarget.value);
							setRowPage(0);
							setSelected(null);
						}}
					/>
					<label class="flex items-center gap-2">
						<input
							type="checkbox"
							class="checkbox checkbox-xs"
							checked={labsOnly()}
							onChange={(e) => {
								setLabsOnly(e.currentTarget.checked);
								setRowPage(0);
								setSelected(null);
							}}
						/>
						Lab concepts only
					</label>
					<span class="opacity-60">
						● Scheduled · — No schedule in loaded data
					</span>
				</div>
				<For each={matrix()?.warnings}>
					{(w) => (
						<p role="status" class="px-4 py-1 text-xs text-warning">
							{w}
						</p>
					)}
				</For>
				<Show when={timeline()}>
					{(tl) => (
						<div class="border-b border-base-300">
							<button
								class="w-full px-4 py-1 text-xs text-left flex items-center gap-2 hover:bg-base-200"
								onClick={() => setTimelineOpen((o) => !o)}
							>
								<span
									class="inline-block transition-transform"
									style={{
										transform: timelineOpen()
											? "rotate(90deg)"
											: "rotate(0deg)",
									}}
								>
									▶
								</span>
								Timeline
								<Show when={tl().valid}>
									<span class="opacity-40 ml-1">
										Day {Math.min(...tl().visits.map((v) => v.day))} –{" "}
										{Math.max(...tl().visits.map((v) => v.day))}
									</span>
								</Show>
							</button>
							<Show when={timelineOpen()}>
								<div class="px-4 pb-3">
									<SoATimeline timeline={tl()} />
								</div>
							</Show>
						</div>
					)}
				</Show>
				<div class="flex flex-1 min-h-0">
					<div class="flex flex-col flex-1 min-w-0">
						<div class="flex flex-wrap gap-4 px-4 py-2 text-xs border-b border-base-300">
							<div class="flex items-center gap-2">
								<button
									aria-label="Previous assessments"
									class="btn btn-xs"
									disabled={rowPage() === 0}
									onClick={() => setRowPage((p) => p - 1)}
								>
									←
								</button>
								<span>
									Assessments {rows().length ? rowPage() * ROW_PAGE + 1 : 0}–
									{Math.min((rowPage() + 1) * ROW_PAGE, rows().length)} of{" "}
									{rows().length}
								</span>
								<button
									aria-label="Next assessments"
									class="btn btn-xs"
									disabled={(rowPage() + 1) * ROW_PAGE >= rows().length}
									onClick={() => setRowPage((p) => p + 1)}
								>
									→
								</button>
							</div>
							<div class="flex items-center gap-2">
								<button
									aria-label="Previous visits"
									class="btn btn-xs"
									disabled={visitPage() === 0}
									onClick={() => setVisitPage((p) => p - 1)}
								>
									←
								</button>
								<span>
									Visits {visits().length ? visitPage() * VISIT_PAGE + 1 : 0}–
									{Math.min((visitPage() + 1) * VISIT_PAGE, visits().length)} of{" "}
									{visits().length}
								</span>
								<button
									aria-label="Next visits"
									class="btn btn-xs"
									disabled={(visitPage() + 1) * VISIT_PAGE >= visits().length}
									onClick={() => setVisitPage((p) => p + 1)}
								>
									→
								</button>
							</div>
						</div>
						<div class="flex-1 overflow-auto">
							<table class="soa-table text-xs">
								<caption class="sr-only">
									Planned activities and lab concepts by ordered visit. Select a
									scheduled cell for details.
								</caption>
								<thead>
									<tr>
										<th scope="col">Assessment</th>
										<For each={shownVisits()}>
											{(v) => (
												<th scope="col">
													<div>{v.name}</div>
													<div class="font-normal opacity-60 mt-1">
														{v.modality.join(", ")}
													</div>
												</th>
											)}
										</For>
									</tr>
								</thead>
								<tbody>
									<For each={shownRows()}>
										{(row) => (
											<tr>
												<th scope="row">
													<div>{row.name}</div>
													<div class="opacity-50 font-normal mt-1">
														{row.kind}
													</div>
												</th>
												<For each={shownVisits()}>
													{(visit) => {
														const instances = () =>
															matrix()?.cells.get(visit.id)?.get(row.id) || [];
														return (
															<td>
																<Show
																	when={instances().length}
																	fallback={<span class="opacity-30">—</span>}
																>
																	<button
																		class={`btn btn-sm w-full ${selected()?.row.id === row.id && selected()?.visit.id === visit.id ? "btn-primary" : "btn-ghost text-primary"}`}
																		aria-label={`${row.name} at ${visit.name}: ${instances().length} scheduled instance(s)`}
																		onClick={() => setSelected({ row, visit })}
																	>
																		●
																		<Show when={instances().length > 1}>
																			{" "}
																			{instances().length}
																		</Show>
																	</button>
																</Show>
															</td>
														);
													}}
												</For>
											</tr>
										)}
									</For>
								</tbody>
							</table>
							<Show when={!rows().length || !visits().length}>
								<p class="p-8 text-sm opacity-60">
									No matching assessments or visits in this schedule.
								</p>
							</Show>
						</div>
					</div>
					<Show when={selected()}>
						{(selection) => (
							<aside
								class="w-80 max-w-[40%] shrink-0 border-l border-base-300 overflow-auto p-4 text-xs space-y-4"
								aria-label="Schedule cell details"
							>
								<div class="flex justify-between gap-2">
									<h3 class="font-semibold text-sm">{selection().row.name}</h3>
									<button
										class="btn btn-ghost btn-xs"
										aria-label="Close cell details"
										onClick={() => setSelected(null)}
									>
										✕
									</button>
								</div>
								<p>
									Visit {selection().visit.name} · {selection().row.kind}
								</p>
								<p>
									Epochs:{" "}
									{[
										...new Set(selectedInstances().flatMap((i) => i.epochs)),
									].join(", ") || "Not specified"}
								</p>
								<div>
									<h4 class="font-semibold mb-2">Timing relationships</h4>
									<For
										each={details().timings}
										fallback={
											<p class="opacity-60">
												No timing relationship in loaded data.
											</p>
										}
									>
										{(t) => (
											<div class="border border-base-300 rounded p-2 mb-2 space-y-1">
												<p>
													{timingVisitNames(t.from)} → {timingVisitNames(t.to)}
												</p>
												<p>Offset: {t.value || "Not specified"}</p>
												<p>
													Window:{" "}
													{t.windowLabel ||
														`lower ${t.windowLower || "unspecified"}, upper ${t.windowUpper || "unspecified"}`}
												</p>
												<p class="opacity-60">
													{[...t.type, ...t.relativeType].join(" · ")}
												</p>
											</div>
										)}
									</For>
								</div>
								<div>
									<h4 class="font-semibold mb-2">Applicable conditions</h4>
									<For
										each={details().conditions}
										fallback={
											<p class="opacity-60">
												No linked condition in loaded data.
											</p>
										}
									>
										{(c) => (
											<p class="border border-base-300 rounded p-2 mb-2 whitespace-pre-wrap">
												{c.text || "Condition has no text"}
											</p>
										)}
									</For>
								</div>
								<details>
									<summary class="cursor-pointer">Source instance IDs</summary>
									<For each={selectedInstances()}>
										{(i) => (
											<p class="font-mono break-all mt-2 opacity-60">{i.id}</p>
										)}
									</For>
								</details>
							</aside>
						)}
					</Show>
				</div>
			</Show>
		</section>
	);
}
