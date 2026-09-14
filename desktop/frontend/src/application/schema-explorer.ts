import type { SchemaEvent, SchemaSnapshot } from "../domain/schema";

export interface SchemaPort {
	get(options?: { refresh?: boolean }): Promise<SchemaSnapshot>;
	subscribe(listener: (event: SchemaEvent) => void): () => void;
}

export class SchemaExplorer {
	constructor(private readonly port: SchemaPort) {}

	load(): Promise<SchemaSnapshot> {
		return this.port.get();
	}

	refresh(): Promise<SchemaSnapshot> {
		return this.port.get({ refresh: true });
	}

	subscribe(listener: (event: SchemaEvent) => void): () => void {
		return this.port.subscribe(listener);
	}
}
