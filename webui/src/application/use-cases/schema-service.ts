import type { SchemaEvent } from "@/domain";

import type { OperationOptions, SchemaEventsPort, SchemaPort } from "../ports";

export class SchemaService {
	constructor(
		private readonly schema: SchemaPort,
		private readonly events?: SchemaEventsPort,
	) {}

	get(options?: OperationOptions) {
		return this.schema.get(options);
	}

	revalidate(options?: OperationOptions) {
		return this.schema.revalidate(options);
	}

	refresh(options?: OperationOptions) {
		return this.schema.refresh(options);
	}

	subscribe(listener: (event: SchemaEvent) => void): () => void {
		return this.events?.subscribe(listener) ?? (() => {});
	}
}
