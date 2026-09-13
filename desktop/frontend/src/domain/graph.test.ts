import { expect, test } from "bun:test";
import {
	mergeGraphElements,
	graphTopologyKey,
	removeGraphElement,
	toGraphElements,
	type GraphElement,
} from "./graph";

const node = (id: string): GraphElement => ({
	group: "nodes",
	data: { id, label: "Study" },
});

test("graphTopologyKey ignores result order and non-topology properties", () => {
	const edge: GraphElement = {
		group: "edges",
		data: { id: "edge", source: "one", target: "two", label: "knows" },
	};
	const first = [node("one"), node("two"), edge];
	const second = [
		{ ...edge, data: { ...edge.data, weight: 2 } },
		{ ...node("two"), data: { ...node("two").data, name: "Two" } },
		node("one"),
	];

	expect(graphTopologyKey(first)).toBe(graphTopologyKey(second));
});

test("mergeGraphElements adds and refreshes entities by ID", () => {
	const current = [node("one")];
	const merged = mergeGraphElements(current, [
		{ group: "nodes", data: { id: "one", name: "Updated" } },
		node("two"),
	]);

	expect(merged).toHaveLength(2);
	expect(merged[0].data).toEqual({
		id: "one",
		label: "Study",
		name: "Updated",
	});
});

test("removeGraphElement removes a node and its connected edges", () => {
	const one = node("one");
	const elements: GraphElement[] = [
		one,
		node("two"),
		{ group: "edges", data: { id: "edge", source: "one", target: "two" } },
	];

	expect(removeGraphElement(elements, one)).toEqual([node("two")]);
});

test("toGraphElements rejects malformed bridge values", () => {
	expect(
		toGraphElements([
			{ group: "nodes", data: { id: "one" } },
			{ group: "vertices", data: { id: "bad" } },
			{ group: "edges", data: null },
		]),
	).toEqual([{ group: "nodes", data: { id: "one" } }]);
});
