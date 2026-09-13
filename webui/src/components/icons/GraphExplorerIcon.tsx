import type { IconBaseProps } from "./IconBase";

import IconBase from "./IconBase";

export const GraphExplorerIcon = (props: IconBaseProps) => {
	return (
		<IconBase {...props}>
			<path
				d="M5 15 9 6l6 3 4-5m-4 5 3 9-7 1-6-4M9 6l2 13"
				stroke="currentColor"
				strokeWidth="1.6"
				strokeLinecap="round"
				strokeLinejoin="round"
			/>
			<circle cx="5" cy="15" r="2" fill="currentColor" />
			<circle cx="9" cy="6" r="1.7" fill="currentColor" />
			<circle cx="15" cy="9" r="2.3" fill="currentColor" />
			<circle cx="19" cy="4" r="1.5" fill="currentColor" />
			<circle cx="18" cy="18" r="2" fill="currentColor" />
			<circle cx="11" cy="19" r="1.6" fill="currentColor" />
		</IconBase>
	);
};

export default GraphExplorerIcon;
