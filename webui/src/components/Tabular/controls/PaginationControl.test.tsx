// @vitest-environment happy-dom
import { fireEvent, render, screen } from "@testing-library/react";

import { TooltipProvider } from "@/components";

import { PaginationControl } from "./PaginationControl";

describe("PaginationControl", () => {
	it("supports search results without claiming an exact total", () => {
		const onPageIndexChange = vi.fn();

		render(
			<TooltipProvider>
				<PaginationControl
					pageIndex={0}
					pageSize={10}
					visibleRows={10}
					hasNextPage
					onPageIndexChange={onPageIndexChange}
					onPageSizeChange={vi.fn()}
				/>
			</TooltipProvider>,
		);

		expect(
			screen.getByText("Displaying 1-10 matching results"),
		).toBeInTheDocument();
		expect(screen.queryByRole("button", { name: "Last page" })).toBeNull();

		fireEvent.click(screen.getByRole("button", { name: "Next page" }));
		expect(onPageIndexChange).toHaveBeenCalledWith(1);
	});
});
