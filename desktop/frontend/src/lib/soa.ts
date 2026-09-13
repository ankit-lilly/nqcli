export interface Named {
	id: string;
	name: string;
}
export interface Visit extends Named {
	sourceId?: string;
	previousId: string;
	nextId: string;
	modality: string[];
}
export interface Instance {
	id: string;
	visits: string[];
	activities: string[];
	epochs: string[];
}
export interface Timing {
	id: string;
	value: string;
	windowLower: string;
	windowUpper: string;
	windowLabel: string;
	from: string[];
	to: string[];
	type: string[];
	relativeType: string[];
}
export interface Condition {
	id: string;
	text: string;
	activities: string[];
	instances: string[];
}
export interface SoAData {
	visits: Visit[];
	activities: Named[];
	labs: Named[];
	instances: Instance[];
	timings: Timing[];
	conditions: Condition[];
}
export interface Version extends Named {
	status: string;
	amendment: string[];
	designs: Named[];
}
export interface Versions {
	latest: string[];
	versions: Version[];
}
export interface Row extends Named {
	kind: "Activity" | "Lab concept";
}

// Follow explicit visit links. Preserve all vertices and report malformed or partial chains.
export function orderVisits(visits: Visit[]): {
	visits: Visit[];
	warning: string;
} {
	const sourceIds = new Map(
		visits.filter((v) => v.sourceId).map((v) => [v.sourceId!, v.id]),
	);
	visits = visits.map((v) => ({
		...v,
		previousId: sourceIds.get(v.previousId) || v.previousId,
		nextId: sourceIds.get(v.nextId) || v.nextId,
	}));
	const byId = new Map(visits.map((v) => [v.id, v]));
	const seen = new Set<string>();
	const ordered: Visit[] = [];
	let broken = false;
	const walk = (start: Visit) => {
		let current: Visit | undefined = start;
		while (current && !seen.has(current.id)) {
			seen.add(current.id);
			ordered.push(current);
			if (!current.nextId) break;
			const next: Visit | undefined = byId.get(current.nextId);
			if (!next || next.previousId !== current.id || seen.has(next.id))
				broken = true;
			current = next;
		}
	};
	const roots = visits.filter((v) => !v.previousId || !byId.has(v.previousId));
	for (const root of roots) {
		if (root.previousId) broken = true;
		walk(root);
	}
	for (const v of visits)
		if (!seen.has(v.id)) {
			broken = true;
			walk(v);
		}
	if (visits.length && roots.length !== 1) broken = true;
	return {
		visits: ordered,
		warning: broken
			? "Visit ordering contains missing links, separate chains, or a cycle; check the source schedule."
			: "",
	};
}

export function prepareSoA(input: SoAData) {
	const limits: Record<keyof SoAData, number> = {
		visits: 1000,
		activities: 1000,
		labs: 1000,
		instances: 2000,
		timings: 2000,
		conditions: 500,
	};
	const data = {} as SoAData;
	const warnings: string[] = [];
	for (const key of Object.keys(limits) as (keyof SoAData)[]) {
		const values = input[key] || [];
		if (values.length > limits[key])
			warnings.push(
				`${key}: showing first ${limits[key]}. This schedule is incomplete.`,
			);
		(data[key] as unknown[]) = values.slice(0, limits[key]);
	}
	const ordered = orderVisits(data.visits);
	data.visits = ordered.visits;
	if (ordered.warning) warnings.push(ordered.warning);
	const rows: Row[] = [
		...data.activities.map((a) => ({ ...a, kind: "Activity" as const })),
		...data.labs.map((a) => ({ ...a, kind: "Lab concept" as const })),
	];
	const cells = new Map<string, Map<string, Instance[]>>();
	const knownRows = new Set(rows.map((r) => r.id));
	const knownVisits = new Set(data.visits.map((v) => v.id));
	let missing = false;
	for (const instance of data.instances) {
		for (const visit of new Set(instance.visits)) {
			if (!knownVisits.has(visit)) missing = true;
			let entries = cells.get(visit);
			if (!entries) {
				entries = new Map();
				cells.set(visit, entries);
			}
			for (const activity of new Set(instance.activities)) {
				if (!knownRows.has(activity)) missing = true;
				const existing = entries.get(activity) || [];
				existing.push(instance);
				entries.set(activity, existing);
			}
		}
	}
	if (missing)
		warnings.push(
			"Some scheduled instances reference visits or assessments outside the loaded collections.",
		);
	return { data, rows, cells, warnings };
}

export function cellDetails(
	data: SoAData,
	instances: Instance[],
	rowId: string,
) {
	const ids = new Set(instances.map((i) => i.id));
	return {
		timings: data.timings.filter((t) =>
			[...t.from, ...t.to].some((id) => ids.has(id)),
		),
		conditions: data.conditions.filter((c) => {
			const activityMatch =
				c.activities.length === 0 || c.activities.includes(rowId);
			const contextMatch =
				c.instances.length === 0 || c.instances.some((id) => ids.has(id));
			return (
				activityMatch &&
				contextMatch &&
				(c.activities.length > 0 || c.instances.length > 0)
			);
		}),
	};
}
