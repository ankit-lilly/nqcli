# Web UI Architecture

The web UI uses ports and adapters so graph behavior is independent of React,
Jotai, TanStack Query, Cytoscape, and the nq HTTP transport.

## Dependency Direction

Dependencies point inward:

```text
React presentation -> bootstrap -> adapters -> application -> domain
                                      |
                                      +-> connector compatibility layer
```

- `src/domain` contains graph, schema, connection, profile, and query values. It
  has no browser or framework dependencies.
- `src/application` defines ports and coordinates use cases. It depends only on
  the domain.
- `src/adapters` implements ports for the nq API, SSE, Jotai state, and
  Cytoscape.
- `src/bootstrap` is the composition root. A presentation framework asks it for
  an application service rather than constructing transport adapters itself.
- Existing `components`, `hooks`, `modules`, and `routes` are the React
  presentation layer.

`bun run typecheck` runs `scripts/check-boundaries.ts` before TypeScript. The
check rejects imports that point against these rules.

## Adding Behavior

1. Put durable data rules and value types in `domain`.
2. Express external needs as a focused port in `application/ports`.
3. Coordinate the behavior in an application use case.
4. Implement the port under `adapters`.
5. Wire the adapter in `bootstrap` and consume the service from the UI.

Transport payload validation belongs in an adapter. React hooks should manage
rendering lifecycle and query-cache integration, not graph or transport rules.

## Framework Migration

A Solid or server-rendered shell can use `createApplicationForConnection` and
the application services without importing React. Query and schema refresh
behavior can therefore move screen by screen while the existing UI remains
operational.

`CytoscapeSurfaceAdapter` owns Cytoscape creation, plugin registration, and
lifecycle. Its `cytoscape` property is a transitional escape hatch for legacy
React hooks that still apply styles and interaction behavior directly. Move
those capabilities behind `GraphSurfacePort` before removing the React graph
shell; no all-at-once rewrite is required.

The lifecycle, selection, viewport commands, and pointer event normalization
live in the local `@nq/graph-surface` package under `ui/graph-surface`. Both the
React Web UI adapter and the Solid desktop host use this package. Framework
hosts own component state; transport adapters own HTTP or Wails calls; neither
layer owns Cytoscape's lifecycle.
