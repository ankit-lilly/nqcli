import { SearchIcon } from "lucide-react";
import {
	type ComponentPropsWithRef,
	type ReactNode,
	useMemo,
	useState,
} from "react";

import {
	Chip,
	EdgeSymbol,
	Input,
	Panel,
	PanelContent,
	PanelHeader,
	PanelHeaderActions,
	PanelHeaderCloseButton,
	PanelTitle,
	VertexSymbolByType,
} from "@/components";
import {
	createEdgeConnectionId,
	useActiveSchema,
	useDisplayEdgeTypeConfigs,
	useDisplayVertexTypeConfigs,
} from "@/core";
import { cn } from "@/utils";

import type {
	SchemaGraphSelection,
	SchemaGraphSelectionItem,
} from "../SchemaGraph";

import { useSchemaViewSidebar } from "./schemaViewLayout";

export type SchemaInventoryProps = {
	selection: SchemaGraphSelection;
	onSelectionChange?: (item: SchemaGraphSelectionItem) => void;
};

/** Complete, searchable inventory of the items rendered in the schema graph. */
export function SchemaInventory({
	selection,
	onSelectionChange,
}: SchemaInventoryProps) {
	const { closeSidebar } = useSchemaViewSidebar();
	const schema = useActiveSchema();
	const vertexConfigs = useDisplayVertexTypeConfigs();
	const edgeConfigs = useDisplayEdgeTypeConfigs();
	const [search, setSearch] = useState("");
	const normalizedSearch = search.trim().toLocaleLowerCase();

	const vertices = useMemo(
		() =>
			[...vertexConfigs.values()].filter((vertex) =>
				matchesSearch(normalizedSearch, vertex.type, vertex.displayLabel),
			),
		[normalizedSearch, vertexConfigs],
	);
	const edgeConnections = useMemo(
		() =>
			(schema.edgeConnections ?? [])
				.filter((connection) => {
					const edgeLabel =
						edgeConfigs.get(connection.edgeType)?.displayLabel ??
						connection.edgeType;
					return matchesSearch(
						normalizedSearch,
						edgeLabel,
						connection.edgeType,
						connection.sourceVertexType,
						connection.targetVertexType,
					);
				})
				.toSorted((left, right) =>
					connectionSortKey(left).localeCompare(connectionSortKey(right)),
				),
		[edgeConfigs, normalizedSearch, schema.edgeConnections],
	);

	return (
		<Panel className="size-full" variant="sidebar">
			<PanelHeader>
				<PanelTitle>Schema</PanelTitle>
				<PanelHeaderActions>
					<PanelHeaderCloseButton onClose={closeSidebar} />
				</PanelHeaderActions>
			</PanelHeader>
			<div className="relative shrink-0 border-b p-3">
				<SearchIcon className="text-muted-foreground pointer-events-none absolute top-1/2 left-6 size-4 -translate-y-1/2" />
				<Input
					type="search"
					value={search}
					onChange={(event) => setSearch(event.target.value)}
					placeholder="Search schema"
					aria-label="Search schema"
					className="pl-9"
				/>
			</div>
			<PanelContent className="gap-6 p-3">
				<InventorySection title="Node labels" count={vertices.length}>
					{vertices.map((vertex) => {
						const item = { type: "vertex-type", id: vertex.type } as const;
						return (
							<InventoryButton
								key={vertex.type}
								selected={isSelected(selection, item)}
								onClick={() => onSelectionChange?.(item)}
							>
								<VertexSymbolByType
									vertexType={vertex.type}
									className="size-7"
								/>
								<span className="min-w-0 truncate">{vertex.displayLabel}</span>
							</InventoryButton>
						);
					})}
				</InventorySection>

				<InventorySection title="Relationships" count={edgeConnections.length}>
					{edgeConnections.map((connection) => {
						const id = createEdgeConnectionId(connection);
						const item = { type: "edge-connection", id } as const;
						const displayLabel =
							edgeConfigs.get(connection.edgeType)?.displayLabel ??
							connection.edgeType;
						return (
							<InventoryButton
								key={id}
								selected={isSelected(selection, item)}
								onClick={() => onSelectionChange?.(item)}
							>
								<EdgeSymbol className="size-7 shrink-0" />
								<span className="min-w-0 text-left">
									<span className="block truncate">{displayLabel}</span>
									<span className="text-muted-foreground block truncate text-xs font-normal">
										{connection.sourceVertexType} to{" "}
										{connection.targetVertexType}
									</span>
								</span>
							</InventoryButton>
						);
					})}
				</InventorySection>
			</PanelContent>
		</Panel>
	);
}

function InventorySection({
	title,
	count,
	children,
}: {
	title: string;
	count: number;
	children: ReactNode;
}) {
	return (
		<section className="space-y-2">
			<h2 className="flex items-center text-sm font-semibold">
				{title}
				<Chip variant="neutral-subtle" className="ml-auto">
					{count}
				</Chip>
			</h2>
			<div className="space-y-1">{children}</div>
			{count === 0 && (
				<p className="text-muted-foreground py-3 text-center text-sm">
					No matches
				</p>
			)}
		</section>
	);
}

function InventoryButton({
	selected,
	className,
	...props
}: ComponentPropsWithRef<"button"> & { selected: boolean }) {
	return (
		<button
			type="button"
			aria-pressed={selected}
			className={cn(
				"hover:bg-muted focus-visible:ring-primary flex min-h-11 w-full items-center gap-3 rounded-md px-2 py-1.5 text-left text-sm font-medium focus-visible:ring-2 focus-visible:outline-hidden",
				selected && "bg-primary-subtle text-primary",
				className,
			)}
			{...props}
		/>
	);
}

function matchesSearch(search: string, ...values: string[]) {
	return (
		search === "" ||
		values.some((value) => value.toLocaleLowerCase().includes(search))
	);
}

function connectionSortKey(connection: {
	edgeType: string;
	sourceVertexType: string;
	targetVertexType: string;
}) {
	return `${connection.edgeType}\0${connection.sourceVertexType}\0${connection.targetVertexType}`;
}

function isSelected(
	selection: SchemaGraphSelection,
	item: SchemaGraphSelectionItem,
) {
	if (!selection) return false;
	if (selection.type === "multiple") {
		return selection.items.some(
			(selected) => selected.type === item.type && selected.id === item.id,
		);
	}
	return selection.type === item.type && selection.id === item.id;
}
