// Package projectionfreshness applies an explicit freshness safety gate to
// historical-neighbor evidence used for trajectory prediction.
//
// The guard is deliberately separate from the Pattern Confidence score. It
// can block historical continuation even when similarity is high, preventing
// stale patterns from being presented as current predictive evidence.
package projectionfreshness
