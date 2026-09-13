import { test, expect } from "bun:test";
import {
	orderVisits,
	prepareSoA,
	cellDetails,
	type SoAData,
	type Visit,
} from "./soa";
const visit = (id: string, previousId = "", nextId = ""): Visit => ({
	id,
	name: id,
	previousId,
	nextId,
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
test("uses visit links rather than names or response order", () => {
	const result = orderVisits([
		visit("10", "2"),
		visit("1", "", "2"),
		visit("2", "1", "10"),
	]);
	expect(result.visits.map((v) => v.id)).toEqual(["1", "2", "10"]);
	expect(result.warning).toBe("");
});
test("cycles and missing links retain all visits and warn", () => {
	for (const list of [
		[visit("a", "b", "b"), visit("b", "a", "a")],
		[visit("a", "missing")],
	]) {
		const result = orderVisits(list);
		expect(result.visits.length).toBe(list.length);
		expect(result.warning).not.toBe("");
	}
});
test("one instance can schedule multiple activities and labs without conflating names", () => {
	const data = empty();
	data.visits = [visit("v")];
	data.activities = [{ id: "a", name: "Blood" }];
	data.labs = [{ id: "b", name: "Blood" }];
	data.instances = [
		{
			id: "i",
			visits: ["v"],
			activities: ["a", "b", "a"],
			epochs: ["Screening"],
		},
	];
	const result = prepareSoA(data);
	expect(result.cells.get("v")?.get("a")?.length).toBe(1);
	expect(result.cells.get("v")?.get("b")?.length).toBe(1);
	expect(result.rows.length).toBe(2);
});
test("conditions respect both assessment and instance scope", () => {
	const data = empty();
	const instance = { id: "i", visits: ["v"], activities: ["a"], epochs: [] };
	data.conditions = [
		{ id: "yes", text: "Fasting", activities: ["a"], instances: ["i"] },
		{ id: "other", text: "Other visit", activities: ["a"], instances: ["j"] },
		{ id: "global", text: "", activities: [], instances: [] },
	];
	data.timings = [
		{
			id: "t",
			from: ["anchor"],
			to: ["i"],
			value: "P4W",
			windowLower: "P3D",
			windowUpper: "P3D",
			windowLabel: "",
			type: [],
			relativeType: [],
		},
	];
	const result = cellDetails(data, [instance], "a");
	expect(result.conditions.map((c) => c.id)).toEqual(["yes"]);
	expect(result.timings.length).toBe(1);
});
test("truncated collections are explicitly marked incomplete", () => {
	const data = empty();
	data.visits = Array.from({ length: 1001 }, (_, i) => visit(String(i)));
	const result = prepareSoA(data);
	expect(result.data.visits.length).toBe(1000);
	expect(result.warnings.join(" ")).toContain("incomplete");
});
test("resolves source UUID links when Neptune vertex IDs differ", () => {
	const result = orderVisits([
		{ ...visit("n2", "s1"), sourceId: "s2" },
		{ ...visit("n1", "", "s2"), sourceId: "s1" },
	]);
	expect(result.visits.map((v) => v.id)).toEqual(["n1", "n2"]);
	expect(result.warning).toBe("");
});
