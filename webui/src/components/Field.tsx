import { cva, type VariantProps } from "cva";
import type { ComponentPropsWithRef } from "react";

import { cn } from "@/utils";

import { Label } from "./Label";

function FieldSet({ className, ...props }: ComponentPropsWithRef<"fieldset">) {
	return (
		<fieldset
			data-slot="field-set"
			className={cn(
				"flex min-w-0 flex-col gap-6",
				"has-[>[data-slot=checkbox-group]]:gap-3 has-[>[data-slot=radio-group]]:gap-3",
				className,
			)}
			{...props}
		/>
	);
}

function FieldLegend({
	className,
	variant = "legend",
	...props
}: ComponentPropsWithRef<"legend"> & { variant?: "legend" | "label" }) {
	return (
		<legend
			data-slot="field-legend"
			data-variant={variant}
			className={cn(
				"mb-3 font-medium",
				"data-[variant=legend]:text-base",
				"data-[variant=label]:text-sm",
				className,
			)}
			{...props}
		/>
	);
}

function FieldGroup({ className, ...props }: ComponentPropsWithRef<"div">) {
	return (
		<div
			data-slot="field-group"
			className={cn(
				"group/field-group @container/field-group flex w-full flex-col gap-7 data-[slot=checkbox-group]:gap-3 *:data-[slot=field-group]:gap-4",
				className,
			)}
			{...props}
		/>
	);
}

const fieldVariants = cva({
	base: "group/field data-[invalid=true]:text-danger-foreground flex w-full gap-3",
	variants: {
		orientation: {
			vertical: ["flex-col *:w-full [&>.sr-only]:w-auto"],
			horizontal: [
				"flex-row items-center",
				"*:data-[slot=field-label]:flex-auto",
				"has-[>[data-slot=field-content]]:items-start has-[>[data-slot=field-content]]:[&>[role=checkbox],[role=radio]]:mt-px",
			],
			responsive: [
				"flex-col *:w-full @md/field-group:flex-row @md/field-group:items-center @md/field-group:*:w-auto [&>.sr-only]:w-auto",
				"@md/field-group:*:data-[slot=field-label]:flex-auto",
				"@md/field-group:has-[>[data-slot=field-content]]:items-start @md/field-group:has-[>[data-slot=field-content]]:[&>[role=checkbox],[role=radio]]:mt-px",
			],
		},
	},
	defaultVariants: {
		orientation: "vertical",
	},
});

function Field({
	className,
	orientation = "vertical",
	...props
}: ComponentPropsWithRef<"div"> & VariantProps<typeof fieldVariants>) {
	return (
		<div
			role="group"
			data-slot="field"
			data-orientation={orientation}
			className={cn(fieldVariants({ orientation }), className)}
			{...props}
		/>
	);
}

function FieldLabel({
	className,
	...props
}: ComponentPropsWithRef<typeof Label>) {
	return (
		<Label
			data-slot="field-label"
			className={cn(
				"group/field-label peer/field-label flex w-fit gap-2 leading-snug group-data-[disabled=true]/field:opacity-50",
				"has-[>[data-slot=field]]:w-full has-[>[data-slot=field]]:flex-col has-[>[data-slot=field]]:rounded-md has-[>[data-slot=field]]:border *:data-[slot=field]:p-4",
				"has-data-[state=checked]:bg-primary/5 has-data-[state=checked]:border-primary",
				className,
			)}
			{...props}
		/>
	);
}

export { Field, FieldLabel, FieldGroup, FieldLegend, FieldSet };
