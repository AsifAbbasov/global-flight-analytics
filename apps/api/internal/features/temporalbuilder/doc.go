// Package temporalbuilder derives deterministic UTC temporal features from
// a validated trajectory observation window.
//
// The builder treats FlightTrajectory.StartTime and EndTime as the
// authoritative feature-window boundaries. Point timestamps are used only
// as supporting evidence: missing, zero, or out-of-window point timestamps
// do not change the computed window features, but they are surfaced as
// explicit analytical limitations.
package temporalbuilder
