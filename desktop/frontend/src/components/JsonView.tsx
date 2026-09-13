import {
	type Accessor,
	createSignal,
	createMemo,
	onCleanup,
	Show,
} from "solid-js";
import hljs from "highlight.js/lib/core";
import json from "highlight.js/lib/languages/json";
hljs.registerLanguage("json", json);
const PAGE_SIZE = 50_000;
export default function JsonView(props: { result: Accessor<string> }) {
	const [copied, setCopied] = createSignal(false);
	const [page, setPage] = createSignal(0);
	let timer: ReturnType<typeof setTimeout> | undefined;
	onCleanup(() => clearTimeout(timer));
	const pages = createMemo(() =>
		Math.max(1, Math.ceil(props.result().length / PAGE_SIZE)),
	);
	const currentPage = () => Math.min(page(), pages() - 1);
	const preview = createMemo(() =>
		props
			.result()
			.slice(currentPage() * PAGE_SIZE, (currentPage() + 1) * PAGE_SIZE),
	);
	const highlighted = createMemo(() => {
		if (props.result().length > PAGE_SIZE) return null;
		try {
			return hljs.highlight(preview(), { language: "json" }).value;
		} catch {
			return null;
		}
	});
	async function copy() {
		try {
			if (navigator.clipboard?.writeText) {
				await navigator.clipboard.writeText(props.result());
			} else {
				fallbackCopy(props.result());
			}
			setCopied(true);
			clearTimeout(timer);
			timer = setTimeout(() => setCopied(false), 1500);
		} catch {
			try {
				fallbackCopy(props.result());
				setCopied(true);
				clearTimeout(timer);
				timer = setTimeout(() => setCopied(false), 1500);
			} catch {
				setCopied(false);
			}
		}
	}

	function fallbackCopy(text: string) {
		const ta = document.createElement("textarea");
		ta.value = text;
		ta.style.position = "fixed";
		ta.style.opacity = "0";
		document.body.appendChild(ta);
		ta.focus();
		ta.select();
		document.execCommand("copy");
		document.body.removeChild(ta);
	}
	return (
		<div class="h-full flex flex-col">
			<div class="flex items-center gap-2 px-3 py-2 border-b border-base-300 text-xs">
				<Show when={pages() > 1}>
					<span>
						Text preview · page {currentPage() + 1} of {pages()}
					</span>
					<button
						class="btn btn-xs"
						disabled={currentPage() === 0}
						onClick={() => setPage(currentPage() - 1)}
					>
						Previous
					</button>
					<button
						class="btn btn-xs"
						disabled={currentPage() + 1 >= pages()}
						onClick={() => setPage(currentPage() + 1)}
					>
						Next
					</button>
				</Show>
				<button
					onClick={() => void copy()}
					class="btn btn-ghost btn-xs ml-auto"
				>
					{copied() ? "Copied" : "Copy full result"}
				</button>
			</div>
			<div class="flex-1 overflow-auto p-4">
				<pre class="m-0 whitespace-pre-wrap break-words">
					<Show
						when={highlighted() !== null}
						fallback={<code>{preview()}</code>}
					>
						<code class="language-json" innerHTML={highlighted()!} />
					</Show>
				</pre>
			</div>
		</div>
	);
}
