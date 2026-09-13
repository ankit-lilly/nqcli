import type { SchemaSnapshot } from "../domain/schema";

export interface SchemaPort {
	get(options?: { refresh?: boolean }): Promise<SchemaSnapshot>;
}

export class SchemaExplorer {
	constructor(private readonly port: SchemaPort) {}

	load(): Promise<SchemaSnapshot> {
		return this.port.get();
	}

	refresh(): Promise<SchemaSnapshot> {
		return this.port.get({ refresh: true });
	}
}
