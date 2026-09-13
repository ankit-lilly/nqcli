import type { SchemaEventsPort } from "@/application";
import type { SchemaEvent } from "@/domain";

type EventSourceFactory = (url: string) => EventSource;

export class SseSchemaEventsAdapter implements SchemaEventsPort {
	constructor(
		private readonly baseUrl: string,
		private readonly createEventSource: EventSourceFactory = (url) =>
			new EventSource(url),
		private readonly onError: (error: unknown) => void = () => {},
	) {}

	subscribe(listener: (event: SchemaEvent) => void): () => void {
		const source = this.createEventSource(`${this.baseUrl}/schema/events`);
		const onSchema = (event: MessageEvent) => {
			try {
				listener(JSON.parse(event.data) as SchemaEvent);
			} catch (error) {
				this.onError(error);
			}
		};

		source.addEventListener("schema", onSchema);
		source.onerror = this.onError;

		return () => {
			source.removeEventListener("schema", onSchema);
			source.close();
		};
	}
}
