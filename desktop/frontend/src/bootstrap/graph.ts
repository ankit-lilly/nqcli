import { WailsGraphQueryAdapter } from "../adapters/wails-graph-query";
import { WailsSchemaAdapter } from "../adapters/wails-schema";
import { GraphExplorer } from "../application/graph-explorer";
import { SchemaExplorer } from "../application/schema-explorer";

export const graphExplorer = new GraphExplorer(new WailsGraphQueryAdapter());
export const schemaExplorer = new SchemaExplorer(new WailsSchemaAdapter());
