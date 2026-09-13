import type { EntityRawId } from "@/domain";
import { availableLayoutsConfig, type LayoutName } from "./layout-config";
import type { CytoscapeType } from "./model";

type ExpandedCytoscapeLayoutOptions = {
	fixedNodeConstraint?: {
		nodeId: EntityRawId;
		position: { x: number; y: number };
	}[];
};
export const runLayout = (
	cyReference: CytoscapeType,
	layoutName: LayoutName,
	additionalLayoutsConfig: {
		[layoutName: string]: Partial<
			cytoscape.LayoutOptions & ExpandedCytoscapeLayoutOptions
		>;
	} = {},
	useAnimation = false,
) => {
	const _layout = {
		...availableLayoutsConfig[layoutName],
		...additionalLayoutsConfig[layoutName],
	};
	if (_layout) {
		_layout.animate = useAnimation;
		if (layoutName === "F_COSE") {
			// using the fixedNodeConstraint of the newer version of cytoscape-fcose to achieve a better relayout when
			// there are locked nodes
			const nodesToRunLayout = cyReference.nodes("[!__isGroupNode]:locked");
			_layout.fixedNodeConstraint = nodesToRunLayout.map((node) => ({
				nodeId: node.data().id,
				position: node.position(),
			}));
		}
		const layout = cyReference.layout(_layout as cytoscape.LayoutOptions);
		layout.run();
		return;
	}

	throw new Error(
		"Layout configuration not found, if you are using a custom layout make sure to pass down the" +
			" layout configuration through the additionalLayouts Prop ",
	);
};

export const graphLayouts = Object.keys(availableLayoutsConfig);
