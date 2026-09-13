import { SearchIcon } from "lucide-react";
import { useSearchParams } from "react-router";

import { Input, SelectField, type SelectOption, Spinner } from "@/components";
import { useSearchableAttributes } from "@/core";
import type { VertexType } from "@/domain";
import { useDebounceValue } from "@/hooks";
import { SEARCH_TOKENS } from "@/utils";

const SEARCH_DELAY_MS = 400;
const SEARCH_PARAM = "q";
const ATTRIBUTE_PARAM = "field";

export type DataExplorerSearchController = {
	searchTerm: string;
	debouncedSearchTerm: string;
	selectedAttribute: string;
	searchByAttributes: string[];
	isActive: boolean;
	attributeOptions: SelectOption[];
	onSearchTermChange(value: string): void;
	onAttributeChange(value: string | string[]): void;
};

export function useDataExplorerSearch(
	vertexType: VertexType,
): DataExplorerSearchController {
	const [searchParams, setSearchParams] = useSearchParams();
	const searchableAttributes = useSearchableAttributes(vertexType);
	const searchTerm = searchParams.get(SEARCH_PARAM) ?? "";
	const debouncedSearchTerm = useDebounceValue(
		searchTerm.trim(),
		SEARCH_DELAY_MS,
	);
	const attributeOptions: SelectOption[] = [
		{
			label: "All searchable fields",
			value: SEARCH_TOKENS.ALL_ATTRIBUTES,
		},
		{ label: "ID", value: SEARCH_TOKENS.NODE_ID },
		...searchableAttributes.map((attribute) => ({
			label: attribute.displayLabel,
			value: attribute.name,
		})),
	];
	const requestedAttribute =
		searchParams.get(ATTRIBUTE_PARAM) ?? SEARCH_TOKENS.ALL_ATTRIBUTES;
	const selectedAttribute = attributeOptions.some(
		(option) => option.value === requestedAttribute,
	)
		? requestedAttribute
		: SEARCH_TOKENS.ALL_ATTRIBUTES;
	const searchByAttributes =
		selectedAttribute === SEARCH_TOKENS.ALL_ATTRIBUTES
			? [
					SEARCH_TOKENS.NODE_ID,
					...searchableAttributes.map((attribute) => attribute.name),
				]
			: [selectedAttribute];

	const updateSearch = (update: (next: URLSearchParams) => void) => {
		setSearchParams(
			(previous) => {
				const next = new URLSearchParams(previous);
				update(next);
				next.delete("page");
				return next;
			},
			{ replace: true },
		);
	};

	return {
		searchTerm,
		debouncedSearchTerm,
		selectedAttribute,
		searchByAttributes,
		isActive: debouncedSearchTerm.length > 0,
		attributeOptions,
		onSearchTermChange(value) {
			updateSearch((next) => {
				if (value.length === 0) next.delete(SEARCH_PARAM);
				else next.set(SEARCH_PARAM, value);
			});
		},
		onAttributeChange(value) {
			const attribute = Array.isArray(value) ? value[0] : value;
			updateSearch((next) => {
				if (attribute === SEARCH_TOKENS.ALL_ATTRIBUTES) {
					next.delete(ATTRIBUTE_PARAM);
				} else {
					next.set(ATTRIBUTE_PARAM, attribute);
				}
			});
		},
	};
}

export function DataExplorerSearch({
	controller,
	isSearching,
}: {
	controller: DataExplorerSearchController;
	isSearching: boolean;
}) {
	const selectedLabel =
		controller.attributeOptions.find(
			(option) => option.value === controller.selectedAttribute,
		)?.label ?? "fields";

	return (
		<div className="flex min-w-[28rem] items-center gap-2">
			<div className="relative min-w-56 flex-1">
				<SearchIcon className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2" />
				<Input
					type="search"
					value={controller.searchTerm}
					onChange={(event) =>
						controller.onSearchTermChange(event.target.value)
					}
					placeholder={`Search ${selectedLabel.toLocaleLowerCase()}`}
					aria-label="Search nodes"
					className="h-11 px-9"
				/>
				{isSearching && (
					<span
						role="status"
						aria-label="Searching"
						className="text-muted-foreground absolute top-1/2 right-3 -translate-y-1/2"
					>
						<Spinner className="size-4" />
					</span>
				)}
			</div>
			<SelectField
				className="w-48"
				value={controller.selectedAttribute}
				onValueChange={controller.onAttributeChange}
				options={controller.attributeOptions}
				label="Search field"
				labelPlacement="inner"
				aria-label="Search field"
			/>
		</div>
	);
}
