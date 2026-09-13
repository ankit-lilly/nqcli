// @vitest-environment happy-dom
import type { PropsWithChildren } from "react";

import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router";
import { describe, expect, test } from "vitest";

import { RouteButtonGroup } from "./RouteButton";
import { TooltipProvider } from "./Tooltip";

function LocationDisplay() {
	const location = useLocation();
	return <div data-testid="location">{location.pathname}</div>;
}

function Wrapper({ children }: PropsWithChildren) {
	return <TooltipProvider>{children}</TooltipProvider>;
}

describe("RouteButtonGroup", () => {
	test("should render all route buttons", () => {
		render(
			<MemoryRouter initialEntries={["/graph-explorer"]}>
				<RouteButtonGroup active="graph-explorer" />
			</MemoryRouter>,
			{ wrapper: Wrapper },
		);

		expect(screen.getByRole("button", { name: "Graph" })).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Data Table" }),
		).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Schema" })).toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "Settings" }),
		).not.toBeInTheDocument();
	});

	test("should set data-active on the active route button", () => {
		render(
			<MemoryRouter initialEntries={["/data-explorer"]}>
				<RouteButtonGroup active="data-explorer" />
			</MemoryRouter>,
			{ wrapper: Wrapper },
		);

		expect(screen.getByRole("button", { name: "Data Table" })).toHaveAttribute(
			"data-active",
		);
		expect(screen.getByRole("button", { name: "Graph" })).not.toHaveAttribute(
			"data-active",
		);
		expect(screen.getByRole("button", { name: "Schema" })).not.toHaveAttribute(
			"data-active",
		);
	});

	test("should navigate when clicking a route button", async () => {
		const user = userEvent.setup();

		render(
			<MemoryRouter initialEntries={["/graph-explorer"]}>
				<RouteButtonGroup active="graph-explorer" />
				<LocationDisplay />
			</MemoryRouter>,
			{ wrapper: Wrapper },
		);

		expect(screen.getByTestId("location")).toHaveTextContent("/graph-explorer");

		await user.click(screen.getByRole("button", { name: "Schema" }));

		expect(screen.getByTestId("location")).toHaveTextContent(
			"/schema-explorer",
		);
	});

	test("should render nav menu trigger button", () => {
		render(
			<MemoryRouter initialEntries={["/graph-explorer"]}>
				<RouteButtonGroup active="graph-explorer" />
			</MemoryRouter>,
			{ wrapper: Wrapper },
		);

		expect(
			screen.getByRole("button", { name: "Navigation" }),
		).toBeInTheDocument();
	});

	test("should open nav menu with all routes on click", async () => {
		const user = userEvent.setup();

		render(
			<MemoryRouter initialEntries={["/graph-explorer"]}>
				<RouteButtonGroup active="graph-explorer" />
			</MemoryRouter>,
			{ wrapper: Wrapper },
		);

		await user.click(screen.getByRole("button", { name: "Navigation" }));

		const menu = screen.getByRole("menu");
		expect(
			within(menu).getByRole("menuitem", { name: "Graph" }),
		).toBeInTheDocument();
		expect(
			within(menu).getByRole("menuitem", { name: "Data Table" }),
		).toBeInTheDocument();
		expect(
			within(menu).getByRole("menuitem", { name: "Schema" }),
		).toBeInTheDocument();
		expect(
			within(menu).queryByRole("menuitem", { name: "Settings" }),
		).not.toBeInTheDocument();
	});

	test("should navigate when selecting a route from nav menu", async () => {
		const user = userEvent.setup();

		render(
			<MemoryRouter initialEntries={["/graph-explorer"]}>
				<RouteButtonGroup active="graph-explorer" />
				<LocationDisplay />
			</MemoryRouter>,
			{ wrapper: Wrapper },
		);

		expect(screen.getByTestId("location")).toHaveTextContent("/graph-explorer");

		await user.click(screen.getByRole("button", { name: "Navigation" }));
		await user.click(screen.getByRole("menuitem", { name: "Schema" }));

		expect(screen.getByTestId("location")).toHaveTextContent(
			"/schema-explorer",
		);
	});
});
