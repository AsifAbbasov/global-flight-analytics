// Package historicalread defines bounded, deterministic historical reads over
// persisted flights, trajectories, observations, and Route Intelligence results.
//
// The package deliberately returns compact analytical source records rather than
// exposing repository models or unbounded database rows. Every read is constrained
// by one UTC window, one analytical as-of cutoff, stable ordering, and explicit
// dataset limits.
package historicalread
