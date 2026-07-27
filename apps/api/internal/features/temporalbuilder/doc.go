// Package temporalbuilder derives deterministic UTC temporal features from
// a validated trajectory observation window.
//
// The builder treats FlightTrajectory.StartTime and EndTime as the
// authoritative feature-window boundaries. Historical as-of filtering is
// enforced by extractor snapshot validation before this builder is invoked.
// Point timestamps are preferred as supporting evidence; when point records are
// not materialized, unique persisted segment-boundary timestamps provide a
// transparent fallback without changing the authoritative window.
//
// Duration is exposed as complete seconds under the centralized
// flightfeatures.CurrentTemporalDurationRoundingPolicy contract.
package temporalbuilder
