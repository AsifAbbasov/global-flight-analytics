// Package featurestore defines immutable storage semantics for validated
// FlightFeatures snapshots.
//
// Snapshot identity includes trajectory, schema, processing version and exact
// as-of time. Idempotent replay requires matching input and semantic output
// fingerprints. New PostgreSQL payloads use an explicit versioned persistence
// contract while legacy unversioned rows remain readable.
//
// The package provides a bounded concurrency-safe in-memory implementation and
// a production PostgreSQL implementation with the same identifier, fingerprint,
// validation-proof and context contracts.
package featurestore
