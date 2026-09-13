export type OCProperties = Record<string, string | number | boolean>;

export type OCVertex = {
	"~id": string;
	"~entityType": string;
	"~labels": Array<string>;
	"~properties": OCProperties;
};

export type OCEdge = {
	"~id": string;
	"~entityType": string;
	"~start": string;
	"~end": string;
	"~type": string;
	"~properties": OCProperties;
};

export type OpenCypherFetch = <TResult = any>(
	queryTemplate: string,
) => Promise<TResult>;
