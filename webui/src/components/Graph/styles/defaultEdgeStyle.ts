import type { RenderedEdgeStyle } from "../Graph.model";

const defaultEdgeStyle: RenderedEdgeStyle = {
	visible: true,
	opacity: 1,
	color: "#f8fafc",
	width: 2,
	lineColor: "#94a3b8",
	curveStyle: "bezier",
	sourceArrowShape: "none",
	sourceArrowColor: "#94a3b8",
	targetArrowShape: "triangle",
	targetArrowColor: "#94a3b8",
	lineCap: "square",
	lineStyle: "solid",
	text: {
		background: "#334155",
		border: {
			width: 0,
			opacity: 1,
			color: "#1f2937",
			style: "solid",
		},
		color: "#ffffff",
		fontSize: 7,
		hAlign: "center",
		maxWidth: 80,
		minZoomedFontSize: 6,
		opacity: 0.76,
		padding: 2,
		rotation: "autorotate",
		shape: "round-rectangle",
		vAlign: "bottom",
		vMargin: 0,
		wrap: "wrap",
	},
};

export default defaultEdgeStyle;
