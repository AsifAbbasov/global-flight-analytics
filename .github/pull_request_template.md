## Change summary

Describe the user-visible, domain, data, infrastructure, or maintenance outcome.

## Risk and invariants

State the concrete risks this change can introduce and the invariants that must remain true.

- Risk:
- Protected invariant:
- Compatibility impact:

## Evidence

Link or describe the code, migration, query plan, benchmark, issue, or runtime evidence that supports the change.

## Verification

List the commands and checks that were actually executed. State explicitly when a relevant check was not run.

```text
<commands and results>
```

## Documentation

- [ ] No documentation change is required.
- [ ] Existing documentation was updated.
- [ ] A new numbered document or amendment was added.

## Reviewer classification

Review findings must follow `docs/82_CODE_REVIEW_STANDARD.md`.

- **Blocker** — demonstrated correctness, integrity, security, migration, concurrency, or required-gate failure.
- **Required change** — material risk that must be fixed or rejected with evidence before merge.
- **Suggestion** — non-blocking improvement.
- **Nit** — minor non-blocking issue.

A Blocker or Required change must include Location, Evidence, Risk, Required change, and Verification.

## Author checklist

- [ ] The change uses the simplest correct design.
- [ ] Invalid and unavailable states remain distinguishable from valid zero values.
- [ ] Important rules have regression tests.
- [ ] Database performance claims include measured evidence when applicable.
- [ ] No review claim relies only on function length, one vocabulary word, `nil`, or a principle label.
- [ ] The reviewed commit and unexecuted checks are explicit.
