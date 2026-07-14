// Package featurestore defines immutable storage semantics for validated
// FlightFeatures snapshots.
//
// A snapshot is uniquely identified by trajectory ID, schema version and
// as-of time. Repeating a write with the same input fingerprint is
// idempotent. Reusing the same snapshot key with different evidence is a
// conflict, because silently replacing historical features would break
// replay and analytical reproducibility.
//
// This package includes a concurrency-safe in-memory implementation. A
// PostgreSQL adapter remains separate until the repository migration-number
// conflict is audited against the deployed schema_migrations table.
package featurestore
