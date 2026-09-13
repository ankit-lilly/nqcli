import { WailsGraphQueryAdapter } from "../adapters/wails-graph-query";
import { GraphExplorer } from "../application/graph-explorer";

export const graphExplorer = new GraphExplorer(new WailsGraphQueryAdapter());
