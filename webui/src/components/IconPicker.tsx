import { useState } from "react";

import {
	allIconNamesSorted,
	getLucideIcon,
	getLucideName,
	type IconName,
	toLucideIconRef,
} from "@/utils/lucideIcons";

import { Button } from "./Button";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover";

export function IconPicker({
	currentIconUrl,
	onSelect,
	children,
}: {
	currentIconUrl?: string;
	onSelect: (iconUrl: string, iconImageType: string) => void;
	children: React.ReactElement;
}) {
	const [open, setOpen] = useState(false);
	const selectedName = getLucideName(currentIconUrl);

	function select(name: IconName) {
		onSelect(toLucideIconRef(name), "image/svg+xml");
		setOpen(false);
	}

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>{children}</PopoverTrigger>
			<PopoverContent side="bottom" align="start" className="w-72">
				<div className="grid grid-cols-6 gap-1">
					{allIconNamesSorted.map((name) => {
						const Icon = getLucideIcon(name);
						return (
							<Button
								key={name}
								tooltip={name}
								aria-pressed={name === selectedName}
								size="icon"
								variant={name === selectedName ? "primary" : "ghost"}
								onClick={() => select(name)}
							>
								<Icon />
							</Button>
						);
					})}
				</div>
			</PopoverContent>
		</Popover>
	);
}
