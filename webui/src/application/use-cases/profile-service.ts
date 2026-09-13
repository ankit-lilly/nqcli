import type { OperationOptions, ProfilePort } from "../ports";

export class ProfileService {
	constructor(private readonly profiles: ProfilePort) {}

	get(options?: OperationOptions) {
		return this.profiles.get(options);
	}

	switchTo(profile: string, options?: OperationOptions) {
		return this.profiles.switchTo(profile, options);
	}
}
