import { describe, expect, it } from "vitest";

import {
	allIconNamesSorted,
	getLucideName,
	getLucideSvgString,
	isValidLucideIconName,
	toLucideIconRef,
} from "./lucideIcons";

describe("lucide icon registry", () => {
	it("stores curated icons as symbolic references", () => {
		expect(toLucideIconRef("book-open")).toBe("lucide:book-open");
		expect(getLucideName("lucide:book-open")).toBe("book-open");
		expect(isValidLucideIconName("book-open")).toBe(true);
		expect(isValidLucideIconName("not-curated")).toBe(false);
	});

	it("renders every curated icon to SVG", async () => {
		for (const name of allIconNamesSorted) {
			expect(await getLucideSvgString(name)).toContain("<svg");
		}
	});
});
