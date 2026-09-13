import { SseSchemaEventsAdapter } from "./sse-schema-events";

class FakeEventSource {
	onerror: ((event: Event) => void) | null = null;
	readonly listeners = new Map<string, EventListener>();
	closed = false;

	addEventListener(type: string, listener: EventListener): void {
		this.listeners.set(type, listener);
	}

	removeEventListener(type: string): void {
		this.listeners.delete(type);
	}

	close(): void {
		this.closed = true;
	}

	emit(type: string, data: string): void {
		this.listeners.get(type)?.(new MessageEvent(type, { data }));
	}
}

describe("SseSchemaEventsAdapter", () => {
	it("maps schema events and releases the browser connection", () => {
		const source = new FakeEventSource();
		const listener = vi.fn();
		const adapter = new SseSchemaEventsAdapter(
			"/api",
			() => source as unknown as EventSource,
		);

		const unsubscribe = adapter.subscribe(listener);
		source.emit(
			"schema",
			JSON.stringify({ type: "schema.sync.completed", status: "ready" }),
		);

		expect(listener).toHaveBeenCalledWith({
			type: "schema.sync.completed",
			status: "ready",
		});
		unsubscribe();
		expect(source.closed).toBe(true);
		expect(source.listeners.has("schema")).toBe(false);
	});

	it("reports invalid event payloads without notifying subscribers", () => {
		const source = new FakeEventSource();
		const listener = vi.fn();
		const onError = vi.fn();
		const adapter = new SseSchemaEventsAdapter(
			"/api",
			() => source as unknown as EventSource,
			onError,
		);

		adapter.subscribe(listener);
		source.emit("schema", "not-json");

		expect(listener).not.toHaveBeenCalled();
		expect(onError).toHaveBeenCalledOnce();
	});
});
