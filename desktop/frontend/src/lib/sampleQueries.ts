export interface SampleQuery {
	name: string;
	description: string;
	query: string;
	view: "graph" | "json";
	usesTrial?: boolean;
}

const latest =
	"coalesce(out('has_latest_version'),out('has_version').order().by('createdAt',desc).limit(1))";
// Emit the Study -> StudyVersion path explicitly so the root stays connected in graph view.
const latestPath =
	"coalesce(outE('has_latest_version').inV().path(),outE('has_version').inV().order().by('createdAt',desc).limit(1).path())";

export const SAMPLE_QUERIES: SampleQuery[] = [
	{
		name: "Study Overview",
		description:
			"One study's latest version with design, source, collaborators, epochs, encounters, activities, timelines, and more.",
		query:
			"g.V().hasLabel('Study').limit(1).outE('has_latest_version').inV().union(bothE('has_source_version').otherV().limit(1).union(path(),outE('has_collaborator').inV().limit(1).path()),outE('has_design').inV().limit(1).union(path(),outE('has_epoch').inV().limit(1).path(),outE('has_encounter').inV().limit(1).path(),outE('has_activity').inV().limit(1).path(),outE('has_phase').inV().limit(1).path(),outE('has_study_type').inV().limit(1).path(),outE('has_indication').inV().limit(1).path(),outE('has_therapeutic_area').inV().limit(1).path(),outE('has_model').inV().limit(1).path(),outE('has_population').inV().limit(1).path(),outE('has_timeline').inV().limit(1).union(path(),outE('has_timing').inV().limit(1).path(),outE('has_instance').inV().limit(1).path(),outE('has_role_relationship').inV().limit(1).path())))",
		view: "graph",
	},
	{
		name: "Study + Versions",
		description:
			"A study's version history, including the latest-version relationship (up to 50 versions).",
		query:
			"{{study}}.union(outE('has_version').inV().limit(50).path(),outE('has_latest_version').inV().path())",
		usesTrial: true,
		view: "graph",
	},
	{
		name: "Full Trial Graph",
		description:
			"Bounded overview of the current version: designs, epochs, visits, activities, timelines, and timings. This is a preview, not the complete graph.",
		query: `{{study}}.union(identity().path(),${latestPath},${latest}.outE('has_design').inV().limit(3).union(identity().path(),outE('has_epoch').inV().limit(10).path(),outE('has_encounter').inV().limit(20).path(),outE('has_activity').inV().limit(20).path(),outE('has_phase','has_study_type').inV().path(),outE('has_timeline').inV().limit(3).union(identity().path(),outE('has_timing').inV().limit(10).path(),outE('has_instance').inV().limit(10).path())))`,
		usesTrial: true,
		view: "graph",
	},
	{
		name: "Epochs & Encounters",
		description:
			"Current-version study periods and visits, with up to 20 of each per design.",
		query: `{{study}}.union(identity().path(),${latestPath},${latest}.outE('has_design').inV().limit(3).union(identity().path(),outE('has_epoch').inV().limit(20).path(),outE('has_encounter').inV().limit(20).path()))`,
		usesTrial: true,
		view: "graph",
	},
	{
		name: "Timelines & Timing",
		description:
			"Current-version timelines and their timing relationships, up to 20 timings per timeline.",
		query: `{{study}}.union(identity().path(),${latestPath},${latest}.outE('has_design').inV().limit(3).outE('has_timeline').inV().limit(3).union(identity().path(),outE('has_timing').inV().limit(20).path()))`,
		usesTrial: true,
		view: "graph",
	},
];

export function buildSampleQuery(sample: SampleQuery): string {
	return sample.query.replace(
		"{{study}}",
		() => "g.V().hasLabel('Study').limit(1)",
	);
}
