// Package weatheruncertainty applies trusted weather context to the explicit
// uncertainty already published by Projection Intelligence.
//
// The modifier may preserve or increase projection uncertainty. It never
// reduces an existing uncertainty radius, never changes projected coordinates,
// and never treats weather as proof of pilot intent, controller intent,
// rerouting reason, or maneuver cause.
//
// All thresholds and weights are project-derived research heuristics and are
// not operational aviation limits.
package weatheruncertainty
