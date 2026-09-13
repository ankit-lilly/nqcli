/* oxlint-disable no-console */
import { beforeEach, describe, expect, test, vi } from "vitest";

vi.unmock("@/utils/logger");

describe("logger", () => {
	beforeEach(() => {
		vi.resetModules();
		vi.spyOn(console, "debug").mockImplementation(() => {});
		vi.spyOn(console, "log").mockImplementation(() => {});
		vi.spyOn(console, "warn").mockImplementation(() => {});
		vi.spyOn(console, "error").mockImplementation(() => {});
	});

	test("suppresses diagnostic output in production", async () => {
		vi.doMock("./env", () => ({ env: { DEV: false, PROD: true } }));
		const { default: logger } = await import("./logger");

		logger.debug("debug");
		logger.log("log");
		logger.warn("warn");
		logger.error("error");

		expect(console.debug).not.toHaveBeenCalled();
		expect(console.log).not.toHaveBeenCalled();
		expect(console.warn).toHaveBeenCalledWith("warn");
		expect(console.error).toHaveBeenCalledWith("error");
	});

	test("emits diagnostic output in development", async () => {
		vi.doMock("./env", () => ({ env: { DEV: true, PROD: false } }));
		const { default: logger } = await import("./logger");

		logger.debug("debug");
		logger.log("log");

		expect(console.debug).toHaveBeenCalledWith("debug");
		expect(console.log).toHaveBeenCalledWith("log");
	});
});
