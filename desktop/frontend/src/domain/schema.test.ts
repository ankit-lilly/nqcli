import { expect, test } from "bun:test";
import { schemaGraphElements, type SchemaSnapshot } from "./schema";

test("schema graph drops connections whose endpoint was not discovered", () => {
	const snapshot: SchemaSnapshot = {
		status: "ready",
		vertices: [
			{ type: "Study", attributes: [], total: 2 },
			{ type: "StudyVersion", attributes: [], total: 4 },
		],
		edges: [],
		edgeConnections: [
			{
				sourceVertexType: "Study",
				edgeType: "has_version",
				targetVertexType: "StudyVersion",
			},
			{
				sourceVertexType: "Missing",
				edgeType: "bad",
				targetVertexType: "Study",
			},
		],
	};

	const elements = schemaGraphElements(snapshot);
	expect(elements.filter((element) => element.group === "nodes")).toHaveLength(
		2,
	);
	expect(elements.filter((element) => element.group === "edges")).toHaveLength(
		1,
	);
});
