# Strata Weave

Strata Weave is a field-record service for archaeological excavations. It keeps excavation areas, stratigraphic units, Harris relationships, finds, samples, laboratory results, and review workflows in a durable SQLite database.

## Run

```text
go run ./cmd/strataweave -db strata.db -addr :8080
```

The service exposes JSON APIs under `/api` and serves a small operational page at `/`.

## Domain rules

- A stratigraphic unit can only be linked to an earlier or later unit once both units belong to the same trench.
- A field record moves from draft to submitted to reviewed or rejected.
- A sample cannot be sent to a laboratory before its parent find is reviewed.
- Relationship insertion rejects cycles in the Harris matrix.
- Batch telemetry imports are cancellable and preserve per-row errors.
