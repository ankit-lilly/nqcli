declare module "cytoscape-dagre" {
	import type { Ext } from "cytoscape";
	const extension: Ext;
	export default extension;
}
declare module "/wails/runtime.js" {
	export const Call: {
		ByName<T = unknown>(name: string, ...args: unknown[]): Promise<T>;
	};
	export type CancellablePromise<T> = Promise<T> & { cancel(): Promise<void> };
}
