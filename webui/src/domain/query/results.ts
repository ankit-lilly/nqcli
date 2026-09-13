import type {
	EdgeId,
	EdgeType,
	EntityProperties,
	ScalarValue,
	VertexId,
	VertexType,
} from "../graph";

export type ResultScalar = {
	entityType: "scalar";
	name?: string;
	value: ScalarValue;
};

export type ResultVertex = {
	entityType: "vertex";
	id: VertexId;
	name?: string;
	types: VertexType[];
	attributes?: EntityProperties;
};

export type PatchedResultVertex = Omit<
	ResultVertex,
	"entityType" | "attributes"
> & {
	entityType: "patched-vertex";
	attributes: EntityProperties;
};

export type ResultEdge = {
	entityType: "edge";
	id: EdgeId;
	name?: string;
	type: EdgeType;
	sourceId: VertexId;
	targetId: VertexId;
	attributes?: EntityProperties;
};

export type PatchedResultEdge = Omit<
	ResultEdge,
	"sourceId" | "targetId" | "entityType" | "attributes"
> & {
	entityType: "patched-edge";
	attributes: EntityProperties;
	source: PatchedResultVertex;
	target: PatchedResultVertex;
};

export type ResultBundle = {
	entityType: "bundle";
	name?: string;
	values: ResultEntity[];
};

export type PatchedResultBundle = Omit<ResultBundle, "values"> & {
	values: PatchedResultEntity[];
};

export type ResultEntity =
	| ResultVertex
	| ResultEdge
	| ResultScalar
	| ResultBundle;

export type PatchedResultEntity =
	| PatchedResultVertex
	| PatchedResultEdge
	| ResultScalar
	| PatchedResultBundle;

export type QueryResult = {
	results: ResultEntity[];
	rawResponse: unknown;
};
