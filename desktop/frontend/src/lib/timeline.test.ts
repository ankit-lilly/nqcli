import { test, expect } from "bun:test";
import { parseDurationDays, buildTimeline } from "./timeline";
import type { SoAData, Visit } from "./soa";

const visit = (id: string, prev = "", next = ""): Visit => ({
	id,
	name: id,
	previousId: prev,
	nextId: next,
	modality: [],
});
const empty = (): SoAData => ({
	visits: [],
	activities: [],
	labs: [],
	instances: [],
	timings: [],
	conditions: [],
});

test("parseDurationDays handles standard ISO 8601 durations", () => {
	expect(parseDurationDays("P4W")).toBe(28);
	expect(parseDurationDays("P3D")).toBe(3);
	expect(parseDurationDays("P0D")).toBe(0);
	expect(parseDurationDays("P1Y2M3W4D")).toBe(365 + 60 + 21 + 4);
	expect(parseDurationDays("P")).toBe(0);
});

test("parseDurationDays ignores sub-day components", () => {
	expect(parseDurationDays("PT2H")).toBe(0);
	expect(parseDurationDays("P1DT6H")).toBe(1);
});

test("parseDurationDays returns null for non-ISO values", () => {
	expect(parseDurationDays("")).toBeNull();
	expect(parseDurationDays("Screening")).toBeNull();
	expect(parseDurationDays("Week 4")).toBeNull();
});

test("simple chain produces correct day positions", () => {
	const data = empty();
	data.instances = [
		{ id: "i1", visits: ["v1"], activities: [], epochs: [] },
		{ id: "i2", visits: ["v2"], activities: [], epochs: [] },
		{ id: "i3", visits: ["v3"], activities: [], epochs: [] },
	];
	data.timings = [
		{
			id: "t1",
			from: ["i1"],
			to: ["i2"],
			value: "P2W",
			windowLower: "",
			windowUpper: "",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
		{
			id: "t2",
			from: ["i2"],
			to: ["i3"],
			value: "P4W",
			windowLower: "",
			windowUpper: "",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
	];
	const visits = [
		visit("v1", "", "v2"),
		visit("v2", "v1", "v3"),
		visit("v3", "v2"),
	];
	const result = buildTimeline(data, visits);
	expect(result.valid).toBe(true);
	expect(result.visits.map((v) => v.day)).toEqual([0, 14, 42]);
});

test("window bounds are parsed into days", () => {
	const data = empty();
	data.instances = [
		{ id: "i1", visits: ["v1"], activities: [], epochs: [] },
		{ id: "i2", visits: ["v2"], activities: [], epochs: [] },
	];
	data.timings = [
		{
			id: "t1",
			from: ["i1"],
			to: ["i2"],
			value: "P2W",
			windowLower: "P3D",
			windowUpper: "P5D",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
	];
	const visits = [visit("v1", "", "v2"), visit("v2", "v1")];
	const result = buildTimeline(data, visits);
	expect(result.visits[1].windowLowerDays).toBe(-3);
	expect(result.visits[1].windowUpperDays).toBe(5);
});

test("unparseable timing values produce evenly spaced fallback", () => {
	const data = empty();
	data.instances = [
		{ id: "i1", visits: ["v1"], activities: [], epochs: [] },
		{ id: "i2", visits: ["v2"], activities: [], epochs: [] },
	];
	data.timings = [
		{
			id: "t1",
			from: ["i1"],
			to: ["i2"],
			value: "Screening to Treatment",
			windowLower: "",
			windowUpper: "",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
	];
	const visits = [visit("v1", "", "v2"), visit("v2", "v1")];
	const result = buildTimeline(data, visits);
	expect(result.valid).toBe(false);
	expect(result.visits[0].day).toBe(0);
	expect(result.visits[1].day).toBe(7);
});

test("empty timings produce evenly spaced visits", () => {
	const data = empty();
	const visits = [
		visit("v1", "", "v2"),
		visit("v2", "v1", "v3"),
		visit("v3", "v2"),
	];
	const result = buildTimeline(data, visits);
	expect(result.valid).toBe(false);
	expect(result.visits.map((v) => v.day)).toEqual([0, 7, 14]);
});

test("epoch bands group consecutive visits", () => {
	const data = empty();
	data.instances = [
		{ id: "i1", visits: ["v1"], activities: [], epochs: ["Screening"] },
		{ id: "i2", visits: ["v2"], activities: [], epochs: ["Treatment"] },
		{ id: "i3", visits: ["v3"], activities: [], epochs: ["Treatment"] },
		{ id: "i4", visits: ["v4"], activities: [], epochs: ["Follow-up"] },
	];
	data.timings = [
		{
			id: "t1",
			from: ["i1"],
			to: ["i2"],
			value: "P2W",
			windowLower: "",
			windowUpper: "",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
		{
			id: "t2",
			from: ["i2"],
			to: ["i3"],
			value: "P4W",
			windowLower: "",
			windowUpper: "",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
		{
			id: "t3",
			from: ["i3"],
			to: ["i4"],
			value: "P8W",
			windowLower: "",
			windowUpper: "",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
	];
	const visits = [
		visit("v1", "", "v2"),
		visit("v2", "v1", "v3"),
		visit("v3", "v2", "v4"),
		visit("v4", "v3"),
	];
	const result = buildTimeline(data, visits);
	expect(result.epochBands).toEqual([
		{ name: "Screening", startDay: 0, endDay: 0 },
		{ name: "Treatment", startDay: 14, endDay: 42 },
		{ name: "Follow-up", startDay: 98, endDay: 98 },
	]);
});

test("gap visits are interpolated between positioned neighbors", () => {
	const data = empty();
	data.instances = [
		{ id: "i1", visits: ["v1"], activities: [], epochs: [] },
		{ id: "i3", visits: ["v3"], activities: [], epochs: [] },
	];
	data.timings = [
		{
			id: "t1",
			from: ["i1"],
			to: ["i3"],
			value: "P4W",
			windowLower: "",
			windowUpper: "",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
	];
	const visits = [
		visit("v1", "", "v2"),
		visit("v2", "v1", "v3"),
		visit("v3", "v2"),
	];
	const result = buildTimeline(data, visits);
	expect(result.valid).toBe(true);
	expect(result.visits[0].day).toBe(0);
	expect(result.visits[1].day).toBeCloseTo(14, 0);
	expect(result.visits[2].day).toBe(28);
});
