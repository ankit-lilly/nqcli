import { useEffect, useRef, useState } from "react";

import { CytoscapeSurfaceAdapter } from "@/adapters/cytoscape";
import { useDeepMemo } from "@/hooks";

import type { GraphProps } from "../Graph";
import type { Config, CytoscapeType } from "../Graph.model";

export interface UseInitCytoscapeProps
	extends Required<
		Pick<GraphProps, keyof Omit<Config, "minZoom" | "maxZoom" | "pan">>
	> {
	wrapper?: HTMLElement;
	minZoom?: number;
	maxZoom?: number;
	pan?: { x: number; y: number };
	onLayoutRunningChanged?: (isRunning: boolean) => void;
	onZoomChanged?: (e: unknown) => void;
	onPanChanged?: (e: unknown) => void;
}

const useInitCytoscape = ({
	wrapper,
	onLayoutRunningChanged,
	onPanChanged,
	onZoomChanged,
	...config
}: UseInitCytoscapeProps) => {
	const [cy, setCy] = useState<CytoscapeType | undefined>();

	const memoizedConfig = useDeepMemo(() => config, [config]);

	// holds the event handlers so we do not need to keep re-attaching them
	const eventHandlerRefs = useRef({
		onLayoutRunningChanged,
		onPanChanged,
		onZoomChanged,
	});

	useEffect(() => {
		eventHandlerRefs.current = {
			onLayoutRunningChanged,
			onPanChanged,
			onZoomChanged,
		};
	}, [onLayoutRunningChanged, onPanChanged, onZoomChanged]);

	useEffect(() => {
		if (wrapper) {
			const surface = new CytoscapeSurfaceAdapter(memoizedConfig, {
				onLayoutRunningChanged: (running) =>
					eventHandlerRefs.current.onLayoutRunningChanged?.(running),
				onZoomChanged: (zoom) => eventHandlerRefs.current.onZoomChanged?.(zoom),
				onPanChanged: (pan) => eventHandlerRefs.current.onPanChanged?.(pan),
			});
			surface.mount(wrapper);
			setCy(surface.cytoscape);

			return () => {
				surface.destroy();
				setCy(undefined);
			};
		}
		// since this is to init cytoscape, this should only run when wrapper is set
	}, [memoizedConfig, wrapper]);

	return cy;
};

export default useInitCytoscape;
