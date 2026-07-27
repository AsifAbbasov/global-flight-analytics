// Package trajectorybuilder derives deterministic trajectory-structure features
// from one canonical evidence view.
//
// Point observations are filtered to the authoritative trajectory window,
// ordered deterministically and collapsed to one point per observation instant.
// Sampling and point-path calculations consume that same canonical sequence.
// Coverage gaps split path continuity, and segment fallback treats every usable
// segment as an independent observed part so unobserved jumps are never counted
// as flown distance. Count availability, zero-valued metrics and persisted
// point-count fallback are represented explicitly through group evidence.
package trajectorybuilder
