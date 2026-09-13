/** ASCII characters used in strings */
export const ASCII = {
	/** Non breaking space */
	NBSP: "\u00A0",
	/** Small arrow character */
	RSAQUO: "\u203A",
	/** Left double angle quotes */
	LAQUO: "\u00ab",
	/** Right double angle quotes */
	RAQUO: "\u00bb",
	/** Right arrow character */
	RARR: "\u2192",
} as const;

export const DEFAULT_BATCH_REQUEST_SIZE = 100;

/** The name of the special property representing the node ID */
export const RESERVED_ID_PROPERTY = "~id";

/** The name of the property representing the list of types of the node */
export const RESERVED_TYPES_PROPERTY = "types";

/** The root URL for the app used for reloading fresh. */
export const RELOAD_URL =
	import.meta.env.BASE_URL.substring(-1) !== "/"
		? import.meta.env.BASE_URL + "/"
		: import.meta.env.BASE_URL;

/** Labels used in the UI */
export const LABELS = {
	/** The application name used in UI and saved-file metadata. */
	APP_NAME: "Graph Explorer",
	/** Shown when a type is missing  */
	MISSING_TYPE: `${ASCII.LAQUO}No Type${ASCII.RAQUO}`,
	/** Shown when a value is missing */
	MISSING_VALUE: `${ASCII.LAQUO}No Value${ASCII.RAQUO}`,
	/** Shown when a value is empty (like empty string) */
	EMPTY_VALUE: `${ASCII.LAQUO}Empty Value${ASCII.RAQUO}`,

	/** Labels related to the sidebar UI */
	SIDEBAR: {
		/** The title for the complete schema inventory panel */
		SCHEMA: "Schema",
		/** The title for the selection details panel */
		SELECTION_DETAILS: "Selection Details",
		/** The title for the combined node and edge styles panel */
		STYLES: "Styles",
	},
} as const;

/** Searchable tokens */
export const SEARCH_TOKENS = {
	/** Token to search over all vertex types */
	ALL_VERTEX_TYPES: "__all",
	/** Token to search over all attributes */
	ALL_ATTRIBUTES: "__all",
	/** Token to search by node ID */
	NODE_ID: "__id",
} as const;
