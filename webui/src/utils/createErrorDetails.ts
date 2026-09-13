import { ZodError } from "zod";

import { QueryValueError } from "@/connector/queryValueError";

import { NetworkError } from "./NetworkError";
import { ServerConnectionError } from "./ServerConnectionError";

export type ErrorDetails = {
	name: string;
	/** Undefined when the error has no meaningful message to display. */
	message: string | undefined;
	data?: string;
};

/** Extracts a name and message from an unknown error for display in error detail dialogs. */
export function createErrorDetails(error: unknown): ErrorDetails {
	if (error instanceof ServerConnectionError) {
		const data: Record<string, unknown> = { url: error.url };
		if (error.cause) {
			data.cause = serializeCause(error.cause);
		}
		return {
			name: error.name,
			message: error.message,
			data: JSON.stringify(data, null, 2),
		};
	}
	if (error instanceof NetworkError) {
		const name = error.statusText
			? `${error.statusCode} ${error.statusText}`
			: `${error.statusCode}`;
		const data =
			error.data != null ? JSON.stringify(error.data, null, 2) : undefined;
		return data
			? { name, message: error.message, data }
			: { name, message: error.message };
	}
	if (error instanceof ZodError) {
		return {
			name: "ZodError",
			message: error.message,
			data: JSON.stringify(error.issues, null, 2),
		};
	}
	if (error instanceof QueryValueError) {
		return {
			name: error.name,
			message: error.message,
			data: JSON.stringify(error.details, null, 2),
		};
	}
	if (isErrorLike(error)) {
		const data = error.cause
			? JSON.stringify(serializeCause(error.cause), null, 2)
			: undefined;
		return data
			? { name: error.name, message: error.message, data }
			: { name: error.name, message: error.message };
	}
	return { name: "Unknown Error", message: JSON.stringify(error, null, 2) };
}

function serializeCause(cause: unknown): unknown {
	if (isErrorLike(cause)) {
		const result: Record<string, unknown> = {
			name: cause.name,
			message: cause.message,
		};
		if ("code" in cause) {
			result.code = cause.code;
		}
		if (cause.cause) {
			result.cause = serializeCause(cause.cause);
		}
		return result;
	}
	return cause;
}

type ErrorLike = {
	name: string;
	message: string;
	cause?: unknown;
	code?: unknown;
};

function isErrorLike(value: unknown): value is ErrorLike {
	return (
		value instanceof Error ||
		(typeof DOMException !== "undefined" && value instanceof DOMException)
	);
}
