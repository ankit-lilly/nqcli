// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";

import { parseExportedGraph } from "@/modules/GraphViewer/exportedGraph";

import graphExportDecimalString from "./__fixtures__/graph-export-v1-decimal-string.json?raw";

/**
 * GOLDEN FILES — REAL EXPORTED FORMATS
 *
 * Each fixture in `__fixtures__/` is a byte-for-byte file as a shipped build
 * wrote it, loaded here with `?raw` so the test imports the exact on-disk bytes
 * (not a re-serialized object). The version integer in `fileEnvelope.ts` means
 * nothing on its own; these are what actually prove a current build still
 * imports every graph export generation we have ever written — including
 * graph-export's `"1.0"` decimal-string encoding, the wire form it still
 * writes today.
 *
 * DO NOT edit a fixture to make a test pass. A fixture is a historical artifact;
 * if a current build can no longer read it, that is a backward-compatibility
 * break to fix in the parser (add a migration / version case), not in the file.
 * When a new generation ships, ADD a fixture — never mutate an existing one.
 */

function asFile(contents: string, name: string): File {
	return new File([contents], name, { type: "application/json" });
}

describe("golden graph-export files import on the current build", () => {
	test("graph-export-v1-decimal-string.json", async () => {
		const parsed = await parseExportedGraph(
			asFile(graphExportDecimalString, "graph-export-v1-decimal-string.json"),
		);

		expect(parsed.connection).toStrictEqual({
			dbUrl: "https://example.cluster-abc.us-west-2.neptune.amazonaws.com:8182",
			queryEngine: "gremlin",
		});
		expect(parsed.vertices).toStrictEqual(new Set(["1", "2", 3]));
		expect(parsed.edges).toStrictEqual(new Set(["10", "11"]));
	});
});
