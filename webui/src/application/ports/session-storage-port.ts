export interface SessionStoragePort<Session> {
	load(key: string): Promise<Session | undefined>;
	save(key: string, session: Session): Promise<void>;
	remove(key: string): Promise<void>;
}
