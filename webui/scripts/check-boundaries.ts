const layers = [
	{
		name: "domain",
		root: "src/domain",
		forbidden: [
			"@/application",
			"@/adapters",
			"@/bootstrap",
			"@/components",
			"@/connector",
			"@/core",
			"@/hooks",
			"@/modules",
			"@/routes",
			"jotai",
			"react",
			"cytoscape",
			"@tanstack/",
		],
	},
	{
		name: "application",
		root: "src/application",
		forbidden: [
			"@/adapters",
			"@/bootstrap",
			"@/components",
			"@/connector",
			"@/core",
			"@/hooks",
			"@/modules",
			"@/routes",
			"jotai",
			"react",
			"cytoscape",
			"@tanstack/",
		],
	},
	{
		name: "cytoscape adapter",
		root: "src/adapters/cytoscape",
		forbidden: [
			"@/bootstrap",
			"@/components",
			"@/connector",
			"@/core",
			"@/hooks",
			"@/modules",
			"@/routes",
			"jotai",
			"react",
			"@tanstack/",
		],
	},
] as const;

const importPattern = /(?:from\s+|import\s*\()(["'])([^"']+)\1/g;
const violations: string[] = [];

for (const layer of layers) {
	const glob = new Bun.Glob("**/*.{ts,tsx}");
	for await (const relativePath of glob.scan(layer.root)) {
		const path = `${layer.root}/${relativePath}`;
		const source = await Bun.file(path).text();

		for (const match of source.matchAll(importPattern)) {
			const specifier = match[2];
			const forbidden = layer.forbidden.find(
				(prefix) => specifier === prefix || specifier.startsWith(prefix),
			);
			if (forbidden) {
				violations.push(
					`${path}: ${layer.name} cannot import ${JSON.stringify(specifier)}`,
				);
			}
		}
	}
}

if (violations.length > 0) {
	console.error(violations.join("\n"));
	process.exit(1);
}

console.log("Architecture boundaries are valid");
