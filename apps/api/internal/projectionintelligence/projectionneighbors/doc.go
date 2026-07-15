// Package projectionneighbors selects deterministic historical trajectory
// neighbors for local continuation analysis.
//
// Selection is strictly bounded by an as-of time. Current-flight points after
// that time are excluded, and candidate trajectories must end before the
// current trajectory begins. The package produces evidence only; it does not
// yet generate forecast coordinates.
package projectionneighbors
