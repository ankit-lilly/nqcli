import type { DefaultConnection } from "@/domain";

import type { OperationOptions } from "./operation";

export interface ConnectionPort {
	getDefault(options?: OperationOptions): Promise<DefaultConnection>;
}
