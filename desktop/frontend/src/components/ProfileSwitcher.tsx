import Popover from "@corvu/popover";
import { Check, ChevronDown } from "lucide-solid";
import { For, Show, createSignal, onMount } from "solid-js";
import { DesktopService } from "../../bindings/github.com/ankit-lilly/nqcli/internal/desktop";

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
	const [open, setOpen] = createSignal(false);

	onMount(async () => {
		try {
			const result = await DesktopService.GetProfile();
			setInfo({ ...result, profiles: result.profiles || [] });
			props.onProfileChange?.(result.profile, result.env);
		} catch (cause) {
			setError(String(cause));
		} finally {
			setLoaded(true);
		}
	});

	async function switchTo(profile: string) {
		if (switching() || profile === info().profile) return;
		setOpen(false);
		setSwitching(true);
		setError("");
		try {
			await props.onSwitchStart();
			const response = await DesktopService.SwitchProfile({ profile });
			if (response.error) throw new Error(response.error);
			setInfo((current) => ({
				...current,
				profile: response.profile,
				env: response.env,
			}));
			props.onProfileChange?.(response.profile, response.env);
		} catch (cause) {
			setError(cause instanceof Error ? cause.message : String(cause));
		} finally {
			setSwitching(false);
			props.onSwitchEnd();
		}
	}

	const envColor = () => ENV_COLORS[info().env] || ENV_COLORS.dev;

	return (
		<div class="profile-switcher no-drag">
			<Show
				when={loaded()}
				fallback={
					<div class="btn btn-ghost btn-xs gap-1.5">
						<span class="loading loading-spinner loading-xs" />
						<span class="text-[11px] text-base-content/50">Loading...</span>
					</div>
				}
			>
				<Popover
					open={open()}
					onOpenChange={setOpen}
					placement="bottom-start"
					strategy="fixed"
					floatingOptions={{ offset: 6, flip: true, shift: { padding: 8 } }}
				>
					<Popover.Trigger
						type="button"
						class="btn btn-ghost btn-xs gap-1.5"
						disabled={switching()}
					>
						<Show
							when={!switching()}
							fallback={<span class="loading loading-spinner loading-xs" />}
						>
							<span
								class="size-2 shrink-0 rounded-full"
								style={{ "background-color": envColor() }}
							/>
							<span>{info().profile || "default"}</span>
							<span
								class="rounded bg-base-content/5 px-1.5 py-0.5 text-[10px] font-bold uppercase"
								style={{ color: envColor() }}
							>
								{info().env}
							</span>
							<ChevronDown size={12} class="opacity-45" />
						</Show>
					</Popover.Trigger>
					<Popover.Portal>
						<Popover.Content class="z-[1000] w-52 border border-base-300 bg-base-100 p-1.5 text-base-content shadow-xl">
							<Popover.Label class="px-2 py-1 text-[10px] font-semibold uppercase text-base-content/45">
								AWS profiles
							</Popover.Label>
							<For each={info().profiles}>
								{(profile) => (
									<button
										type="button"
										class={`flex w-full items-center gap-2 px-2 py-1.5 text-left text-xs hover:bg-base-200 ${profile === info().profile ? "bg-primary/10 text-primary" : ""}`}
										onClick={() => void switchTo(profile)}
									>
										<span class="w-3">
											{profile === info().profile ? <Check size={12} /> : null}
										</span>
										<span class="truncate">{profile}</span>
									</button>
								)}
							</For>
							<Show when={info().profiles.length === 0}>
								<p class="px-2 py-2 text-xs text-base-content/45">
									No profiles found
								</p>
							</Show>
						</Popover.Content>
					</Popover.Portal>
				</Popover>
			</Show>
			<Show when={error()}>
				<span role="alert" class="ml-2 text-xs text-error">
					{error()}
				</span>
			</Show>
		</div>
	);
}
