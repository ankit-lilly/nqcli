import { DesktopService } from "../../bindings/github.com/ankit-lilly/nqcli/internal/desktop";
import { createSignal, onMount, Show, For } from "solid-js";

interface Props {
	onSwitchStart: () => Promise<void>;
	onSwitchEnd: () => void;
	onProfileChange?: (profile: string, env: string) => void;
}

interface ProfileInfo {
	profile: string;
	env: string;
	profiles: string[];
}

const ENV_COLORS: Record<string, string> = {
	dev: "oklch(0.72 0.19 142)",
	qa: "oklch(0.75 0.18 55)",
	prod: "oklch(0.63 0.24 25)",
};

export default function ProfileSwitcher(props: Props) {
	const [info, setInfo] = createSignal<ProfileInfo>({
		profile: "",
		env: "dev",
		profiles: [],
	});
	const [switching, setSwitching] = createSignal(false);
	const [error, setError] = createSignal("");
	const [loaded, setLoaded] = createSignal(false);
	let detailsRef: HTMLDetailsElement | undefined;

	onMount(async () => {
		try {
			const result = await DesktopService.GetProfile();
			setInfo({ ...result, profiles: result.profiles || [] });
			props.onProfileChange?.(result.profile, result.env);
		} catch (e) {
			setError(String(e));
		} finally {
			setLoaded(true);
		}
	});

	async function switchTo(profile: string) {
		if (switching() || profile === info().profile) return;
		setSwitching(true);
		setError("");
		(document.activeElement as HTMLElement)?.blur();
		try {
			await props.onSwitchStart();
			const resp = await DesktopService.SwitchProfile({ profile });
			if (resp.error) throw new Error(resp.error);
			setInfo((prev) => ({ ...prev, profile: resp.profile, env: resp.env }));
			props.onProfileChange?.(resp.profile, resp.env);
		} catch (e) {
			setError(e instanceof Error ? e.message : String(e));
		} finally {
			setSwitching(false);
			props.onSwitchEnd();
		}
	}

	const envColor = () => ENV_COLORS[info().env] || ENV_COLORS.dev;

	return (
		<Show
			when={loaded()}
			fallback={
				<div class="profile-switcher btn btn-ghost btn-xs gap-1.5 no-drag">
					<span class="loading loading-spinner loading-xs" />
					<span class="text-base-content/50 text-[11px]">Loading…</span>
				</div>
			}
		>
			<details
				class="profile-switcher dropdown"
				ref={(el) => {
					detailsRef = el;
				}}
			>
				<Show when={error()}>
					<span role="alert" class="text-xs text-error">
						{error()}
					</span>
				</Show>
				<summary class="btn btn-ghost btn-xs gap-1.5 list-none [&::-webkit-details-marker]:hidden">
					<Show
						when={!switching()}
						fallback={
							<>
								<span class="loading loading-spinner loading-xs" />
								<span class="text-base-content/50">Switching...</span>
							</>
						}
					>
						<span
							class="w-2 h-2 rounded-full shrink-0"
							style={{ "background-color": envColor() }}
						/>
						<span>{info().profile || "default"}</span>
						<span
							class="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded"
							style={{
								color: envColor(),
								"background-color": `color-mix(in oklch, ${envColor()} 15%, transparent)`,
							}}
						>
							{info().env}
						</span>
						<svg
							class="w-3 h-3 opacity-40"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="m6 9 6 6 6-6" />
						</svg>
					</Show>
				</summary>

				<ul class="dropdown-content menu bg-neutral text-neutral-content rounded-box z-[100] w-52 p-2 shadow-xl mt-2">
					<li class="menu-title text-[10px] text-neutral-content/50">
						AWS Profiles
					</li>
					<For each={info().profiles}>
						{(profile) => (
							<li>
								<button
									onClick={() => {
										switchTo(profile);
										if (detailsRef) detailsRef.open = false;
									}}
									class={profile === info().profile ? "active" : ""}
								>
									<span
										class="w-1.5 h-1.5 rounded-full shrink-0"
										style={{
											"background-color":
												profile === info().profile ? envColor() : "transparent",
										}}
									/>
									<span class="text-xs">{profile}</span>
								</button>
							</li>
						)}
					</For>
					<Show when={info().profiles.length === 0}>
						<li class="disabled">
							<span class="text-xs">No profiles found</span>
						</li>
					</Show>
				</ul>
			</details>
		</Show>
	);
}
