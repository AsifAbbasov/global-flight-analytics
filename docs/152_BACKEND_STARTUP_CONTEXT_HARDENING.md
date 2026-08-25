# Backend Startup Context Hardening

Status: CLOSED
Project: Global Flight Analytics
Reviewed baseline: `cd99532aa40fe7a61fb19580455ac2d1ba5f650c`
Implementation commit: `fa222650ef5425dcf60a05fb949c17cafc36c1a8`
Backend CI: `30708959553` — SUCCESS
Date: 2026-08-01
Canonical reconciliation: 2026-08-25
Canonical finding: `GFA-OPS-438`

## 1. Finding / symptom

The API server's PostgreSQL startup path did not inherit the server's signal-owned
lifecycle context.

Before remediation, `run` owned a context cancelled by the server process lifecycle,
but PostgreSQL pool construction created its connection timeout from a new
`context.Background()`. A `SIGINT` or `SIGTERM` received while PostgreSQL pool creation
or the startup ping was still in progress therefore could not immediately cancel that
work.

Canonical finding:

```text
GFA-OPS-438
API PostgreSQL startup was detached from the signal-owned lifecycle context
```

## 2. Root cause

The database helper mixed two separate responsibilities:

1. choosing a bounded PostgreSQL connection deadline; and
2. choosing the root context from which that deadline was derived.

The timeout itself was already bounded, but the helper unconditionally created it from
`context.Background()`:

```text
context.Background()
→ context.WithTimeout(...)
→ pgxpool.New
→ pool.Ping
```

That made the database helper the owner of process-lifecycle cancellation even when the
API server already had a more authoritative signal-owned context.

This is distinct from `GFA-OPS-436`. Finding 436 covers the repository-wide class in
which reusable functions or methods accepted a caller context and then directly
replaced that caller-owned parameter. This finding concerns a composition/startup path
that created a new root context before deriving a timeout. The direct assignment AST
rule from Document 147 does not by itself prove the absence of every such detached root.

## 3. Failure scenario

A representative failure sequence was:

1. API process begins startup;
2. the server lifecycle context is active and is configured to be cancelled on
   `SIGINT` / `SIGTERM`;
3. PostgreSQL pool creation begins with its own timeout derived from
   `context.Background()`;
4. the process receives a termination signal while PostgreSQL initialization is blocked
   or slow;
5. the server lifecycle context is cancelled;
6. PostgreSQL startup work does not observe that cancellation because it belongs to a
   detached root;
7. shutdown/startup termination must wait for the independent database timeout or some
   other PostgreSQL completion condition.

The bug did not require an infinite database wait. The bounded timeout limited the
maximum duration, but bounded delay is not equivalent to correct lifecycle ownership.

## 4. Impact

The defect affected operational lifecycle correctness rather than analytical data
semantics.

Potential effects included:

- delayed response to deployment shutdown signals during API startup;
- slower container termination or restart when PostgreSQL was unreachable or slow;
- divergence between the server's declared lifecycle and work performed on its behalf;
- less predictable cancellation/error behavior during startup incidents.

There is no evidence that this defect corrupted persisted aviation data or produced
incorrect analytical values.

## 5. Severity rationale

Severity: **P2 retrospective**.

P2 is appropriate because process cancellation ownership is a real reliability and
operability contract, and the defect was reachable in the API server startup path.
However, the connection operation already had a finite configured timeout, so the
failure was bounded rather than an indefinite hang, and no source-backed evidence shows
data corruption, security compromise or safety-critical output.

The severity is a retrospective engineering classification. It is not represented as a
historical severity label from the original implementation unless separately recorded in
repository evidence.

## 6. Existing guarantees violated

The defect violated the repository's context and lifecycle principles:

- caller/process lifecycle cancellation should propagate through delegated work;
- a child timeout should narrow an existing caller deadline rather than silently replace
  its ownership root;
- server startup and shutdown behavior should remain governed by the signal-owned server
  context;
- cancellation causes should remain observable through wrapped errors.

The defect was also conceptually adjacent to the caller-context ownership audit in
Document 147, while remaining technically outside the narrow direct-assignment class of
that audit.

## 7. Considered solutions

The remediation space included:

1. keep `NewPostgresPool` unchanged and accept independent startup cancellation;
2. change the existing `NewPostgresPool` signature repository-wide to require a
   context;
3. add a context-aware constructor and migrate only the API server composition path;
4. remove the configured timeout and rely only on the server lifecycle context;
5. add a second goroutine/cancellation bridge around the existing helper.

## 8. Chosen remediation

The implementation introduced:

```text
NewPostgresPoolContext(ctx, databaseURL, connectTimeout)
```

The context-aware function:

- rejects a nil caller context;
- rejects a non-positive connection timeout with explicit classification;
- derives `connectCtx` with `context.WithTimeout(ctx, connectTimeout)`;
- uses that same bounded child for `pgxpool.New` and startup `Ping`;
- preserves caller cancellation through the returned error chain.

The API server now passes its signal-owned context through:

```text
SIGINT / SIGTERM
→ run(ctx, ...)
→ openServerDatabase(ctx, ...)
→ NewPostgresPoolContext(ctx, ...)
→ context.WithTimeout(ctx, configured timeout)
→ pgxpool.New
→ pool.Ping
```

The existing `NewPostgresPool` remains as a compatibility wrapper for existing commands
and verification tools and delegates to `NewPostgresPoolContext(context.Background(),
...)`.

## 9. Why this solution was selected

The selected solution fixes the production server lifecycle defect while keeping the
change narrow.

It provides the API composition root with correct context ownership without forcing an
unrelated repository-wide signature migration across commands whose lifecycle policies
were not part of this finding. It also keeps the database package responsible for the
connection timeout while allowing the caller to own the broader lifecycle.

The resulting composition expresses both limits explicitly:

```text
caller/process cancellation
AND
configured PostgreSQL startup deadline
```

Whichever occurs first terminates the bounded operation.

## 10. Rejected alternatives

### Keep the detached background root

Rejected because a configured timeout does not justify ignoring an earlier caller or
process cancellation signal.

### Change every `NewPostgresPool` caller in the same patch

Rejected as unnecessary scope expansion. Existing commands and verification tools had
separate lifecycle ownership and were not required to close the API-server defect.

### Remove the PostgreSQL timeout

Rejected because caller cancellation and maximum connection duration are independent
controls. A server that receives no shutdown signal still requires a bounded startup
attempt.

### Add a cancellation goroutine around the old helper

Rejected because the correct ownership can be expressed directly through Go's context
tree without an additional goroutine, channel or cancellation bridge.

## 11. Trade-offs

The compatibility wrapper intentionally remains able to create a root background
context. This means the repository now exposes both:

- a context-aware constructor for callers with lifecycle ownership; and
- a compatibility constructor for callers that deliberately do not supply one.

That dual API is less minimal than changing the signature everywhere, but it avoided
turning a targeted server correction into a broad migration.

The trade-off is explicit: future production composition paths that already own a
lifecycle context should use the context-aware API. The existence of the legacy wrapper
must not be interpreted as permission for the API server to detach from its lifecycle.

## 12. Regression tests / protection

The implementation commit added focused coverage proving:

1. `NewPostgresPoolContext` rejects a nil caller context;
2. non-positive connection timeouts preserve explicit error classification;
3. an already cancelled caller context remains observable as `context.Canceled`;
4. database-optional server startup remains valid;
5. `openServerDatabase` rejects a missing server context;
6. server startup preserves the PostgreSQL cancellation cause while retaining
   `SERVER_DATABASE_CONNECTION_FAILED` classification.

The repository's broader Go tests, `go vet`, architecture/contract audit, code-review
policy audit and Backend Context Ownership audit also remained in Backend Quality.

## 13. Adversarial review findings

The canonical review identified two important distinctions:

- the old connection timeout meant this was not an unbounded-hang finding; the failure
  was specifically detached lifecycle cancellation;
- the permanent caller-context AST audit from Document 147 is not a complete detector
  for helper-mediated or composition-root `context.Background()` creation. Therefore
  this finding is not collapsed into `GFA-OPS-436` and its prevention claim is not
  overstated.

No evidence supports upgrading the issue into a database integrity or analytical
correctness finding.

## 14. Remediation iterations

The source-backed engineering remediation is concentrated in:

```text
fa222650ef5425dcf60a05fb949c17cafc36c1a8
fix: propagate server startup context to postgres
```

That commit introduced the context-aware database constructor, server propagation,
explicit error classifications, focused tests, this documentation increment and its
index entry.

No later corrective implementation commit is required to establish closure of this
specific defect from the recovered repository evidence.

## 15. Residual risks / limitations

The remediation does not prove that every `context.Background()` root in the repository
is wrong or eliminated. Independent process roots, explicit compatibility wrappers,
detached cleanup with governed ownership and other composition roots can be legitimate.

The direct-assignment AST audit from Document 147 remains narrower than a full
interprocedural context-lifecycle analysis. Reviewers still need to examine newly
introduced background roots semantically.

The legacy `NewPostgresPool` wrapper remains intentionally available. If a future
long-lived production path uses it despite already owning cancellation, that usage must
be reviewed independently.

## 16. Operational / deployment consequences

No migration, schema change, new service, dependency or infrastructure component was
required.

Operationally, a server shutdown signal received during PostgreSQL initialization can
now cancel the API's startup database work immediately, subject to normal Go/pgx
cancellation behavior, rather than waiting only on the independently rooted connection
timeout.

Database-optional startup behavior remains unchanged.

## 17. Exact evidence

Reviewed baseline:

```text
cd99532aa40fe7a61fb19580455ac2d1ba5f650c
```

Implementation:

```text
fa222650ef5425dcf60a05fb949c17cafc36c1a8
fix: propagate server startup context to postgres
```

The implementation diff shows:

- `run` passing its `ctx` into `openServerDatabase`;
- `openServerDatabase` requiring a non-nil context;
- the API server calling `database.NewPostgresPoolContext`;
- `NewPostgresPoolContext` deriving its timeout from the caller context;
- `pgxpool.New` and `pool.Ping` using the derived child context;
- focused cancellation and classification tests.

Historical exact-commit GitHub Actions evidence:

```text
Backend CI 30708959553 = SUCCESS
```

The run's required jobs include successful Backend Quality and Backend Race Safety;
Backend Quality successfully executed Go tests, Go vet, architecture/contract review,
code-review policy and the Backend Context Ownership audit on the implementation SHA.
The run also records the remaining required backend jobs as successful in GitHub
Actions.

## 18. Final canonical status

```text
Finding: GFA-OPS-438
Severity: P2 retrospective
Status: CLOSED
Implementation: fa222650ef5425dcf60a05fb949c17cafc36c1a8
Historical Backend CI: 30708959553 = SUCCESS
Open findings owned by this document: 0
```

The old document header `Implementation prepared; exact-commit Continuous Integration
closure pending` is superseded by the recovered exact-commit Actions evidence.

## 19. Prevention / future guard

Future API/server composition code that already owns a lifecycle context must derive
bounded PostgreSQL startup work from that context rather than create a detached root.

Protection consists of:

- the context-aware constructor and focused tests;
- server startup cancellation tests;
- Backend Quality Go tests and `go vet`;
- the repository context-ownership and review-policy gates;
- code-review scrutiny of newly introduced root contexts that fall outside the narrow
  direct-assignment AST detector.

A future proposal to remove the compatibility wrapper or migrate every command should be
handled as its own evidence-based cleanup. It is not required to keep this production
server finding closed.
