import { useAtomValue } from "jotai";

import {
	DetailsIcon,
	ExpandGraphIcon,
	FilterIcon,
	SearchIcon,
	StylingIcon,
} from "@/components";
import {
	CollapsibleSidebar,
	SidebarTabs,
	SidebarTabsContent,
	SidebarTabsList,
	SidebarTabsTrigger,
} from "@/components/CollapsibleSidebar";
import { useGraphViewSidebar } from "@/core";
import { totalFilteredCount } from "@/core/StateProvider/filterCount";
import EntitiesFilter from "@/modules/EntitiesFilter";
import EntityDetails from "@/modules/EntityDetails";
import NodeExpand from "@/modules/NodeExpand";
import { SearchSidebarPanel } from "@/modules/SearchSidebar";
import { graphViewStylesTabAtom, Styles } from "@/modules/Styles";
import { LABELS } from "@/utils";

export function Sidebar() {
	const filteredEntitiesCount = useAtomValue(totalFilteredCount);
	const {
		activeSidebarItem,
		isSidebarOpen,
		sidebarWidth,
		setSidebarWidth,
		closeSidebar,
		toggleSidebar,
	} = useGraphViewSidebar();

	return (
		<CollapsibleSidebar
			isSidebarOpen={isSidebarOpen}
			sidebarWidth={sidebarWidth}
			onResize={setSidebarWidth}
		>
			<SidebarTabs value={activeSidebarItem ?? ""} orientation="vertical">
				<SidebarTabsList>
					<SidebarTabsTrigger
						value="search"
						title="Search"
						onToggle={() => toggleSidebar("search")}
					>
						<SearchIcon />
					</SidebarTabsTrigger>
					<SidebarTabsTrigger
						value="details"
						title="Details"
						onToggle={() => toggleSidebar("details")}
					>
						<DetailsIcon />
					</SidebarTabsTrigger>
					<SidebarTabsTrigger
						value="expand"
						title="Expand"
						onToggle={() => toggleSidebar("expand")}
					>
						<ExpandGraphIcon />
					</SidebarTabsTrigger>
					<SidebarTabsTrigger
						value="filters"
						title="Filters"
						badge={filteredEntitiesCount > 0}
						onToggle={() => toggleSidebar("filters")}
					>
						<FilterIcon />
					</SidebarTabsTrigger>
					<SidebarTabsTrigger
						value="styles"
						title={LABELS.SIDEBAR.STYLES}
						onToggle={() => toggleSidebar("styles")}
					>
						<StylingIcon />
					</SidebarTabsTrigger>
				</SidebarTabsList>
				<SidebarTabsContent value="search">
					<SearchSidebarPanel />
				</SidebarTabsContent>
				<SidebarTabsContent value="details">
					<EntityDetails />
				</SidebarTabsContent>
				<SidebarTabsContent value="expand">
					<NodeExpand />
				</SidebarTabsContent>
				<SidebarTabsContent value="filters">
					<EntitiesFilter />
				</SidebarTabsContent>
				<SidebarTabsContent value="styles">
					<Styles onClose={closeSidebar} tabAtom={graphViewStylesTabAtom} />
				</SidebarTabsContent>
			</SidebarTabs>
		</CollapsibleSidebar>
	);
}
