# Security and Operational Hardening

Status: PRE-MERGE REVIEW CANDIDATE — TECHNICAL BASELINE CI_VERIFIED; FINAL EXACT-HEAD VALIDATION OWNED BY PR/CHECK EVIDENCE  
Date: 2026-09-08  
Repository: `AsifAbbasov/global-flight-analytics`  
Pull request: `#172` — `security: harden API and operational resilience`  
Base branch: `main`  
Baseline SHA: `9e806f2f44c24d69f695ae37ab02a0f5a0df94f3`  
Technical pre-documentation candidate: `46f5ce361f01d016a63ee2077f961e86b702eafb`

---

## 1. Purpose

This document records the bounded security and operational hardening increment implemented in PR #172. The increment strengthens existing production and repository controls without changing the product architecture, adding paid infrastructure, or introducing a new application runtime.

The hardening slice addresses four concrete areas:

1. API transport and logging privacy;
2. repository secret scanning and software-bill-of-materials evidence;
3. bounded Go fuzz testing for security-sensitive normalization and client-identity code;
4. a disposable PostgreSQL backup/restore verification drill.

This document is an evidence and review record. It does not create a new product stage and does not create a synthetic finding ID solely because a hardening package exists.

---

## 2. Explicit technology and scope boundary

The increment preserves the existing project stack and cost constraints.

```text
PYTHON=NONE
REDIS=NONE
VALKEY=NONE
SECOND_OBSERVABILITY_STACK=NONE
NEW_PAID_INFRA=NONE
PRODUCTION_DATABASE_MUTATION=NONE
NEW_DATABASE_TABLE=NONE
NEW_DATABASE_MIGRATION=NONE
```

Python is not added to application source, runtime, CI, security tooling, backup tooling, fuzzing, or documentation-owned execution paths.

Redis/Valkey is intentionally not introduced as a speculative distributed rate-limiting or cache dependency. The existing observability stack is reused rather than duplicated.

---

## 3. API request-log privacy hardening

### Previous risk

Ordinary request logging could retain resolved client IP addresses and raw request paths. A raw path can contain resource identifiers, while query strings can expose user-controlled or operationally sensitive values. Error-path logging had a similar risk if raw request values or underlying error contents were emitted.

### Implemented boundary

`apps/api/internal/middleware/request_logger.go` now records the bounded Fiber route template instead of the raw request target and does not emit the resolved client IP.

`apps/api/internal/server/protection.go` applies the same route-template boundary to server-side 5xx logging and avoids serializing the raw underlying error into the log record.

When no route template is available, the bounded fallback is `unmatched` rather than the raw URL.

### Regression protection

The package includes dedicated tests for:

- ordinary request-log client-IP suppression;
- raw path/query suppression;
- server error-handler privacy;
- panic-recovery privacy.

The intended logging contract is diagnostic usefulness without durable storage of per-request IP addresses or raw identifier-bearing request targets.

---

## 4. HTTP transport bounds and security headers

The Fiber application configuration now explicitly bounds:

```text
ReadBufferSize=4096
Concurrency=4096
```

These values are defensive process limits. They are not a claim of measured maximum production capacity and must not be interpreted as a throughput target.

The API security-header middleware now also emits:

```text
Strict-Transport-Security: max-age=31536000
```

HSTS is meaningful only for HTTPS delivery through the production TLS edge. Local plain-HTTP development does not independently demonstrate browser HSTS behavior.

The existing security-header contract remains intact, including content-type, framing, referrer, permissions, CSP, and cross-domain policy protections.

---

## 5. Repository secret scanning

### 5.1 Selected-Actions governance conflict discovered during implementation

The first candidate attempted to consume third-party Gitleaks/SBOM GitHub Actions. The repository already enforces a stricter Actions policy: GitHub-owned Actions are allowed, external Action use is restricted, and immutable SHA pinning is required.

The first `Security Hardening` workflow therefore failed before any job started. The repository policy was not weakened to make the new workflow pass.

### 5.2 Chosen remediation

Gitleaks is installed as a version- and checksum-pinned binary:

```text
GITLEAKS_VERSION=8.30.0
GITLEAKS_LINUX_X64_SHA256=79a3ab579b53f71efd634f3aaf7e04a0fa0cf206b7ed434638d1547a2470a66e
```

The workflow uses only already-permitted GitHub-owned Actions in `uses:` expressions.

### 5.3 Detector self-test

Before scanning the repository, CI creates a temporary Git repository and constructs a synthetic AWS-shaped access-key identifier from separate safe fragments at runtime. The complete secret-shaped string is not stored in this repository.

The self-test is fail-closed:

- if Gitleaks detects the synthetic finding, the detector test passes;
- if Gitleaks returns a clean result for the synthetic finding, CI fails and the repository scan is not trusted.

GitHub push protection itself rejected an earlier attempt to store the complete synthetic credential-shaped test value in the workflow. The test was redesigned rather than bypassing push protection.

### 5.4 Historical scan and 17 classified findings

A full-history Gitleaks run scanned hundreds of repository commits and reported 17 `generic-api-key` findings. They were not blindly suppressed.

Each finding was traced to its exact `commit:file:rule:line` fingerprint and reviewed in historical context. The 17 findings were classified as non-secret values belonging to these categories:

- generated OpenAPI SHA-256 provenance comments in `packages/api-client/src/generated.ts`;
- SHA-256 digests and synthetic keys used by disposable CI/API-load/container verification;
- local-only Compose test/startup values;
- observability/config test fixtures;
- domain `IdentityKey` test fixtures.

No production credential was identified among these 17 findings. This statement is intentionally limited to these 17 findings and is not a claim that repository history can never contain another credential-shaped value.

### 5.5 Exact fingerprint allowlist

`.gitleaksignore` contains exactly the 17 reviewed fingerprints. There is no wildcard, path-wide suppression, or disabling of the `generic-api-key` rule.

The workflow additionally verifies:

```text
EXPECTED_GITLEAKS_IGNORE_COUNT=17
IGNORE_FILE_DUPLICATES=0
FINGERPRINT_FORMAT=VALID
```

Any future additional finding remains unignored and therefore fails the secret-scan job unless it is independently reviewed and the explicit count/policy is changed in a later pull request.

---

## 6. CycloneDX source SBOM

Syft is installed as a version- and checksum-pinned binary:

```text
SYFT_VERSION=1.51.1
SYFT_LINUX_AMD64_SHA256=8fcb33017a0dc1058298c923c436d19dfa68ae93968e0b423248542e3afb9fc3
```

CI generates:

```text
artifacts/security/gfa-source-sbom.cdx.json
```

The output is validated as CycloneDX JSON and uploaded as a GitHub Actions artifact with 30-day retention.

This is dependency/source inventory evidence. It is not a signed build attestation, SLSA provenance statement, or proof that every runtime package is vulnerability-free.

---

## 7. Go fuzz safety

Two bounded fuzz targets were added using the existing Go toolchain:

1. trusted-proxy/client-identity resolution;
2. traffic-state normalization.

The CI fuzz budget is 15 seconds per target. The tests assert deterministic/bounded behavior and panic resistance over generated input.

This is continuous adversarial-input coverage, not exhaustive formal verification. The bounded CI duration is deliberately selected to keep the control compatible with the project's zero-cost CI constraints.

---

## 8. PostgreSQL backup/restore drill

`scripts/run-postgres-backup-restore-drill.sh` performs a disposable logical-backup verification against `postgres:16.14-alpine3.24`.

The drill:

1. starts an isolated PostgreSQL container;
2. applies the canonical repository migrations through the backend image;
3. inserts a deterministic local fixture;
4. creates a custom-format `pg_dump` backup;
5. computes and verifies SHA-256 evidence;
6. restores into a separate database;
7. verifies critical-table existence;
8. reconciles source and restored row counts;
9. verifies the deterministic fixture in the restored database;
10. proves the source disposable database remained unchanged;
11. uploads only checksum/evidence metadata, not the database dump itself.

Critical tables checked by the current drill include:

```text
airports
ingestion_runs
flight_states
flight_trajectories
```

The drill does not connect to or mutate production Neon. It proves that the repository's canonical schema and a representative local dataset can survive the exercised logical dump/restore path. It does not establish production RPO, RTO, point-in-time recovery, provider-side snapshot policy, or a full Neon disaster-recovery SLA.

---

## 9. Rejected and intermediate validation history

Validation history is preserved rather than rewritten as an uninterrupted success path.

### Candidate `ca921d90877bf1e02e9443e03d88282e9be43de4`

`Security Hardening` run #1 ended in `startup_failure` with zero jobs because newly referenced third-party Actions were rejected by the repository selected-Actions policy.

Disposition: rejected. The allowlist was not weakened.

### Candidate `193b8aec8768a958b6942c9bdd59c349e3bfb4d9`

Third-party Actions were replaced by checksum-pinned binaries. Backup/restore, fuzz, and SBOM jobs passed, but the Gitleaks detector self-test failed because the first synthetic fixture was not matched by the selected rule set.

Disposition: rejected. The detector self-test was made deterministic.

### Candidate `41716cd75c02cea472b70f999e68f6d8894f8197`

The Gitleaks self-test passed. Full-history scanning then correctly surfaced 17 findings.

Disposition: rejected as a clean-candidate SHA. Findings required classification before any allowlist could be accepted.

### Candidate `1c7404366cf55af24a0d500cb7cc30303100b975`

The workflow added redacted finding metadata (`RuleID`, file, commit) without printing secret values. All 17 findings were `generic-api-key`.

Disposition: diagnostic candidate only.

### Candidate `e706e2db823cb9cf4b2bb3f34618fdc8dae2cb10`

The diagnostic was extended to emit Gitleaks fingerprints without secret values, allowing exact finding-level classification.

Disposition: diagnostic candidate only.

### Candidate `edaacbd2e4cc091d799e37683d7bebbd6b132560`

The exact 17 reviewed fingerprints were added to `.gitleaksignore`.

Its hardening workflow was superseded/cancelled after the branch advanced again, so no final validation status is transferred from this SHA.

### Technical baseline `46f5ce361f01d016a63ee2077f961e86b702eafb`

The ignore-scope guard was added. This is the first technical candidate on which the entire existing repository CI matrix, the new hardening workflow, and Vercel all succeeded.

---

## 10. Exact pre-documentation validation evidence

Exact SHA:

```text
46f5ce361f01d016a63ee2077f961e86b702eafb
```

GitHub workflow results:

| Workflow | Run | Result |
|---|---:|---|
| Backend CI | `34169790682` / #896 | SUCCESS |
| Frontend CI | `34169790747` / #558 | SUCCESS |
| OpenAPI Contract | `34169790681` / #187 | SUCCESS |
| API Load Baseline | `34169790699` / #415 | SUCCESS |
| CodeQL | `34169790702` / #539 | SUCCESS |
| Security Hardening | `34169790688` / #7 | SUCCESS |

Backend CI #896 includes successful Go formatting, Go tests, `go vet`, architecture/contract audits, repository policy audits, pinned Go vulnerability analysis, race tests, PostgreSQL 16 integration, backend container build, non-root runtime verification, production historical materializer verification, and container health smoke tests.

Security Hardening #7 includes successful:

```text
Secret Scan
Go Fuzz Safety
CycloneDX SBOM
PostgreSQL Backup Restore Drill
Security Hardening Gate
```

The exact same SHA also has Vercel status `SUCCESS` with description `Deployment has completed`.

This section is intentionally called **pre-documentation evidence**. It is immutable historical evidence and is not transferred to a later review head.

### Final exact-head evidence ownership

The final pre-merge head SHA cannot be embedded into this file without changing that SHA and creating recursive evidence drift. Therefore final mutable pre-merge evidence is deliberately owned by:

1. the PR #172 body, which records the current candidate SHA and exact run IDs;
2. GitHub Actions checks bound to that exact SHA;
3. Vercel commit status bound to that exact SHA;
4. PR mergeability and unresolved-review-thread state read directly from GitHub.

This document owns the stable gate definition and immutable historical baseline; it does not self-embed a final commit identifier.

---

## 11. Relationship to GFA-SEC-445

`GFA-SEC-445` remains an independent canonical finding:

```text
Repository security automation and hosting-side security settings were not comprehensively enforced
```

PR #172 materially improves repository security evidence, including working full-history secret scanning and empirical observation that GitHub push protection rejected a credential-shaped test value.

However, the existing `GFA-SEC-445` closure boundary also includes external hosting-side settings whose current state was previously not independently verified, including Dependabot/Secret Scanning/settings-policy state.

This document therefore does **not** mark `GFA-SEC-445` closed. The canonical finding register must remain unchanged unless that finding's own external-settings verification contract is independently satisfied.

---

## 12. Operational and deployment consequences

- no production database is touched by the restore drill;
- no new paid service is required;
- no Python runtime/tooling is introduced;
- no Redis/Valkey dependency is introduced;
- CI duration increases because fuzzing, SBOM generation, repository-history scanning and backup/restore now execute as hardening evidence;
- SBOM and restore evidence are retained as bounded GitHub Actions artifacts;
- request logs intentionally contain less raw client/request identity data;
- API header/concurrency bounds are now explicit rather than implicit defaults.

---

## 13. Residual risks and limitations

1. The 17 historical Gitleaks ignores are exact reviewed fingerprints, but future generated checksum values can still trigger the generic rule and require explicit review.
2. The logical PostgreSQL drill is not a production Neon disaster-recovery exercise and does not prove RPO/RTO.
3. CycloneDX output is inventory evidence, not a signed provenance or complete runtime attestation.
4. Fifteen-second fuzz budgets increase adversarial input coverage but are not exhaustive.
5. HSTS effectiveness depends on the HTTPS production edge.
6. `ReadBufferSize=4096` and `Concurrency=4096` are bounded policy values, not full production capacity sizing. API Load passing demonstrates no detected regression in the repository baseline, not universal load safety.
7. `GFA-SEC-445` remains independently `IN_PROGRESS` until its hosting-side verification boundary is satisfied.
8. The current hardening PR has not been merged; canonical main remains unchanged.

---

## 14. Review-ready gate

PR #172 may be considered review-ready only when its **current exact head SHA** independently proves:

```text
BACKEND_CI=SUCCESS
FRONTEND_CI=SUCCESS
OPENAPI_CONTRACT=SUCCESS
API_LOAD_BASELINE=SUCCESS
CODEQL=SUCCESS
PLAYWRIGHT_E2E=SUCCESS
SECURITY_HARDENING=SUCCESS
VERCEL=SUCCESS
DOCUMENT_INDEX_ALIGNMENT=PASS
PR_BODY_ALIGNMENT=PASS
MERGEABLE=TRUE
UNRESOLVED_REVIEW_BLOCKERS=0
```

No status from `46f5ce361f01d016a63ee2077f961e86b702eafb` or any other earlier SHA may be transferred to the current head.

Because embedding the final head SHA in this file would itself change the head SHA, satisfaction of this gate is recorded externally in the PR body and exact commit checks rather than by another self-referential documentation commit.

---

## 15. Merge gate

```text
MERGE_AUTHORIZATION=NOT_GRANTED
```

The pull request must not be merged until the user explicitly authorizes the exact current head SHA. If the head moves after authorization, the authorization is invalid and a new exact-SHA authorization is required.

---

## 16. Post-merge closure gate

A successful PR merge alone is not canonical closure.

After an explicitly authorized merge, closure requires:

1. the exact canonical `main` merge SHA;
2. independent post-merge GitHub CI evaluation for all applicable workflows;
3. independent Vercel status evaluation for that main SHA;
4. documentation/registry truth reconciliation;
5. confirmation that no rejected pre-merge result has been reused as post-merge evidence.

Only after those conditions are satisfied may this hardening increment be described as canonically closed.

---

## 17. Current disposition

```text
SECURITY_OPERATIONAL_HARDENING=PRE_MERGE_REVIEW_CANDIDATE
TECHNICAL_BASELINE_SHA=46f5ce361f01d016a63ee2077f961e86b702eafb
TECHNICAL_BASELINE_GITHUB_CI=6_OF_6_SUCCESS
TECHNICAL_BASELINE_SECURITY_HARDENING=SUCCESS
TECHNICAL_BASELINE_VERCEL=SUCCESS
FINAL_EXACT_HEAD_VALIDATION=OWNED_BY_PR_AND_COMMIT_CHECKS
DOCUMENT_INDEX_ALIGNMENT=PASS
GFA_SEC_445=IN_PROGRESS_INDEPENDENT_BOUNDARY
PYTHON=NONE
REDIS=NONE
VALKEY=NONE
MERGE_AUTHORIZATION=NOT_GRANTED
```