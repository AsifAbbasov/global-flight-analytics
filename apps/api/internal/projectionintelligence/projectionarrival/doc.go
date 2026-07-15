// Package projectionarrival attaches a conservative estimated-arrival
// interval to an existing Projection Intelligence result.
//
// The estimator consumes a validated route destination, a validated position
// projection, and the current trajectory snapshot available at the projection
// as-of time. It estimates entry into an explicit airport arrival radius, not
// runway touchdown, gate arrival, or an operational Air Traffic Control time.
package projectionarrival
