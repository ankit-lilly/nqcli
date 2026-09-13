import { type Accessor, type Setter } from "solid-js";
import type { Theme } from "../lib/theme";

interface Props {
	theme: Accessor<Theme>;
	setTheme: Setter<Theme>;
}

export default function ThemeSwitcher(props: Props) {
	const isDark = () => props.theme() === "dracula";

	function toggle() {
		props.setTheme(isDark() ? "light" : "dracula");
	}

	return (
		<button
			class="btn btn-ghost btn-xs gap-1.5 no-drag"
			onClick={toggle}
			title={isDark() ? "Switch to light mode" : "Switch to dark mode"}
		>
			<svg
				width="14"
				height="14"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				{isDark() ? (
					<>
						<circle cx="12" cy="12" r="4" />
						<path d="M12 2v2" />
						<path d="M12 20v2" />
						<path d="m4.93 4.93 1.41 1.41" />
						<path d="m17.66 17.66 1.41 1.41" />
						<path d="M2 12h2" />
						<path d="M20 12h2" />
						<path d="m6.34 17.66-1.41 1.41" />
						<path d="m19.07 4.93-1.41 1.41" />
					</>
				) : (
					<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
				)}
			</svg>
			<span class="text-[11px]">{isDark() ? "Light" : "Dark"}</span>
		</button>
	);
}
