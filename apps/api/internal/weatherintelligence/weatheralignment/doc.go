// Package weatheralignment aligns trusted canonical weather evidence to
// existing four-dimensional flight trajectory points.
//
// Alignment considers latitude, longitude, altitude, and observation time.
// Surface weather may align only to ground trajectory points. Airborne points
// require trusted trajectory-context weather with a usable vertical reference.
//
// Alignment expresses contextual proximity. It never proves pilot intent,
// controller intent, rerouting reason, or maneuver cause.
package weatheralignment
