// Package extractor assembles versioned FlightFeatures from a caller-provided
// trajectory snapshot.
//
// The extractor does not truncate trajectory evidence. The caller must supply a
// trajectory whose end time does not exceed Request.AsOfTime. This keeps replay
// and historical extraction leakage-safe without silently inventing partial
// segments. Group builders calculate feature values; the extractor owns
// orchestration, identity, observation-window metadata, provenance, initial
// completeness, and deterministic input fingerprinting. Feature validation and
// persistence remain separate stages.
package extractor
