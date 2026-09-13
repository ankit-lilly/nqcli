import {
	type Accessor,
	type Setter,
	onMount,
	onCleanup,
	createEffect,
} from "solid-js";
import {
	EditorView,
	keymap,
	placeholder,
	lineNumbers,
	highlightActiveLine,
	highlightActiveLineGutter,
} from "@codemirror/view";
import { EditorState, Compartment } from "@codemirror/state";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { StreamLanguage } from "@codemirror/language";
import { cypher } from "@codemirror/legacy-modes/mode/cypher";
import { groovy } from "@codemirror/legacy-modes/mode/groovy";
import {
	syntaxHighlighting,
	defaultHighlightStyle,
	bracketMatching,
} from "@codemirror/language";
import { oneDark } from "@codemirror/theme-one-dark";
import { searchKeymap, highlightSelectionMatches } from "@codemirror/search";
import { formatQuery } from "../lib/formatQuery";

interface Props {
	query: Accessor<string>;
	setQuery: Setter<string>;
	queryType: Accessor<string>;
	onExecute: () => void;
}

const MONO =
	"'JetBrains Mono', 'SF Mono', 'SFMono-Regular', Consolas, 'Liberation Mono', monospace";

const lightTheme = EditorView.theme({
	"&": { fontSize: "13px", height: "100%" },
	".cm-content": { fontFamily: MONO, caretColor: "var(--color-base-content)" },
	".cm-gutters": {
		backgroundColor: "transparent",
		borderRight: "none",
		color: "color-mix(in oklch, var(--color-base-content) 25%, transparent)",
	},
	".cm-activeLineGutter": {
		backgroundColor: "transparent",
		color: "color-mix(in oklch, var(--color-base-content) 50%, transparent)",
	},
	".cm-activeLine": {
		backgroundColor:
			"color-mix(in oklch, var(--color-base-content) 4%, transparent)",
	},
	".cm-cursor": { borderLeftColor: "var(--color-base-content)" },
	".cm-selectionBackground": {
		backgroundColor:
			"color-mix(in oklch, var(--color-primary) 15%, transparent) !important",
	},
	"&.cm-focused .cm-selectionBackground": {
		backgroundColor:
			"color-mix(in oklch, var(--color-primary) 20%, transparent) !important",
	},
	".cm-scroller": { fontFamily: MONO, lineHeight: "1.6", overflow: "auto" },
});

function isDark(): boolean {
	return document.documentElement.getAttribute("data-theme") === "dracula";
}

export default function QueryEditor(props: Props) {
	let containerRef!: HTMLDivElement;
	let view: EditorView | undefined;
	let skipUpdate = false;
	const themeCompartment = new Compartment();
	const languageCompartment = new Compartment();
	const language = () =>
		StreamLanguage.define(props.queryType() === "cypher" ? cypher : groovy);

	function getThemeExtension() {
		return isDark() ? oneDark : lightTheme;
	}

	onMount(() => {
		const runQuery = keymap.of([
			{
				key: "Mod-Enter",
				run: () => {
					props.onExecute();
					return true;
				},
			},
			{
				key: "Shift-Alt-f",
				run: (v) => {
					const formatted = formatQuery(
						v.state.doc.toString(),
						props.queryType() as "gremlin" | "cypher",
					);
					v.dispatch({
						changes: { from: 0, to: v.state.doc.length, insert: formatted },
					});
					return true;
				},
			},
		]);

		const updateListener = EditorView.updateListener.of((update) => {
			if (update.docChanged && !skipUpdate) {
				props.setQuery(update.state.doc.toString());
			}
		});

		const state = EditorState.create({
			doc: props.query(),
			extensions: [
				lineNumbers(),
				history(),
				bracketMatching(),
				highlightActiveLine(),
				highlightActiveLineGutter(),
				highlightSelectionMatches(),
				syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
				languageCompartment.of(language()),
				EditorView.lineWrapping,
				keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
				runQuery,
				placeholder("Enter a Gremlin or Cypher query..."),
				updateListener,
				themeCompartment.of(getThemeExtension()),
			],
		});

		view = new EditorView({ state, parent: containerRef });
		view.focus();

		const observer = new MutationObserver(() => {
			view?.dispatch({
				effects: themeCompartment.reconfigure(getThemeExtension()),
			});
		});
		observer.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ["data-theme"],
		});
		onCleanup(() => observer.disconnect());
	});

	createEffect(() => {
		const q = props.query();
		if (view && view.state.doc.toString() !== q) {
			skipUpdate = true;
			view.dispatch({
				changes: { from: 0, to: view.state.doc.length, insert: q },
			});
			skipUpdate = false;
		}
	});

	createEffect(() => {
		const extension = language();
		view?.dispatch({ effects: languageCompartment.reconfigure(extension) });
	});

	onCleanup(() => view?.destroy());

	return (
		<div
			ref={containerRef!}
			class="min-h-28 max-h-60 rounded-lg border border-base-300 overflow-auto bg-base-100 [&_.cm-editor]:h-full [&_.cm-editor.cm-focused]:outline-none"
		/>
	);
}
