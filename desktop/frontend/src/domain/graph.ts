export type GraphElementData = Record<string, unknown> & { id: string };

export type GraphElement = {
	group: "nodes" | "edges";
	data: GraphElementData;
};

export type GraphSelection = GraphElement | null;

export function toGraphElements(values: unknown[] | null): GraphElement[] {
	if (!values) return [];
	return values.flatMap((value) => {
		if (!value || typeof value !== "object") return [];
		const candidate = value as {
			group?: unknown;
			data?: Record<string, unknown> | null;
		};
		if (
			(candidate.group !== "nodes" && candidate.group !== "edges") ||
			!candidate.data ||
			typeof candidate.data.id !== "string"
		) {
			return [];
		}
		return [
			{ group: candidate.group, data: candidate.data as GraphElementData },
		];
	});
}

/** Adds or refreshes elements by ID while retaining stable ordering. */
export function mergeGraphElements(
	current: GraphElement[],
	incoming: GraphElement[],
): GraphElement[] {
	if (incoming.length === 0) return current;
	const merged = new Map(current.map((element) => [element.data.id, element]));
	for (const element of incoming) {
		const previous = merged.get(element.data.id);
		merged.set(
			element.data.id,
			previous
				? { ...element, data: { ...previous.data, ...element.data } }
				: element,
		);
	}
	return [...merged.values()];
}

export function removeGraphElement(
	elements: GraphElement[],
	selected: GraphElement,
): GraphElement[] {
	if (selected.group === "edges") {
		return elements.filter((element) => element.data.id !== selected.data.id);
	}
	return elements.filter(
		(element) =>
			element.data.id !== selected.data.id &&
			element.data.source !== selected.data.id &&
			element.data.target !== selected.data.id,
	);
}

export function visibleVertexIDs(elements: GraphElement[]): string[] {
	return elements
		.filter((element) => element.group === "nodes")
		.map((element) => element.data.id);
}

/** Stable identity for the rendered topology, independent of result ordering. */
export function graphTopologyKey(elements: GraphElement[]): string {
	return elements
		.map((element) =>
			[
				element.group,
				element.data.id,
				element.data.source ?? "",
				element.data.target ?? "",
			].join("\u0000"),
		)
		.sort()
		.join("\u001e");
}
