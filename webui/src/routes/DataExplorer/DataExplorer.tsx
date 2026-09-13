import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";

import {
	Button,
	CheckIcon,
	EmptyState,
	EmptyStateContent,
	EmptyStateDescription,
	EmptyStateTitle,
	NavBar,
	NavBarContent,
	NavBarTitle,
	NoNodeTypesEmptyState,
	Panel,
	PanelContent,
	PanelEmptyState,
	PanelError,
	PanelGroup,
	PanelHeader,
	PersistenceStatusIndicator,
	RouteButtonGroup,
	SchemaDiscoveryBoundary,
	SelectField,
	SendIcon,
	Spinner,
	Workspace,
	WorkspaceContent,
} from "@/components";
import {
	type ColumnDefinition,
	PaginationControl,
	TabularEmptyBodyControls,
	TabularFooterControls,
	type TabularInstance,
} from "@/components/Tabular";
import { ExternalExportControl } from "@/components/Tabular/controls/ExportControl/ExternalExportControl";
import Tabular from "@/components/Tabular/Tabular";
import { nodeCountByNodeTypeQuery, searchQuery } from "@/connector";
import {
	createVertexType,
	type DisplayVertex,
	useConfiguration,
	useDisplayVertexTypeConfig,
	useDisplayVertexTypeConfigs,
	useDisplayVerticesFromVertices,
	useUpdateSchemaFromEntities,
	type Vertex,
	type VertexType,
} from "@/core";
import { useVertexTypeConfig } from "@/core/ConfigurationProvider/useConfiguration";
import { useVertexStyling } from "@/core/StateProvider/graphStyles";
import type { SearchVerticesCommand } from "@/domain";
import { useAddVertexToGraph, useHasVertexBeenAddedToGraph } from "@/hooks";
import useTranslations from "@/hooks/useTranslations";
import {
	LABELS,
	RESERVED_ID_PROPERTY,
	RESERVED_TYPES_PROPERTY,
	SEARCH_TOKENS,
} from "@/utils/constants";

import {
	DataExplorerSearch,
	useDataExplorerSearch,
} from "./DataExplorerSearch";

const DEFAULT_COLUMN = {
	width: 150,
};

export default function DataExplorer() {
	const { vertexType } = useParams();
	const navigate = useNavigate();
	const vtConfigs = useDisplayVertexTypeConfigs().values().toArray();

	if (!vertexType && vtConfigs.length > 0) {
		navigate(`/data-explorer/${encodeURIComponent(vtConfigs[0].type)}`, {
			replace: true,
		});
	}

	if (vtConfigs.length === 0 || !vertexType) {
		return (
			<Layout>
				<PanelGroup className="grid">
					<Panel>
						<PanelContent>
							<NoNodeTypesEmptyState />
						</PanelContent>
					</Panel>
				</PanelGroup>
			</Layout>
		);
	}

	return (
		<Layout>
			<PanelGroup className="grid">
				<DataExplorerContent vertexType={createVertexType(vertexType)} />
			</PanelGroup>
		</Layout>
	);
}

function DataExplorerContent({ vertexType }: { vertexType: VertexType }) {
	const t = useTranslations();
	const navigate = useNavigate();

	// Automatically updates counts if needed
	useQuery(nodeCountByNodeTypeQuery(vertexType));

	const vertexConfig = useVertexTypeConfig(vertexType);
	const displayTypeConfig = useDisplayVertexTypeConfig(vertexType);
	const { pageIndex, pageSize, onPageIndexChange, onPageSizeChange } =
		usePagingOptions();
	const search = useDataExplorerSearch(vertexType);

	const [tableInstance, setTableInstance] =
		useState<TabularInstance<DisplayVertex> | null>(null);
	const columns = useColumnDefinitions(vertexType);

	const query = useDataExplorerQuery(
		vertexType,
		pageSize,
		pageIndex,
		search.debouncedSearchTerm,
		search.searchByAttributes,
	);
	const displayVertices = useDisplayVerticesFromVertices(
		query.data?.vertices.slice(0, pageSize) ?? [],
	)
		.values()
		.toArray();

	const vtConfigs = useDisplayVertexTypeConfigs().values().toArray();
	const vertexTypeOptions = vtConfigs.map((config) => ({
		value: config.type,
		label: config.displayLabel,
	}));

	const onVertexTypeChange = (value: string | string[]) => {
		const newType = Array.isArray(value) ? value[0] : value;
		navigate(`/data-explorer/${encodeURIComponent(newType)}`, {
			replace: true,
		});
	};

	return (
		<Panel>
			<PanelHeader className="justify-between py-3">
				<SelectField
					className="w-64"
					value={vertexType}
					onValueChange={onVertexTypeChange}
					options={vertexTypeOptions}
					label={t("node-type")}
					labelPlacement="inner"
				/>
				<DataExplorerSearch
					controller={search}
					isSearching={search.searchTerm.trim().length > 0 && query.isFetching}
				/>
				<div className="flex items-center gap-2">
					<DisplayNameAndDescriptionOptions vertexType={vertexType} />
					{tableInstance ? (
						<ExternalExportControl
							instance={tableInstance}
							hideOptions
							forceOnlyPage
						/>
					) : null}
				</div>
			</PanelHeader>
			<Tabular
				ref={(instance) => {
					setTableInstance(instance);
				}}
				defaultColumn={DEFAULT_COLUMN}
				data={displayVertices}
				columns={columns}
				fullWidth={true}
				pageIndex={pageIndex}
				pageSize={pageSize}
				disablePagination={true}
				disableFilters={true}
				disableSorting={true}
			>
				<TabularEmptyBodyControls>
					{query.isPending ? (
						<PanelEmptyState title="Loading data..." icon={<Spinner />} />
					) : null}
					{query.isError ? (
						<PanelError error={query.error} onRetry={query.refetch} />
					) : null}
					{query.data?.vertices.length === 0 && (
						<EmptyState>
							<EmptyStateContent>
								<EmptyStateTitle>No Results</EmptyStateTitle>
								<EmptyStateDescription>
									{`No nodes found for "${displayTypeConfig.displayLabel}"`}
								</EmptyStateDescription>
							</EmptyStateContent>
						</EmptyState>
					)}
				</TabularEmptyBodyControls>
				<TabularFooterControls>
					<PaginationControl
						pageIndex={pageIndex}
						onPageIndexChange={onPageIndexChange}
						pageSize={pageSize}
						onPageSizeChange={onPageSizeChange}
						totalRows={
							search.isActive
								? undefined
								: (vertexConfig?.total ?? pageSize * (pageIndex + 2))
						}
						visibleRows={displayVertices.length}
						hasNextPage={(query.data?.vertices.length ?? 0) > pageSize}
					/>
				</TabularFooterControls>
			</Tabular>
		</Panel>
	);
}

function Layout({ children }: { children: React.ReactNode }) {
	const config = useConfiguration();

	return (
		<Workspace>
			<NavBar logoVisible>
				<NavBarContent>
					<NavBarTitle
						title="Data Explorer"
						subtitle={`Connection: ${config?.displayLabel || config?.id}`}
					/>
					<PersistenceStatusIndicator />
				</NavBarContent>
				<RouteButtonGroup active="data-explorer" />
			</NavBar>
			<WorkspaceContent>
				<SchemaDiscoveryBoundary>{children}</SchemaDiscoveryBoundary>
			</WorkspaceContent>
		</Workspace>
	);
}

function DisplayNameAndDescriptionOptions({
	vertexType,
}: {
	vertexType: VertexType;
}) {
	const t = useTranslations();
	const vertexConfig = useVertexTypeConfig(vertexType);
	const displayConfig = useDisplayVertexTypeConfig(vertexType);
	const selectOptions = (() => {
		const options = displayConfig.attributes.map((attr) => ({
			value: attr.name,
			label: attr.displayLabel,
		}));

		options.unshift({
			label: t("node-type"),
			value: RESERVED_TYPES_PROPERTY,
		});
		options.unshift({
			label: t("node-id"),
			value: RESERVED_ID_PROPERTY,
		});

		return options;
	})();

	const { setVertexStyle } = useVertexStyling(vertexType);
	const onDisplayNameChange =
		(field: "name" | "longName") => (value: string | string[]) => {
			if (field === "name") {
				setVertexStyle({ displayNameAttribute: value as string });
			}

			if (field === "longName") {
				setVertexStyle({ longDisplayNameAttribute: value as string });
			}
		};

	return (
		<div className="flex flex-wrap gap-2">
			<SelectField
				className="w-[200px]"
				value={vertexConfig?.displayNameAttribute || ""}
				onValueChange={onDisplayNameChange("name")}
				options={selectOptions}
				label="Display Name"
				labelPlacement="inner"
			/>
			<SelectField
				className="w-[200px]"
				value={vertexConfig?.longDisplayNameAttribute || ""}
				onValueChange={onDisplayNameChange("longName")}
				options={selectOptions}
				label="Display Description"
				labelPlacement="inner"
			/>
		</div>
	);
}

function AddToExplorerButton({ vertex }: { vertex: Vertex }) {
	const addToGraph = useAddVertexToGraph(vertex);
	const isInExplorer = useHasVertexBeenAddedToGraph(vertex.id);

	return (
		<Button
			disabled={isInExplorer}
			variant="outline"
			size="small"
			onClick={addToGraph}
			className="text-nowrap"
		>
			{isInExplorer ? <CheckIcon /> : <SendIcon />}
			{isInExplorer ? "Sent to Explorer" : "Send to Explorer"}
		</Button>
	);
}

function useColumnDefinitions(vertexType: VertexType) {
	const t = useTranslations();
	const displayConfig = useDisplayVertexTypeConfig(vertexType);
	const columns: ColumnDefinition<DisplayVertex>[] = (() => {
		const vtColumns: ColumnDefinition<DisplayVertex>[] =
			displayConfig.attributes.map((attr) => ({
				id: attr.name,
				label: attr.displayLabel,
				accessor: (row) =>
					row.attributes.find((a) => a.name === attr.name)?.displayValue ??
					LABELS.MISSING_VALUE,
			}));
		vtColumns.unshift({
			label: t("node-id"),
			id: SEARCH_TOKENS.NODE_ID,
			accessor: (row) => row.displayId,
			filterable: false,
		});

		vtColumns.push({
			id: "__send_to_explorer",
			label: "",
			filterable: false,
			sortable: false,
			resizable: false,
			width: 180,
			cellComponent: ({ cell }) => (
				<AddToExplorerButton vertex={cell.row.original.original} />
			),
		});

		return vtColumns;
	})();
	return columns;
}

function usePagingOptions() {
	const [searchParams, setSearchParams] = useSearchParams();
	const pageIndex = Number(searchParams.get("page") || 1) - 1;
	const pageSize = Number(searchParams.get("pageSize") || 20);
	const onPageIndexChange = (pageIndex: number) => {
		setSearchParams(
			(prevState) => {
				const next = new URLSearchParams(prevState);
				const currPageSize = Number(prevState.get("pageSize") || 20);
				if (pageIndex === 0) next.delete("page");
				else next.set("page", String(pageIndex + 1));
				next.set("pageSize", String(currPageSize));
				return next;
			},
			{ replace: true },
		);
	};

	const onPageSizeChange = (pageSize: number) => {
		setSearchParams(
			(prevState) => {
				const next = new URLSearchParams(prevState);
				next.delete("page");
				next.set("pageSize", String(pageSize));
				return next;
			},
			{ replace: true },
		);
	};

	return {
		pageIndex,
		pageSize,
		onPageIndexChange,
		onPageSizeChange,
	};
}

function useDataExplorerQuery(
	vertexType: VertexType,
	pageSize: number,
	pageIndex: number,
	searchTerm: string,
	searchByAttributes: string[],
) {
	const updateSchema = useUpdateSchemaFromEntities();
	const isSearchActive = searchTerm.length > 0;

	const searchRequest: SearchVerticesCommand = {
		vertexTypes: [vertexType],
		limit: pageSize + (isSearchActive ? 1 : 0),
		offset: pageIndex * pageSize,
		searchTerm: isSearchActive ? searchTerm : undefined,
		searchByAttributes: isSearchActive ? searchByAttributes : undefined,
		exactMatch: false,
	};

	return useQuery({
		...searchQuery(searchRequest, updateSchema),
		placeholderData: (previousData, previousQuery) => {
			const previousRequest = previousQuery?.queryKey[1] as
				| SearchVerticesCommand
				| undefined;
			return previousRequest?.vertexTypes?.[0] === vertexType
				? previousData
				: undefined;
		},
	});
}
