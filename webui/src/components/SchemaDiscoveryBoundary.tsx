import type { PropsWithChildren } from "react";

import { DatabaseIcon, RefreshCwIcon } from "lucide-react";

import {
	EmptyState,
	EmptyStateActions,
	EmptyStateContent,
	EmptyStateDescription,
	EmptyStateIcon,
	EmptyStateTitle,
	Button,
	Panel,
	PanelContent,
	PanelEmptyState,
	PanelError,
	PanelGroup,
	PanelHeader,
	PanelTitle,
	SyncIcon,
} from "@/components";
import { useConfiguration, useHasActiveSchema } from "@/core";
import { useSchemaSync } from "@/hooks/useSchemaSync";

/**
 * Renders loading, error, or no-schema states for schema discovery.
 * Renders children when a schema has been successfully synced.
 */
export function SchemaDiscoveryBoundary({ children }: PropsWithChildren) {
	const config = useConfiguration();
	const { schemaQuery, refreshSchema, isFetching } = useSchemaSync();
	const hasSchema = useHasActiveSchema();

	// 0. If no connection is configured, show no-connection state
	if (!config) {
		return (
			<Layout>
				<EmptyState className="p-6">
					<EmptyStateIcon>
						<DatabaseIcon />
					</EmptyStateIcon>
					<EmptyStateContent>
						<EmptyStateTitle>No Connection</EmptyStateTitle>
						<EmptyStateDescription>
							No default profile connection is available.
						</EmptyStateDescription>
						<EmptyStateActions>
							<Button
								variant="primary"
								onClick={() => window.location.reload()}
							>
								Retry <RefreshCwIcon />
							</Button>
						</EmptyStateActions>
					</EmptyStateContent>
				</EmptyState>
			</Layout>
		);
	}

	// 1. If loading/fetching, show loading state
	if (isFetching && !hasSchema) {
		return (
			<Layout>
				<PanelEmptyState
					variant="info"
					icon={<SyncIcon className="animate-spin" />}
					title="Synchronizing..."
					subtitle="The connection is being synchronized."
					className="p-6"
				/>
			</Layout>
		);
	}

	// 2. If data exists, render children
	if (hasSchema) {
		return children;
	}

	// 3. If error, show error state
	if (schemaQuery.error) {
		return (
			<Layout>
				<PanelError error={schemaQuery.error} onRetry={refreshSchema} />
			</Layout>
		);
	}

	// 4. No schema available
	return (
		<Layout>
			<PanelEmptyState
				variant="info"
				icon={<SyncIcon />}
				title="No Schema Available"
				subtitle="Synchronize the connection to explore the data."
				onAction={refreshSchema}
				actionLabel="Synchronize"
				className="p-6"
			/>
		</Layout>
	);
}

function Layout({ children }: PropsWithChildren) {
	return (
		<PanelGroup>
			<Panel className="flex-1">
				<PanelHeader>
					<PanelTitle>Schema Sync</PanelTitle>
				</PanelHeader>
				<PanelContent>{children}</PanelContent>
			</Panel>
		</PanelGroup>
	);
}
