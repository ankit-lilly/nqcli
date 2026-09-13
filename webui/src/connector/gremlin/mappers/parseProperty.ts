import type { EntityPropertyValue } from "@/core";

import type { GProperty, GVertexProperty } from "../types";

export default function parseProperty(
	property: GVertexProperty | GProperty,
): EntityPropertyValue {
	const value = property["@value"].value;

	if (value == null) {
		return null;
	}

	if (typeof value === "string") {
		return value;
	}

	if (typeof value === "boolean") {
		return value;
	}

	if (value["@type"] === "g:Date") {
		return new Date(value["@value"]);
	}

	return value["@value"];
}
