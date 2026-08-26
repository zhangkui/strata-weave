# Strata Weave design

The service models the daily work of a small archaeological field team. Excavators register trenches and stratigraphic units, then attach finds, samples, photographs and observations. A recorder submits a field record for review. Reviewers can reject a record with a reason or approve it, which unlocks sample dispatch and dating results.

The Harris matrix is represented as directed `earlier-than` edges. Validation runs across the graph, not only on a single row, so a cycle is rejected before it reaches storage. An import worker accepts batches of instrument observations, checks cancellation between rows, and writes accepted observations in one transaction.

The HTTP layer is intentionally thin. The service layer owns state transitions, permissions encoded as workflow roles, date-window calculations, graph traversal and aggregation. SQLite is opened as a file database and schema migrations are applied on startup.

The first release includes dashboards for field operations, but all visible data is sourced from the Go service. The web page is only an operator view and is not the subject of bugs.
