// Package weathertrust evaluates whether canonical weather evidence may be
// used by later Weather Intelligence modules.
//
// The gate returns allowed, limited, or blocked. It considers contract
// validity, evidence age, forecast lead, field completeness, confidence,
// vertical applicability, and the mandatory weather context-only scope guard.
//
// An allowed or limited result never proves pilot intent, controller intent,
// rerouting reason, or maneuver cause.
package weathertrust
