const INDENT = "  ";

export function formatGremlin(raw: string): string {
	const q = raw.replace(/\s+/g, " ").trim();
	let out = "";
	let depth = 0;
	const stack: boolean[] = [];
	let i = 0;

	while (i < q.length) {
		const ch = q[i];

		if (ch === "'" || ch === '"') {
			const close = q.indexOf(ch, i + 1);
			if (close >= 0) {
				out += q.slice(i, close + 1);
				i = close + 1;
			} else {
				out += q.slice(i);
				break;
			}
			continue;
		}

		if (ch === "." && depth === 0) {
			out += ".\n";
			i++;
			continue;
		}

		if (ch === "(") {
			out += "(";
			i++;
			const nested = hasNestedCall(q, i);
			stack.push(nested);
			if (nested) {
				depth++;
				out += "\n" + INDENT.repeat(depth);
			}
			continue;
		}

		if (ch === ")") {
			const wasNested = stack.pop();
			if (wasNested) {
				depth--;
				out += "\n" + INDENT.repeat(depth);
			}
			out += ")";
			i++;
			continue;
		}

		if (ch === "," && depth > 0) {
			out += ",\n" + INDENT.repeat(depth);
			i++;
			if (q[i] === " ") i++;
			continue;
		}

		out += ch;
		i++;
	}

	return out;
}

function hasNestedCall(s: string, from: number): boolean {
	let parenDepth = 0;
	for (let i = from; i < s.length; i++) {
		const ch = s[i];
		if (ch === "'" || ch === '"') {
			const close = s.indexOf(ch, i + 1);
			if (close >= 0) {
				i = close;
				continue;
			}
			return false;
		}
		if (ch === "(") {
			if (parenDepth === 0) return true;
			parenDepth++;
		}
		if (ch === ")") {
			if (parenDepth === 0) return false;
			parenDepth--;
		}
	}
	return false;
}

export function formatCypher(raw: string): string {
	const keywords =
		/\b(MATCH|WHERE|RETURN|WITH|ORDER BY|LIMIT|SKIP|UNWIND|CREATE|MERGE|DELETE|DETACH DELETE|SET|REMOVE|OPTIONAL MATCH|CALL|YIELD|UNION)\b/gi;
	return raw
		.replace(/\s+/g, " ")
		.trim()
		.replace(keywords, (m) => "\n" + m.toUpperCase())
		.trim();
}

export function formatQuery(raw: string, type: "gremlin" | "cypher"): string {
	return type === "cypher" ? formatCypher(raw) : formatGremlin(raw);
}
