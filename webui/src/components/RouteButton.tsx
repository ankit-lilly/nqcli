import type { ComponentPropsWithRef } from "react";

import { MenuIcon } from "lucide-react";
import { Link } from "react-router";

import { cn } from "@/utils";

import { Button, NavButton } from "./Button";
import Divider from "./Divider";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "./DropdownMenu";
import { NavBarActions } from "./NavBar";
import { ProfileSwitcher } from "./ProfileSwitcher";

const allRoutes = {
	"graph-explorer": { name: "Graph", path: "/graph-explorer" },
	"data-explorer": { name: "Data Table", path: "/data-explorer" },
	"schema-explorer": { name: "Schema", path: "/schema-explorer" },
} as const;

type RouteKey = keyof typeof allRoutes;

export function RouteButtonGroup({ active }: { active: RouteKey }) {
	return (
		<>
			<NavBarActions className="hidden lg:flex">
				<div className="flex h-full">
					{Object.entries(allRoutes).map(([key, route]) => (
						<RouteButton key={key} to={route.path} active={active === key}>
							{route.name}
						</RouteButton>
					))}
				</div>

				<Divider axis="vertical" className="h-6" />

				<div className="flex h-full items-center gap-2">
					<ProfileSwitcher />
				</div>
			</NavBarActions>
			<NavBarActions className="lg:hidden">
				<NavMenu />
			</NavBarActions>
		</>
	);
}

function NavMenu() {
	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<Button variant="ghost" size="icon" tooltip="Navigation">
					<MenuIcon />
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent className="w-40">
				<DropdownMenuGroup>
					{Object.entries(allRoutes).map(([key, route]) => (
						<DropdownMenuItem key={key} asChild>
							<Link to={route.path}>{route.name}</Link>
						</DropdownMenuItem>
					))}
				</DropdownMenuGroup>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}

function RouteButton({
	active,
	children,
	className,
	...props
}: ComponentPropsWithRef<typeof NavButton> & { active: boolean }) {
	return (
		<NavButton
			variant="ghost"
			size="default"
			data-active={active || undefined}
			className={cn(
				"hover:text-primary-foreground text-muted-foreground data-active:border-b-primary-foreground data-active:text-primary-foreground h-full cursor-pointer rounded-none border-y-2 border-y-transparent px-5 font-normal hover:bg-transparent data-active:bg-transparent data-active:font-semibold",
				className,
			)}
			{...props}
		>
			{/* The spans are in place to fix the layout shift that occurs when switching the font weight */}
			<span className="grid place-items-center">
				<span className="col-start-1 row-start-1">{children}</span>
				<span
					className="invisible col-start-1 row-start-1 font-semibold"
					aria-hidden="true"
				>
					{children}
				</span>
			</span>
		</NavButton>
	);
}
