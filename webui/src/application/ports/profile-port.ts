import type { ProfileInfo, ProfileSelection } from "@/domain";

import type { OperationOptions } from "./operation";

export interface ProfilePort {
	get(options?: OperationOptions): Promise<ProfileInfo>;
	switchTo(
		profile: string,
		options?: OperationOptions,
	): Promise<ProfileSelection>;
}
