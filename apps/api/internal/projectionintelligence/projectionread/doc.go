// Package projectionread composes the read-only Production Projection
// Intelligence query path.
//
// It loads an as-of trajectory snapshot, the latest Route Intelligence result
// available at or before that analytical time, route-scoped historical
// trajectories, and a bounded route-history summary. It then delegates all
// prediction decisions to projectionproduction.Composer.
//
// The package never stores a projection result and never mutates Route
// Intelligence or Historical Intelligence state.
package projectionread
