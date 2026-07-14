// Package routepipeline composes the complete production Route Intelligence
// flow for one persisted trajectory.
//
// The pipeline loads a trajectory, resolves airport candidates for its first
// and last usable segment, builds conservative endpoint evidence, composes a
// validated Route Intelligence result, and stores that result idempotently.
package routepipeline
