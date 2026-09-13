export function createRandomName(prefix = ""): string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz";
	const random = Array.from(
		{ length: 10 },
		() => chars[Math.floor(Math.random() * chars.length)],
	).join("");

	return `${prefix}${prefix.length > 0 ? "-" : ""}${random}`;
}

export function createRandomBoolean(): boolean {
	return Math.random() < 0.5;
}

export function createRandomInteger(options?: {
	min?: number;
	max?: number;
}): number {
	const min = options?.min ?? 0;
	const max = options?.max ?? 100000;

	return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function createRandomDouble(min?: number, max?: number) {
	if (min !== undefined && max !== undefined) {
		return Math.random() * (max - min) + min;
	}
	return Math.random() * createRandomInteger();
}

export function createRandomColor(): string {
	const letters = "0123456789ABCDEF".split("");
	let color = "#";
	for (let i = 0; i < 6; i++) {
		color += letters[Math.round(Math.random() * 15)];
	}
	return color;
}

export function createRandomUrlString(): string {
	const scheme = pickRandomElement(["http", "https"]);
	const host = createRandomName("host");
	const port = pickRandomElement([
		"",
		`:${createRandomInteger({ max: 30000 })}`,
	]);
	const path = pickRandomElement(["", `/${createRandomName("path")}`]);
	return `${scheme}://${host}${port}${path}`;
}

export function createRandomDate(start?: Date, end?: Date): Date {
	const startTime = start ? start.getTime() : new Date(1970, 0, 1).getTime();
	const endTime = end ? end.getTime() : new Date().getTime();
	const randomTime = startTime + Math.random() * (endTime - startTime);
	return new Date(randomTime);
}

export function randomlyUndefined<T>(value: T): T | undefined {
	return createRandomBoolean() ? value : undefined;
}

export function createArray<T>(length: number, factory: () => T): T[] {
	return Array.from({ length }, factory);
}

export function createRecord<TValue>(
	length: number,
	factory: () => { key: string; value: TValue },
): Record<string, TValue> {
	const result: Record<string, TValue> = {};

	for (let i = 0; i < length; i++) {
		const newEntry = factory();
		result[newEntry.key] = newEntry.value;
	}

	return result;
}

function pickRandomElement<T>(array: T[]): T {
	return array[Math.floor(Math.random() * array.length)];
}
