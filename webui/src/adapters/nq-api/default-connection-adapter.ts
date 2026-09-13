import * as z from "zod";
import type { ConnectionPort, OperationOptions } from "@/application";
import { type DefaultConnection, queryEngineOptions } from "@/domain";

export const DefaultConnectionSchema = z.object({
	apiBaseUrl: z.string().url(),
	queryEngine: z.enum(queryEngineOptions).default("gremlin"),
	profile: z.string().default(""),
});

export class NqDefaultConnectionAdapter implements ConnectionPort {
	constructor(private readonly baseUrl = location.origin) {}

	async getDefault(options?: OperationOptions): Promise<DefaultConnection> {
		const url = `${this.baseUrl}/defaultConnection`;
		const response = options?.signal
			? await fetch(url, { signal: options.signal })
			: await fetch(url);
		if (!response.ok) {
			throw new Error(`Default connection returned ${response.status}`);
		}
		return DefaultConnectionSchema.parse(await response.json());
	}
}
