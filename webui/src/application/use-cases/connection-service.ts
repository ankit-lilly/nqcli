import type { ConnectionPort, OperationOptions } from "../ports";

export class ConnectionService {
	constructor(private readonly connections: ConnectionPort) {}

	getDefault(options?: OperationOptions) {
		return this.connections.getDefault(options);
	}
}
