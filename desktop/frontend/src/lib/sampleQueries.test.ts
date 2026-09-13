import { expect, test } from "bun:test";
import { SAMPLE_QUERIES, buildSampleQuery } from "./sampleQueries";

function hasBalancedParentheses(query: string) {
	let depth = 0;
	for (const ch of query) {
		if (ch === "(") depth++;
		if (ch === ")") depth--;
		if (depth < 0) return false;
	}
	return depth === 0;
}

test("preserves every sample query", () => {
	expect(SAMPLE_QUERIES.map((q) => q.name)).toEqual([
		"Study Overview",
		"Study + Versions",
		"Full Trial Graph",
		"Epochs & Encounters",
		"Timelines & Timing",
	]);
});

test("all queries render as graph view", () => {
	for (const sample of SAMPLE_QUERIES) {
		expect(sample.view).toBe("graph");
	}
});

test("trial samples substitute the study placeholder", () => {
	for (const sample of SAMPLE_QUERIES.filter((q) => q.usesTrial)) {
		expect(buildSampleQuery(sample)).toStartWith(
			"g.V().hasLabel('Study').limit(1).",
		);
		expect(buildSampleQuery(sample)).not.toContain("{{study}}");
	}
});

test("current graph previews emit the study-to-version path explicitly", () => {
	for (const name of [
		"Full Trial Graph",
		"Epochs & Encounters",
		"Timelines & Timing",
	]) {
		const query = buildSampleQuery(
			SAMPLE_QUERIES.find((q) => q.name === name)!,
		);
		expect(query).toContain("coalesce(outE('has_latest_version').inV().path()");
		expect(query).toContain(
			"coalesce(out('has_latest_version'),out('has_version').order().by('createdAt',desc).limit(1))",
		);
		expect(query).toContain("union(identity().path()");
		expect(hasBalancedParentheses(query)).toBe(true);
	}
});

test("all queries have balanced parentheses", () => {
	for (const sample of SAMPLE_QUERIES) {
		const query = buildSampleQuery(sample);
		expect(hasBalancedParentheses(query)).toBe(true);
	}
});
