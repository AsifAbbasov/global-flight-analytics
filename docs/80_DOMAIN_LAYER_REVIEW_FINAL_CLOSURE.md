# Domain Layer Review Final Closure

## Status

The accepted Domain Layer review findings are closed.

The final closure consists of executable contracts, not only documentation:

- service constructors return controlled errors;
- explicitly named `MustNewService` helpers are retained only for composition roots;
- repository results are validated before they leave Aircraft, Airport, Flight, and Flight State services;
- altitude semantics are owned by an atomic `Altitude` value object;
- aircraft category availability is owned by an `AircraftCategory` value object;
- reconciliation deduplication keys validate their input internally;
- Provider Health policy thresholds use the `BasisPoints` type;
- regression and compile-time closure tests guard these decisions.

## Public record structures

Airport, Flight, Flight State, Trajectory, Weather, Ingestion Run, and Data Quality remain exported record structures intentionally.

They are used by PostgreSQL scanning, provider mapping, transport mapping, analytical fixtures, and batch processing. Converting every field to private storage would be a repository-wide breaking migration with little additional correctness after boundary validation was introduced.

The invariant policy is therefore:

1. constructors and value objects own coupled values;
2. every domain record owns `Validate`;
3. services validate repository output;
4. repositories and integrations validate input before persistence or publication;
5. tests enforce these boundaries.

Direct record literals remain available for scanners, adapters, and tests, but they are not considered trusted domain evidence until validation succeeds.

## Deliberately rejected mechanical recommendations

The following review suggestions are not project requirements:

- replacing every optional `*time.Time` with a custom optional wrapper;
- prohibiting the words `With` and `And`;
- splitting functions solely because they exceed fifty lines;
- deleting service boundaries that now normalize and validate data.

These decisions are explicit architectural choices, not unclassified debt.

## Go module security closure

The final verification gate identified `GO-2026-5970` in `golang.org/x/text v0.29.0`.

The module requirement is upgraded to the minimum fixed version, `v0.39.0`. The project uses Go 1.26.5, which satisfies the dependency's Go 1.25 minimum.

The installer treats `go.mod` and `go.sum` as one transactional unit and requires all of the following before closure:

1. `go mod tidy` completes;
2. `go mod verify` succeeds;
3. `go list -m` resolves `golang.org/x/text` to `v0.39.0`;
4. pinned `govulncheck` reports no reachable vulnerability;
5. the established Stage 14 verifier repeats the security gate.

## Frontend image dependency security closure

The final Stage 14 frontend audit identified `GHSA-f88m-g3jw-g9cj` in the
transitive `Next.js -> sharp -> libvips` dependency path. All `sharp` versions
below `0.35.0` are rejected. The workspace now pins `sharp` to `0.35.3`, the
current fixed release used by this closure.

The closure uses two complementary controls:

1. `apps/web/package.json` directly pins `sharp` to `0.35.3` so the runtime
   image processor is an explicit application dependency;
2. `pnpm-workspace.yaml` overrides every `sharp@<0.35.0` resolution to
   `0.35.3`, including the `sharp ^0.34.5` optional dependency declared by
   Next.js 16.2.9.

The permanent frontend policy reads the workspace manifest, web package
manifest, and lockfile. It rejects vulnerable Sharp versions, missing direct
pins, missing overrides, and a Next.js resolution that does not use the pinned
version. Closure additionally requires:

- `pnpm install --frozen-lockfile`;
- the dependency policy test suite;
- `pnpm audit --prod --audit-level moderate`;
- a successful runtime load reporting `sharp 0.35.3`;
- frontend lint, TypeScript validation, and a production Next.js build;
- the complete Stage 14 verifier repeating the frontend security gate.

## Deterministic frontend font build closure

The Stage 14 production build exposed a network-dependent `next/font/google`
fetch for Geist and Geist Mono. The application no longer downloads fonts from
`fonts.gstatic.com` during compilation. The root layout uses no remote font
loader, while Tailwind receives deterministic system font stacks from
`globals.css`.

The permanent frontend policy rejects `next/font/google`, remote font URLs,
legacy Geist variables, or missing deterministic system font stacks. This keeps
local, continuous integration, and deployment builds reproducible when external
font services are unavailable.
