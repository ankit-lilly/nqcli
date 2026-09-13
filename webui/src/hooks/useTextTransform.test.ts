// @vitest-environment happy-dom
import { renderHookWithJotai } from "@/utils/testing";

import useTextTransform from "./useTextTransform";

describe("useTextTransform", () => {
	it("should not modify text", () => {
		const text = "this is a test";
		const { result } = renderHookWithJotai(() => useTextTransform());
		expect(result.current(text)).toBe(text);
	});

	it("should return the original text if no transformation is needed", () => {
		const text = "This Is A Test";
		const { result } = renderHookWithJotai(() => useTextTransform());
		expect(result.current(text)).toBe(text);
	});

	it("should handle empty string", () => {
		const input = "";
		const { result } = renderHookWithJotai(() => useTextTransform());
		expect(result.current(input)).toBe(input);
	});

	it("should handle strings with invalid characters", () => {
		const input = "str\u{1F600}";
		const { result } = renderHookWithJotai(() => useTextTransform());
		expect(result.current(input)).toBe(input);
	});

	it("should return original input for URI-like strings", () => {
		const input = "http://www.some-uri.com/";
		const { result } = renderHookWithJotai(() => useTextTransform());
		expect(result.current(input)).toBe(input);
	});

	// Random data
	it("should handle random data", () => {
		const input = "str\u{1F600}abcdef";
		const { result } = renderHookWithJotai(() => useTextTransform());
		expect(result.current(input)).toBe(input);
	});
});
