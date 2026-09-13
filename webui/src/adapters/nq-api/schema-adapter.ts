import type { OperationOptions, SchemaPort } from "@/application";
import type { Explorer } from "@/connector";
import type { SchemaSnapshot } from "@/domain";

export class NqSchemaAdapter implements SchemaPort {
	constructor(private readonly explorer: Explorer) {}

	get(options?: OperationOptions): Promise<SchemaSnapshot> {
		return options?.signal
			? this.explorer.fetchSchema({ signal: options.signal })
			: this.explorer.fetchSchema();
	}

	async revalidate(options?: OperationOptions): Promise<void> {
		const response = await fetch(`${this.explorer.connection.url}/schema`, {
			signal: options?.signal,
		});
		if (!response.ok && response.status !== 202) {
			throw new Error(
				`Schema revalidation failed with status ${response.status}`,
			);
		}
	}

	async refresh(options?: OperationOptions): Promise<void> {
		const response = await fetch(
			`${this.explorer.connection.url}/schema?refresh=true`,
			{ method: "POST", signal: options?.signal },
		);
		if (!response.ok) {
			throw new Error(`Schema refresh failed with status ${response.status}`);
		}
	}
}
