// Package trajectorybuilder derives deterministic trajectory-structure
// features from persisted trajectory points, segments, and coverage gaps.
//
// Collection lengths are authoritative for count features. Persisted count
// metadata is retained only as consistency evidence. Sampling metrics use a
// sorted copy of non-zero point timestamps, coverage uses the union of
// clipped gap intervals, and path efficiency uses the same shortest-arc
// geographic semantics as the geographical feature builder.
package trajectorybuilder
