// @vitest-environment happy-dom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import { Provider } from "jotai";
import { beforeEach, expect, test, vi } from "vitest";

import {
	activeConfigurationAtom,
	type AppStore,
	configurationAtom,
	getAppStore,
} from "@/core";
import { createRandomRawConfiguration } from "@/utils/testing";

import AppStatusLoader from "./AppStatusLoader";
import * as defaultConnection from "./defaultConnection";

function renderLoader(store: AppStore) {
	return render(
		<QueryClientProvider client={new QueryClient()}>
			<Provider store={store}>
				<AppStatusLoader>
					<div>ready</div>
				</AppStatusLoader>
			</Provider>
		</QueryClientProvider>,
	);
}

beforeEach(() => vi.restoreAllMocks());

test("loads and selects the server connection", async () => {
	const config = createRandomRawConfiguration();
	vi.spyOn(defaultConnection, "fetchDefaultConnection").mockResolvedValue(
		config,
	);
	const store = getAppStore();
	store.set(configurationAtom, new Map());
	store.set(activeConfigurationAtom, null);

	const { findByText } = renderLoader(store);

	await findByText("ready");
	expect(store.get(configurationAtom).get(config.id)).toBe(config);
	expect(store.get(activeConfigurationAtom)).toBe(config.id);
});

test("preserves cached profiles while selecting the server profile", async () => {
	const cached = createRandomRawConfiguration();
	const current = createRandomRawConfiguration();
	vi.spyOn(defaultConnection, "fetchDefaultConnection").mockResolvedValue(
		current,
	);
	const store = getAppStore();
	store.set(configurationAtom, new Map([[cached.id, cached]]));
	store.set(activeConfigurationAtom, cached.id);

	const { findByText } = renderLoader(store);

	await findByText("ready");
	expect(store.get(configurationAtom).size).toBe(2);
	expect(store.get(activeConfigurationAtom)).toBe(current.id);
});
