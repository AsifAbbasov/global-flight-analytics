# Document 99 — Provenance and Analytical Trust Hardening

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `0ae85ccbff7584a993030c0adcdee3290dd4b7bd`

## 1. Purpose

This increment closes the accepted Analytical Core findings concerning
incomplete source provenance, placeholder source attribution, public failure
detail leakage, and overstated confidence for request-parameter snapshot
metrics.

## 2. Strict source provenance

Every published analytical source must now contain:

```text
a non-placeholder source name;
a known source role;
a complete observation window when an observation window is present;
a non-zero retrieval timestamp;
a retrieval timestamp that does not follow result calculation time.
```

The reserved placeholder name `unknown` is rejected by the analytical source
contract.

Trajectory metadata no longer converts a blank provider name into `unknown`.
Unattributed observations are omitted from the source list and disclosed by the
`unattributed_source_observations` limitation.

All trajectory and request-parameter sources receive a server-owned UTC
`RetrievedAt` timestamp.

## 3. Safe failure publication

The default metric failure path preserves the typed machine code and retry
classification but replaces the operation error text with the stable public
message:

```text
Analytical metric calculation failed.
```

An explicitly supplied custom FailureMapper remains responsible for its own
public contract.

## 4. Honest confidence for request snapshots

Coverage Score and Data Freshness remain deterministic formulas, but their
current public endpoints accept snapshot values supplied by the caller.

Their confidence factors now distinguish:

```text
high formula stability;
low independently verified source coverage.
```

This produces low confidence rather than high confidence for direct
request-parameter calculations. The existing request-parameter limitation
continues to make endpoint results limited rather than complete.

## 5. Verification

The installer performs all changes first in a detached temporary Git worktree
and runs:

```text
complete backend compilation;
targeted tests;
the complete backend test suite.
```

Only after that preflight passes is the working `main` branch changed. The
working tree then runs targeted compilation, targeted tests, race tests, the
complete backend test suite, Go vet, architecture audits, static contract
checks, documentation checks, and whitespace validation.

## 6. Remaining Analytical Core review scope

```text
replace public request-parameter snapshots with server-owned production snapshots;
reject zero analytical reference time;
canonicalize accepted UUID identifiers;
classify and consolidate obsolete analytical foundation packages;
unify metric identifiers across the analytical stack.
```

---

## Canonical remediation history

### GFA-DATA-104 / AC-09 — published analytical provenance was incomplete or placeholder-based

1. **Finding / symptom.** Analytical results could publish incomplete source metadata and could convert a blank provider source into the placeholder name `unknown`.
2. **Root cause.** Publication metadata treated provenance fields as descriptive decoration rather than a validated evidence contract with required source name, role, observation bounds and retrieval time.
3. **Failure scenario.** A result contains a source entry with fabricated `unknown` attribution, missing retrieval time, inconsistent observation bounds or a retrieval timestamp later than calculation; consumers cannot tell which evidence actually supports the metric.
4. **Impact.** Analytical outputs become harder to audit, compare and trust, and missing source ownership can be disguised as positive provenance rather than disclosed as missing evidence.
5. **Severity rationale.** **P1 retrospective.** Provenance is part of the analytical correctness contract: publishing fabricated or structurally incomplete evidence can misrepresent the basis of a metric even when the numeric formula is correct.
6. **Existing guarantees violated.** Source names must identify real evidence owners; roles must be known; temporal evidence must be internally consistent; unavailable attribution must be disclosed as a limitation rather than invented.
7. **Considered solutions.** Allow optional provenance; retain `unknown` as a conventional placeholder; require complete validated analytical sources; drop all sources when any field is missing.
8. **Chosen remediation.** Analytical source validation requires a non-placeholder name, known role, paired observation bounds, non-zero retrieval time and retrieval no later than calculation. Blank trajectory provider names are omitted from sources and counted under `unattributed_source_observations`; server-owned UTC retrieval timestamps are attached to derived evidence.
9. **Why this solution was selected.** It preserves truthful partial evidence without fabricating attribution and turns provenance into an enforceable contract instead of a best-effort label.
10. **Rejected alternatives.** `unknown` creates false source evidence; making all provenance optional defeats auditability; dropping every source because one trajectory is unattributed loses valid evidence from other providers.
11. **Trade-offs.** Some results expose fewer named sources and additional limitations when upstream attribution is missing. This is intentional and more honest than placeholder provenance.
12. **Regression tests / protection.** Source validation tests cover placeholder rejection, retrieval timestamps and observation windows; trajectory metadata tests preserve unattributed-source limitations; the Analytical Core final audit requires strict provenance ownership.
13. **Adversarial review findings.** A non-empty string alone is not sufficient provenance; reserved placeholders must be rejected explicitly. Observation start/end are a pair, and retrieval time must be evaluated relative to calculation time.
14. **Remediation iterations.** The initial hardening made request-parameter evidence explicitly low-confidence; Document 101 later removed caller-owned production snapshots entirely while preserving the stricter source contract.
15. **Residual risks and limitations.** Provenance accuracy still depends on upstream provider/source names being truthful. The analytical layer validates structure and refuses invented placeholders; it does not cryptographically attest external data origin.
16. **Operational or deployment consequences.** No infrastructure change. Consumers may observe new `unattributed_source_observations` limitations where earlier outputs would have shown `unknown`.
17. **Exact evidence.** Historical implementation commit `a31fd8ce3fb6f42a9c90a5153f902c37e7b0f111` (`fix: harden analytical provenance and trust`). Original review ID: `AC-09`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-104=CLOSED`.
19. **Prevention / future guard.** Every new analytical source type must pass the shared provenance validator; placeholder names and structurally incomplete temporal evidence may not be reintroduced as successful publication metadata.

### GFA-SEC-105 / AC-18 — public analytical failures could expose raw operation error text

1. **Finding / symptom.** The default analytical metric failure publication path could include the underlying operation error text in a public response.
2. **Root cause.** Internal diagnostic errors and the external stable error contract shared the same message field instead of being separated by a failure-mapping boundary.
3. **Failure scenario.** A repository, parser, calculation or downstream dependency returns an error containing internal implementation details, identifiers or environment-sensitive text; the default analytical HTTP response publishes that text to an unauthenticated consumer.
4. **Impact.** Internal details can leak through the public API, error contracts become unstable across implementation changes, and clients can become coupled to diagnostic wording.
5. **Severity rationale.** **P1 retrospective.** This is an information-disclosure defect on a public response path, even though the exact sensitivity of any one operation error varies.
6. **Existing guarantees violated.** Public failures must use stable machine codes and safe messages; internal diagnostics must not cross the HTTP trust boundary by default; retry classification must remain machine-readable without exposing raw text.
7. **Considered solutions.** Publish raw errors for debuggability; redact selected substrings; replace the default public message while preserving typed code/retry metadata; suppress all failure detail including machine classification.
8. **Chosen remediation.** The default failure mapper preserves typed code and retry semantics but emits the stable message `Analytical metric calculation failed.`. Explicit custom `FailureMapper` implementations remain responsible for their declared public contract.
9. **Why this solution was selected.** It separates diagnostic and public concerns without removing structured failure semantics needed by clients and tests.
10. **Rejected alternatives.** String redaction is incomplete and brittle; raw errors are not a stable API; deleting machine codes would reduce actionable client behavior unnecessarily.
11. **Trade-offs.** Public responses contain less debugging detail. Operators must use internal logs/telemetry for root-cause diagnostics, which is the correct trust boundary.
12. **Regression tests / protection.** Failure-publication tests require the stable message while preserving error code and retry classification; the Analytical Core final audit checks sanitized failure ownership.
13. **Adversarial review findings.** The safe default must apply even to unexpected operation error types; a custom mapper is an explicit exception and therefore owns its own public-message review instead of silently inheriting raw text.
14. **Remediation iterations.** Sanitization was implemented at the generic metric failure publication seam so all default analytical operations receive the same protection.
15. **Residual risks and limitations.** Deliberately supplied custom failure mappers can publish different messages and must be reviewed independently. Separate server logging redaction is owned by the Server/HTTP findings in Document 94.
16. **Operational or deployment consequences.** No deployment change. External clients should use machine codes rather than internal error strings for control flow.
17. **Exact evidence.** Historical implementation commit `a31fd8ce3fb6f42a9c90a5153f902c37e7b0f111`. Original review ID: `AC-18`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-SEC-105=CLOSED`.
19. **Prevention / future guard.** New public analytical failure paths must pass through a stable mapper; raw `error.Error()` text may not be used as a default client-facing message.

### AC-08 interim mitigation ownership

Document 99 reduced confidence for caller-supplied Coverage Score and Data Freshness snapshots, but this was intentionally an interim trust correction rather than final remediation. The canonical repository finding for `AC-08` is assigned in Document 101, where production endpoints stop accepting caller-owned snapshot evidence and derive the metrics from retained server data. No duplicate GFA finding ID is created here.