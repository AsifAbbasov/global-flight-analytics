// Package projectioncontract defines the versioned, prediction-specific
// contract used by conservative short-horizon projection and estimated
// arrival analytics.
//
// The contract keeps observed inputs separate from estimated outputs, records
// the reasoning class of every method, prevents future-data leakage, requires
// explicit uncertainty, and carries the research-only operational scope guard.
package projectioncontract
