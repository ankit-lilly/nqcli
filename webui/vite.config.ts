import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import { loadEnv } from "vite";
import { coverageConfigDefaults, defineConfig } from "vitest/config";

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), "");

	const backendUrl =
		env.NQ_WEBUI_BACKEND_URL ||
		`http://localhost:${env.NQ_SERVER_PORT || 8080}`;

	return {
		server: {
			host: true,
			port: Number(env.NQ_WEBUI_DEV_PORT) || undefined,
			strictPort: !!env.NQ_WEBUI_DEV_PORT,
			watch: {
				ignored: ["**/*.test.ts", "**/*.test.tsx"],
			},
			proxy: {
				// Forward API requests to nq in dev mode so
				// the browser stays on the same origin and CORS is not needed.
				"^/(defaultConnection|gremlin|logger|openCypher|profiles|schema)(/|$)":
					{
						target: backendUrl,
						changeOrigin: true,
					},
			},
		},
		base: env.NQ_WEBUI_BASE || "/",
		envPrefix: "NQ_WEBUI_",
		define: {
			__NQ_WEBUI_VERSION__: JSON.stringify(process.env.NQ_VERSION ?? "dev"),
		},
		plugins: [
			tailwindcss(),
			react(),
			babel({
				presets: [reactCompilerPreset()],
			}),
		],
		resolve: {
			preserveSymlinks: true,
			tsconfigPaths: true,
		},
		test: {
			globals: true,
			pool: "threads",

			// Setup
			globalSetup: ["src/globalSetup.ts"],
			setupFiles: ["src/setupTests.ts"],

			// Reset state between tests
			clearMocks: true,
			resetMocks: true,
			restoreMocks: true,
			unstubEnvs: true,
			unstubGlobals: true,

			coverage: {
				exclude: [
					"src/components/icons",
					"src/@types",
					"src/index.tsx",
					"src/App.ts",
					"src/setupTests.ts",
					"src/**/*.style.ts",
					"src/**/*.styles.ts",
					"src/**/*.styles.css.ts",
					...coverageConfigDefaults.exclude,
				],
			},
		},
	};
});
