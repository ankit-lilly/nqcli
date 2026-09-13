import {
	ChevronFirstIcon,
	ChevronLastIcon,
	ChevronLeftIcon,
	ChevronRightIcon,
} from "lucide-react";

import { Button, Label, SelectField, toHumanString } from "@/components";
import { cn } from "@/utils";

export type PaginationControlProps = {
	className?: string;
	totalRows?: number;
	visibleRows?: number;
	hasNextPage?: boolean;
	pageIndex: number;
	onPageIndexChange(pageIndex: number): void;
	pageSize: number;
	onPageSizeChange(pageSize: number): void;
	pageOptions?: number[];
};

const DEFAULT_PAGE_OPTIONS = [10, 20, 50];
export function PaginationControl({
	className,
	totalRows,
	visibleRows,
	hasNextPage,
	pageIndex,
	pageSize,
	onPageIndexChange,
	onPageSizeChange,
	pageOptions = DEFAULT_PAGE_OPTIONS,
}: PaginationControlProps) {
	const hasKnownTotal = totalRows !== undefined;
	const knownTotal = totalRows ?? 0;
	const pageCount = hasKnownTotal ? Math.ceil(knownTotal / pageSize) : 0;

	const pagesToRender = (() => {
		if (!hasKnownTotal) return [];
		const pages: string[] = [];
		let startIndex = pageIndex - 2;
		let endIndex = pageIndex + 2;

		if (startIndex < 0) {
			const delta = Math.abs(startIndex);
			startIndex = 0;
			endIndex = delta > endIndex ? pageCount - 1 : endIndex + delta;
		}

		if (endIndex > pageCount - 1) {
			const delta = endIndex - pageCount - 1;
			endIndex = pageCount - 1;
			startIndex = delta > startIndex ? 0 : startIndex - delta - 2;
		}

		for (let i = startIndex; i <= endIndex; i++) {
			if (i + 1 > 0) {
				pages.push((i + 1).toString());
			}
		}

		return pages;
	})();

	const pageRange = (() => {
		const to = pageIndex * pageSize + pageSize;
		return `${pageIndex * pageSize + 1}-${to < knownTotal ? to : knownTotal}`;
	})();
	const displayedRows =
		visibleRows ??
		(hasKnownTotal
			? Math.max(0, Math.min(pageSize, knownTotal - pageIndex * pageSize))
			: 0);
	const hasResults = displayedRows > 0;
	const canGoBack = pageIndex > 0;
	const canGoForward = hasKnownTotal
		? pageIndex + 1 < pageCount
		: Boolean(hasNextPage);

	return (
		<div
			className={cn(
				"flex w-full flex-wrap items-center justify-between gap-2",
				className,
			)}
		>
			<div className="text-muted-foreground">
				{hasKnownTotal
					? knownTotal > 0
						? `Displaying ${pageRange} of ${toHumanString(knownTotal)} results`
						: "No results"
					: hasResults
						? `Displaying ${pageIndex * pageSize + 1}-${pageIndex * pageSize + displayedRows} matching results`
						: "No matching results"}
			</div>

			{(hasResults || canGoBack) && (
				<div className="flex flex-row items-center gap-1">
					<Label className="shrink-0">Page size:</Label>
					<SelectField
						options={pageOptions.map((pageOption) => ({
							label: pageOption.toString(),
							value: pageOption.toString(),
						}))}
						value={pageSize.toString()}
						onValueChange={(value) => onPageSizeChange(parseInt(value, 10))}
					/>
					<Button
						disabled={!canGoBack}
						variant="ghost"
						size="icon-small"
						tooltip="First page"
						onClick={() => onPageIndexChange(0)}
					>
						<ChevronFirstIcon />
					</Button>
					<Button
						disabled={!canGoBack}
						variant="ghost"
						size="icon-small"
						tooltip="Previous page"
						onClick={() => onPageIndexChange(pageIndex - 1)}
					>
						<ChevronLeftIcon />
					</Button>
					{pagesToRender.map((page) => {
						const isCurrentPage = pageIndex === parseInt(page, 10) - 1;
						return (
							<Button
								key={page}
								size="icon-small"
								variant={isCurrentPage ? "primary" : undefined}
								onClick={() => onPageIndexChange(parseInt(page, 10) - 1)}
							>
								{page}
							</Button>
						);
					})}
					{!hasKnownTotal && (
						<Button size="icon-small" variant="primary" disabled>
							{pageIndex + 1}
						</Button>
					)}
					<Button
						disabled={!canGoForward}
						variant="ghost"
						size="icon-small"
						tooltip="Next page"
						onClick={() => onPageIndexChange(pageIndex + 1)}
					>
						<ChevronRightIcon />
					</Button>
					{hasKnownTotal && (
						<Button
							disabled={!canGoForward}
							variant="ghost"
							size="icon-small"
							tooltip="Last page"
							onClick={() => onPageIndexChange(pageCount - 1)}
						>
							<ChevronLastIcon />
						</Button>
					)}
				</div>
			)}
		</div>
	);
}

export default PaginationControl;
