// Package extractor assembles versioned FlightFeatures from a caller-provided
// trajectory snapshot.
//
// The extractor does not truncate trajectory evidence. It rejects trajectories
// whose end time, point observations, segment bounds, or coverage-gap bounds
// exceed Request.AsOfTime. System CreatedAt and UpdatedAt values remain
// provenance timestamps and are not treated as event-time cutoffs. Group
// builders calculate feature values; the extractor owns
// orchestration, identity, observation-window metadata, provenance, initial
// completeness, and deterministic input fingerprinting. Feature validation and
// persistence remain separate stages.
package extractor
