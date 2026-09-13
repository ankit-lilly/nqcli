import { ListTreeIcon } from "lucide-react";
import { DetailsIcon, StylingIcon } from "@/components";
import {
	CollapsibleSidebar,
	SidebarTabs,
	SidebarTabsContent,
	SidebarTabsList,
	SidebarTabsTrigger,
} from "@/components/CollapsibleSidebar";
import { Styles, schemaViewStylesTabAtom } from "@/modules/Styles";
import { LABELS } from "@/utils";

import type {
	SchemaGraphSelection,
	SchemaGraphSelectionItem,
} from "../SchemaGraph";

import { SchemaDetailsContent } from "./SchemaDetailsContent";
import { SchemaInventory } from "./SchemaInventory";
import { useSchemaViewSidebar } from "./schemaViewLayout";

export type SchemaExplorerSidebarProps = {
	selection: SchemaGraphSelection;
	onSelectionChange?: (item: SchemaGraphSelectionItem) => void;
};

export function SchemaExplorerSidebar({
	selection,
	onSelectionChange,
}: SchemaExplorerSidebarProps) {
	const {
		activeSidebarItem,
		isSidebarOpen,
		sidebarWidth,
		setSidebarWidth,
		closeSidebar,
		toggleSidebar,
	} = useSchemaViewSidebar();

	return (
		<CollapsibleSidebar
			isSidebarOpen={isSidebarOpen}
			sidebarWidth={sidebarWidth}
			onResize={setSidebarWidth}
		>
			<SidebarTabs value={activeSidebarItem ?? ""} orientation="vertical">
				<SidebarTabsList>
					<SidebarTabsTrigger
						value="schema"
						title={LABELS.SIDEBAR.SCHEMA}
						onToggle={() => toggleSidebar("schema")}
					>
						<ListTreeIcon />
					</SidebarTabsTrigger>
					<SidebarTabsTrigger
						value="details"
						title={LABELS.SIDEBAR.SELECTION_DETAILS}
						onToggle={() => toggleSidebar("details")}
					>
						<DetailsIcon />
					</SidebarTabsTrigger>
					<SidebarTabsTrigger
						value="styles"
						title={LABELS.SIDEBAR.STYLES}
						onToggle={() => toggleSidebar("styles")}
					>
						<StylingIcon />
					</SidebarTabsTrigger>
				</SidebarTabsList>
				<SidebarTabsContent value="schema">
					<SchemaInventory
						selection={selection}
						onSelectionChange={onSelectionChange}
					/>
				</SidebarTabsContent>
				<SidebarTabsContent value="details">
					<SchemaDetailsContent
						selection={selection}
						onSelectionChange={onSelectionChange}
					/>
				</SidebarTabsContent>
				<SidebarTabsContent value="styles">
					<Styles onClose={closeSidebar} tabAtom={schemaViewStylesTabAtom} />
				</SidebarTabsContent>
			</SidebarTabs>
		</CollapsibleSidebar>
	);
}
