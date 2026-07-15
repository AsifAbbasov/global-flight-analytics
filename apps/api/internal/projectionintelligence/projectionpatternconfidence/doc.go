// Package projectionpatternconfidence evaluates whether selected historical
// neighbors provide enough support, similarity, freshness, and local endpoint
// proximity for a continuation method.
//
// The score is project-derived and experimental. All thresholds and weights
// are caller-provided and must be calibrated later through historical replay.
package projectionpatternconfidence
