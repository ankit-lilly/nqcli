import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
	plugins: [solid(), tailwindcss()],
	resolve: {
		preserveSymlinks: true,
	},
	build: {
		outDir: "dist",
		emptyOutDir: true,
		rollupOptions: {
			external: [/^\/wails\//],
		},
	},
	server: {
		port: 5173,
		strictPort: true,
	},
});
