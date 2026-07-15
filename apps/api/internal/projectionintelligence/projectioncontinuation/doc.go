// Package projectioncontinuation implements the local historical-neighbor
// continuation baseline and its conservative kinematic fallback.
//
// The method translates observed continuations from selected historical
// neighbors onto the current trajectory endpoint, combines them by explicit
// similarity weights, derives uncertainty from configured growth and neighbor
// disagreement, and preserves the research-only Projection Intelligence
// contract.
package projectioncontinuation
