import { expect, test } from "bun:test";
import { formatGremlin, formatCypher } from "./formatQuery";

test("formats a simple gremlin chain", () => {
	const input = "g.V().hasLabel('Study').limit(20)";
	const lines = formatGremlin(input).split("\n");
	expect(lines[0]).toBe("g.");
	expect(lines[1]).toBe("V().");
	expect(lines[2]).toBe("hasLabel('Study').");
	expect(lines[3]).toBe("limit(20)");
});

test("formats nested gremlin with union", () => {
	const input =
		"g.V().hasLabel('Study').limit(1).union(outE('has_version').inV().path(),outE('has_latest_version').inV().path())";
	const formatted = formatGremlin(input);
	expect(formatted).toContain("union(");
	expect(formatted).toContain("outE('has_version')");
	expect(formatted.split("\n").length).toBeGreaterThan(3);
});

test("formats balanced parentheses", () => {
	const input = "g.V().has('name','test').out('knows').path()";
	const formatted = formatGremlin(input);
	const opens = (formatted.match(/\(/g) || []).length;
	const closes = (formatted.match(/\)/g) || []).length;
	expect(opens).toBe(closes);
});

test("formats cypher with keywords on new lines", () => {
	const input = "MATCH (n:Study) WHERE n.name = 'test' RETURN n LIMIT 10";
	const formatted = formatCypher(input);
	expect(formatted).toContain("\nWHERE");
	expect(formatted).toContain("\nRETURN");
	expect(formatted).toContain("\nLIMIT");
});
