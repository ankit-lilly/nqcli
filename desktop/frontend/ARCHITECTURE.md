# Desktop Frontend Architecture

The desktop frontend follows the same inward dependency direction as the Web
UI:

```text
Solid presentation -> bootstrap -> Wails adapter -> application -> domain
                              \
                               -> @nq/graph-surface -> Cytoscape
```

- `src/domain` owns graph values and pure merge/removal rules.
- `src/application` defines graph use cases and the query port.
- `src/adapters` translates the generated Wails contract into application
  values.
- `src/bootstrap` composes adapters and use cases.
- `src/components` owns Solid lifecycle and visual state only.
- `ui/graph-surface` owns framework-neutral Cytoscape lifecycle, selection,
  viewport commands, and normalized pointer events shared with the React UI.

Database query construction remains in `internal/desktop`. The Solid UI sends
typed expansion intent, never Gremlin or openCypher fragments.
