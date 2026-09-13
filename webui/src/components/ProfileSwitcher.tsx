import { useEffect, useMemo, useState } from "react";

import { createProfileApplication } from "@/bootstrap";
import type { ProfileInfo } from "@/domain";

const profileService = createProfileApplication();

const ENV_COLORS: Record<string, string> = {
	dev: "bg-emerald-500",
	qa: "bg-amber-500",
	prod: "bg-red-500",
};

export function ProfileSwitcher() {
	const [info, setInfo] = useState<ProfileInfo | null>(null);
	const [switching, setSwitching] = useState(false);
	const [error, setError] = useState("");

	useEffect(() => {
		const controller = new AbortController();
		profileService
			.get({ signal: controller.signal })
			.then((data) => {
				if (!controller.signal.aborted) {
					setInfo({ ...data, profiles: data.profiles ?? [] });
				}
			})
			.catch((err) => {
				if (!controller.signal.aborted) {
					setError(err instanceof Error ? err.message : String(err));
				}
			});
		return () => {
			controller.abort();
		};
	}, []);

	const profiles = useMemo(() => {
		if (!info) {
			return [];
		}
		const names = new Set(info.profiles);
		if (info.profile) {
			names.add(info.profile);
		}
		return Array.from(names).sort((a, b) => a.localeCompare(b));
	}, [info]);

	async function switchProfile(profile: string) {
		if (!info || switching || profile === info.profile) {
			return;
		}

		setSwitching(true);
		setError("");
		try {
			await profileService.switchTo(profile);
			window.location.assign("/explorer/#/graph-explorer");
		} catch (err) {
			setError(err instanceof Error ? err.message : String(err));
			setSwitching(false);
		}
	}

	if (!info && !error) {
		return (
			<div className="text-muted-foreground flex items-center gap-2 text-sm">
				<span className="size-2 rounded-full bg-muted-foreground/40" />
				Loading profiles
			</div>
		);
	}

	return (
		<div className="flex items-center gap-2">
			<span
				className={`size-2 rounded-full ${ENV_COLORS[info?.env ?? "dev"] ?? ENV_COLORS.dev}`}
				aria-hidden="true"
			/>
			<select
				aria-label="AWS profile"
				value={info?.profile ?? ""}
				disabled={!info || switching}
				onChange={(event) => void switchProfile(event.target.value)}
				className="border-input-border bg-input-background text-foreground focus-visible:border-primary h-9 max-w-48 rounded-md border px-2 text-sm shadow-xs outline-hidden disabled:opacity-50"
			>
				{profiles.length === 0 ? (
					<option value="">default</option>
				) : (
					profiles.map((profile) => (
						<option key={profile} value={profile}>
							{profile || "default"}
						</option>
					))
				)}
			</select>
			{switching && (
				<span className="text-muted-foreground text-sm">Switching...</span>
			)}
			{error && (
				<span
					className="text-danger-foreground max-w-64 truncate text-sm"
					title={error}
				>
					{error}
				</span>
			)}
		</div>
	);
}
