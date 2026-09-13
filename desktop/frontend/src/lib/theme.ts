const STORAGE_KEY = "nqcli-theme";

export const THEMES = ["light", "dracula"] as const;

export type Theme = (typeof THEMES)[number];

export function detectInitialTheme(): Theme {
	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored === "dracula") return "dracula";
	if (stored === "light") return "light";
	return window.matchMedia("(prefers-color-scheme: dark)").matches
		? "dracula"
		: "light";
}

export function applyTheme(theme: Theme) {
	document.documentElement.setAttribute("data-theme", theme);
	localStorage.setItem(STORAGE_KEY, theme);
}
