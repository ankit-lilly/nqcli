import { useQuery } from "@tanstack/react-query";
import { useAtom, useSetAtom } from "jotai";
import {
	type PropsWithChildren,
	startTransition,
	Suspense,
	useEffect,
} from "react";

import { PanelEmptyState, PanelError, Spinner } from "@/components";
import { logger } from "@/utils";

import { fetchDefaultConnection } from "./defaultConnection";
import { activeConfigurationAtom, configurationAtom } from "./StateProvider";

function AppStatusLoader({ children }: PropsWithChildren) {
	return (
		<Suspense fallback={<PreparingEnvironment />}>
			<LoadDefaultConfig>{children}</LoadDefaultConfig>
		</Suspense>
	);
}

function LoadDefaultConfig({ children }: PropsWithChildren) {
	const [activeConfig, setActiveConfig] = useAtom(activeConfigurationAtom);
	const setConfiguration = useSetAtom(configurationAtom);

	const defaultConfigQuery = useQuery({
		queryKey: ["default-connection"],
		queryFn: fetchDefaultConnection,
		staleTime: Infinity,
	});

	const defaultConnection = defaultConfigQuery.data;

	useEffect(() => {
		if (!defaultConnection) {
			// Query hasn't run yet
			return;
		}

		startTransition(() => {
			logger.debug("Loading default connection", defaultConnection);
			setConfiguration((prev) => {
				const updatedConfig = new Map(prev);
				updatedConfig.set(defaultConnection.id, defaultConnection);
				return updatedConfig;
			});
			setActiveConfig(defaultConnection.id);
		});
	}, [setActiveConfig, setConfiguration, defaultConnection]);

	if (defaultConfigQuery.isLoading) {
		return (
			<PanelEmptyState
				title="Loading default connection..."
				subtitle="We are checking for a default connection"
				icon={<Spinner />}
			/>
		);
	}

	if (defaultConfigQuery.isError) {
		return (
			<PanelError
				error={defaultConfigQuery.error}
				onRetry={() => void defaultConfigQuery.refetch()}
			/>
		);
	}

	if (defaultConnection && activeConfig !== defaultConnection.id) {
		return (
			<PanelEmptyState
				title="Reading configuration..."
				subtitle="We are loading the configuration from the file"
				icon={<Spinner />}
			/>
		);
	}

	return <>{children}</>;
}

function PreparingEnvironment() {
	return (
		<PanelEmptyState
			title="Preparing environment..."
			subtitle="We are loading all components"
			icon={<Spinner />}
		/>
	);
}

export default AppStatusLoader;
