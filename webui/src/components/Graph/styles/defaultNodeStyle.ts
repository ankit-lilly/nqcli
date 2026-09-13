import type { RenderedNodeStyle } from "../Graph.model";

const defaultNodeStyle: RenderedNodeStyle = {
	background: "#0f766e",
	backgroundOpacity: 0.55,
	borderColor: "#14b8a6",
	backgroundFit: "none",
	backgroundWidth: "60%",
	backgroundHeight: "60%",
	borderWidth: 1,
	borderStyle: "solid",
	borderOpacity: 0,
	color: "#FFFFFF",
	height: 24,
	opacity: 1,
	padding: 0,
	shape: "ellipse",
	text: {
		fontSize: 7,
		minZoomedFontSize: 6,
		rotation: "autorotate",
		vAlign: "bottom",
		vMargin: 0,
		wrap: "wrap",
		hAlign: "center",
		maxWidth: 80,
		color: "#ffffff",
		background: "#1f2937",
		opacity: 0.76,
		padding: 2,
		shape: "round-rectangle",
		border: {
			width: 0,
			opacity: 0.5,
			color: "#1f2937",
			style: "solid",
		},
	},
	visible: true,
	width: 24,
};

export default defaultNodeStyle;
