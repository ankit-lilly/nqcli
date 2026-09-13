import type { SchemaEvent, SchemaSnapshot } from "@/domain";

import type { OperationOptions } from "./operation";

export interface SchemaPort {
	get(options?: OperationOptions): Promise<SchemaSnapshot>;
	/** Starts stale-while-revalidate without replacing a usable snapshot. */
	revalidate(options?: OperationOptions): Promise<void>;
	refresh(options?: OperationOptions): Promise<void>;
}

export interface SchemaEventsPort {
	subscribe(listener: (event: SchemaEvent) => void): () => void;
}
