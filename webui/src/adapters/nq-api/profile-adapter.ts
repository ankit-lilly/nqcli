import type { OperationOptions, ProfilePort } from "@/application";
import type { ProfileInfo, ProfileSelection } from "@/domain";

type ProfileResponse = ProfileSelection & { error?: string };

export class NqProfileAdapter implements ProfilePort {
	constructor(private readonly baseUrl = "") {}

	async get(options?: OperationOptions): Promise<ProfileInfo> {
		const url = `${this.baseUrl}/profiles`;
		const response = options?.signal
			? await fetch(url, { signal: options.signal })
			: await fetch(url);
		if (!response.ok) {
			throw new Error(`Profile endpoint returned ${response.status}`);
		}

		const info = (await response.json()) as ProfileInfo;
		return { ...info, profiles: info.profiles ?? [] };
	}

	async switchTo(
		profile: string,
		options?: OperationOptions,
	): Promise<ProfileSelection> {
		const response = await fetch(`${this.baseUrl}/profiles/current`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ profile }),
			signal: options?.signal,
		});
		const payload = (await response.json()) as ProfileResponse;
		if (!response.ok || payload.error) {
			throw new Error(
				payload.error || `Profile switch returned ${response.status}`,
			);
		}
		return payload;
	}
}
