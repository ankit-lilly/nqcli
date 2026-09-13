import type { DisplayAttribute } from "./displayAttribute";
import type { DisplayConfigAttribute } from "./displayTypeConfigs";

export function sortAttributeByName(
	a: DisplayConfigAttribute | DisplayAttribute,
	b: DisplayConfigAttribute | DisplayAttribute,
) {
	return a.displayLabel.localeCompare(b.displayLabel);
}
