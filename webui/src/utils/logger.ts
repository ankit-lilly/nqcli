/* oxlint-disable no-console */

import { env } from "./env";

/** Returns true when debug and log calls should be written to the console. */
function isLoggingEnabled() {
	return env.DEV;
}

/** A thin wrapper around console that suppresses debug/log in production. */
export default {
	/** Calls `console.debug` in development. */
	debug(...args: unknown[]) {
		if (isLoggingEnabled()) {
			console.debug(...args);
		}
	},
	/** Calls `console.log` in development. */
	log(...args: unknown[]) {
		if (isLoggingEnabled()) {
			console.log(...args);
		}
	},
	/** Calls `console.warn`. */
	warn(...args: unknown[]) {
		console.warn(...args);
	},
	/** Calls `console.error`. */
	error(...args: unknown[]) {
		console.error(...args);
	},
};
