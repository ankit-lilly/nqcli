import { describe, expect, it, vi } from "vitest";

import { createBackupData, type LocalDb } from "./localDb";

describe("createBackupData", () => {
	it("captures every persisted entry", async () => {
		vi.stubGlobal("__NQ_WEBUI_VERSION__", "test-version");
		const values = new Map<string, unknown>([
			["schema", { vertices: [] }],
			["layout", { direction: "right" }],
		]);
		const db: LocalDb = {
			keys: async () => [...values.keys()],
			getItem: async <T>(key: string) => (values.get(key) as T) ?? null,
		};

		const backup = await createBackupData(db);

		expect(backup.backupSourceVersion).toBe("test-version");
		expect(backup.data).toStrictEqual(Object.fromEntries(values));
	});
});
