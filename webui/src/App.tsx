import { Route, Routes } from "react-router";

import Redirect from "./components/Redirect";
import DataExplorer from "./routes/DataExplorer";
import DefaultLayout from "./routes/DefaultLayout";
import GraphExplorer from "./routes/GraphExplorer";
import SchemaExplorer from "./routes/SchemaExplorer";

export default function App() {
	return (
		<Routes>
			<Route element={<DefaultLayout />}>
				<Route path="/data-explorer" element={<DataExplorer />} />
				<Route path="/data-explorer/:vertexType" element={<DataExplorer />} />
				<Route path="/graph-explorer" element={<GraphExplorer />} />
				<Route path="/schema-explorer" element={<SchemaExplorer />} />
				<Route path="*" element={<Redirect to="/graph-explorer" />} />
			</Route>
		</Routes>
	);
}
