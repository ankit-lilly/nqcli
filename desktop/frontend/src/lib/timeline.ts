import type { SoAData, Visit } from "./soa";

export interface TimelineVisit {
	visitId: string;
	name: string;
	day: number;
	windowLowerDays: number;
	windowUpperDays: number;
	epochs: string[];
}

export interface EpochBand {
	name: string;
	startDay: number;
	endDay: number;
}

export interface TimelineResult {
	visits: TimelineVisit[];
	epochBands: EpochBand[];
	valid: boolean;
}

const DURATION_RE = /^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T.*)?$/;

export function parseDurationDays(iso: string): number | null {
	if (!iso) return null;
	const m = DURATION_RE.exec(iso);
	if (!m) return null;
	const y = parseInt(m[1] || "0", 10);
	const mo = parseInt(m[2] || "0", 10);
	const w = parseInt(m[3] || "0", 10);
	const d = parseInt(m[4] || "0", 10);
	return y * 365 + mo * 30 + w * 7 + d;
}

export function buildTimeline(
	data: SoAData,
	orderedVisits: Visit[],
): TimelineResult {
	if (!orderedVisits.length)
		return { visits: [], epochBands: [], valid: false };

	// Stage 1: Instance-to-visit index
	const instanceToVisits = new Map<string, string[]>();
	for (const inst of data.instances) {
		instanceToVisits.set(inst.id, inst.visits);
	}

	// Stage 2: Visit-to-visit offset graph
	interface Edge {
		to: string;
		offsetDays: number;
		lowerDays: number;
		upperDays: number;
	}
	const graph = new Map<string, Edge[]>();
	const hasInbound = new Set<string>();

	for (const timing of data.timings) {
		const offset = parseDurationDays(timing.value);
		if (offset === null) continue;
		const lower = parseDurationDays(timing.windowLower) || 0;
		const upper = parseDurationDays(timing.windowUpper) || 0;

		const fromVisits = new Set<string>();
		for (const instId of timing.from) {
			for (const v of instanceToVisits.get(instId) || []) fromVisits.add(v);
		}
		const toVisits = new Set<string>();
		for (const instId of timing.to) {
			for (const v of instanceToVisits.get(instId) || []) toVisits.add(v);
		}

		for (const fv of fromVisits) {
			for (const tv of toVisits) {
				if (fv === tv) continue;
				let edges = graph.get(fv);
				if (!edges) {
					edges = [];
					graph.set(fv, edges);
				}
				edges.push({
					to: tv,
					offsetDays: offset,
					lowerDays: lower,
					upperDays: upper,
				});
				hasInbound.add(tv);
			}
		}
	}

	// Stage 3: BFS from anchor
	const visitOrder = orderedVisits.map((v) => v.id);
	const anchor =
		visitOrder.find((id) => graph.has(id) && !hasInbound.has(id)) ||
		visitOrder[0];

	const dayMap = new Map<string, number>();
	const windowMap = new Map<string, { lower: number; upper: number }>();
	dayMap.set(anchor, 0);

	const queue = [anchor];
	while (queue.length) {
		const current = queue.shift()!;
		const currentDay = dayMap.get(current)!;
		for (const edge of graph.get(current) || []) {
			if (dayMap.has(edge.to)) continue;
			dayMap.set(edge.to, currentDay + edge.offsetDays);
			windowMap.set(edge.to, { lower: -edge.lowerDays, upper: edge.upperDays });
			queue.push(edge.to);
		}
	}

	const valid = dayMap.size > 1;

	// Stage 4: Interpolate unpositioned visits
	const positions: (number | null)[] = visitOrder.map(
		(id) => dayMap.get(id) ?? null,
	);

	if (!valid) {
		for (let i = 0; i < positions.length; i++) positions[i] = i * 7;
	} else {
		let i = 0;
		while (i < positions.length) {
			if (positions[i] !== null) {
				i++;
				continue;
			}
			let start = i;
			while (i < positions.length && positions[i] === null) i++;
			const gapLen = i - start;
			const leftDay = start > 0 ? positions[start - 1]! : null;
			const rightDay = i < positions.length ? positions[i]! : null;
			if (leftDay !== null && rightDay !== null) {
				const step = (rightDay - leftDay) / (gapLen + 1);
				for (let j = 0; j < gapLen; j++)
					positions[start + j] = leftDay + step * (j + 1);
			} else if (leftDay !== null) {
				for (let j = 0; j < gapLen; j++)
					positions[start + j] = leftDay + 7 * (j + 1);
			} else if (rightDay !== null) {
				for (let j = gapLen - 1; j >= 0; j--)
					positions[start + j] = rightDay - 7 * (gapLen - j);
			} else {
				for (let j = 0; j < gapLen; j++) positions[start + j] = j * 7;
			}
		}
	}

	// Stage 5: Build epoch bands and visit results
	const visitEpochs = new Map<string, string[]>();
	for (const inst of data.instances) {
		for (const vid of inst.visits) {
			const existing = visitEpochs.get(vid) || [];
			existing.push(...inst.epochs);
			visitEpochs.set(vid, existing);
		}
	}

	const byId = new Map(orderedVisits.map((v) => [v.id, v]));
	const timelineVisits: TimelineVisit[] = visitOrder.map((id, i) => {
		const v = byId.get(id)!;
		const epochs = [...new Set(visitEpochs.get(id) || [])];
		const win = windowMap.get(id);
		return {
			visitId: id,
			name: v.name,
			day: positions[i]!,
			windowLowerDays: win?.lower ?? 0,
			windowUpperDays: win?.upper ?? 0,
			epochs,
		};
	});

	const epochBands: EpochBand[] = [];
	for (const tv of timelineVisits) {
		const epoch = tv.epochs[0];
		if (!epoch) continue;
		const last = epochBands[epochBands.length - 1];
		if (last && last.name === epoch) {
			last.endDay = tv.day;
		} else {
			epochBands.push({ name: epoch, startDay: tv.day, endDay: tv.day });
		}
	}

	return { visits: timelineVisits, epochBands, valid };
}
