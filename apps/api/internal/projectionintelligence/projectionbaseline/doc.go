// Package projectionbaseline implements the first conservative
// short-horizon projection baseline.
//
// The baseline uses only trajectory information available at the requested
// as-of time, applies the existing projection eligibility decision, propagates
// the latest observed ground track with explicit uncertainty growth, and
// returns a research-only Projection Intelligence contract.
package projectionbaseline
