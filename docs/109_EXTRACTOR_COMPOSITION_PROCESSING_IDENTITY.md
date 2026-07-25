# Document 109 — Extractor Composition Processing Identity

Status: IMPLEMENTED
Baseline commit: abd038c10d1d382843dbaefb8b506efeff5fdeda

## 1. Problem

The durable snapshot key already contains `processing_version`, but the extractor
input fingerprint previously represented only the trajectory. Changing
geographic cell precision, component versions, aircraft cache policy, the
not-found classifier or resolved aircraft metadata could therefore reuse the
same idempotent record incorrectly.

## 2. Contract

The production extractor composition now creates a deterministic processing
identity from:

- every extractor component version;
- effective geographic cell precision;
- effective positive and negative aircraft cache durations;
- an explicit aircraft not-found policy version.

The identity is included in the extractor fingerprint input. Resolved aircraft
features are also included, so mutable lookup results cannot silently replay an
older record.

Custom not-found classifiers require an explicit stable policy version.
Typed-nil aircraft lookup dependencies are rejected. The unused `NewExtractor`
shortcut is removed so production callers retain the complete composition and
its identity evidence.

## 3. Generation boundary

The extractor, extractor composition and processing pipeline generations first
advanced to version 2 in this increment. Later temporal-safety corrections use
version 3. Version 1 and version 2 snapshots remain readable through their
explicit stored processing versions and are not overwritten.

## 4. Permanent evidence

`featureprocessingidentityaudit` requires the identity model, effective policy
inputs, output-sensitive fingerprint, regression tests and absence of
`NewExtractor`.

```text
EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY=ENFORCED
GEOGRAPHIC_PRECISION_FINGERPRINT_COLLISION=CLOSED
AIRCRAFT_METADATA_FINGERPRINT_INPUT=ENFORCED
CUSTOM_NOT_FOUND_POLICY_IDENTITY=ENFORCED
TYPED_NIL_AIRCRAFT_LOOKUP_REJECTED=ENFORCED
```
