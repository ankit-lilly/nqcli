import { saveAs } from "file-saver";

import { logger } from "@/utils";
import { LABELS } from "@/utils/constants";
import { toJsonFileData } from "@/utils/fileData";

import { serializeData } from "./serializeData";

/** The read-only LocalForage surface needed to create a recovery backup. */
export interface LocalDb {
	keys(): Promise<string[]>;
	getItem<T>(key: string): Promise<T | null>;
}

/** Saves the current browser state when IndexedDB reports a quota failure. */
export async function saveLocalForageToFile(localDb: LocalDb) {
	const backup = await createBackupData(localDb);
	const fileData = toJsonFileData(serializeData(backup));
	saveAs(fileData, "nq-webui.config.json");
}

export type SerializedBackup = {
	backupSource: string;
	backupSourceVersion: string;
	backupVersion: "1.0";
	backupTimestamp: Date;
	data: Record<string, unknown>;
};

export async function createBackupData(
	localDb: LocalDb,
): Promise<SerializedBackup> {
	logger.debug("Gathering entries from localDb");
	const keys = await localDb.keys();
	const entries = await Promise.all(
		keys.map(async (key) => [key, await localDb.getItem(key)] as const),
	);

	return {
		backupSource: LABELS.APP_NAME,
		backupSourceVersion: __NQ_WEBUI_VERSION__,
		backupVersion: "1.0",
		backupTimestamp: new Date(),
		data: Object.fromEntries(entries),
	};
}
