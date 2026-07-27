// Package operationalbuilder derives deterministic operational features from
// temporally eligible trajectory point telemetry.
//
// PostgreSQL-reconstructed and in-memory points share one explicit telemetry
// availability contract. Values marked unavailable are never interpreted as
// zero measurements. Points are filtered to the authoritative trajectory
// window and ordered chronologically before sequence metrics are calculated.
// Altitude aggregates use one source for the entire trajectory, preferring
// barometric observations and using geometric observations only when no usable
// barometric series exists. Means are observation-weighted and accumulated with
// compensated summation; no interpolation or time weighting is invented.
package operationalbuilder
