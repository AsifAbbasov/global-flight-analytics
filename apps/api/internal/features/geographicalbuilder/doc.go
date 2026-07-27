// Package geographicalbuilder derives deterministic geographic features from
// trajectory coordinates without external services or dependencies.
//
// For validated production trajectories, point coordinates are filtered to the
// authoritative trajectory window and ordered by observation timestamp with a
// deterministic identity tie-breaker. Extractor owns the separate AsOfTime
// future-data gate before any feature builder runs.
//
// When fewer than two usable point observations exist, ordered non-invalid
// segment endpoints may provide a limited fallback. Segment fallback preserves
// global endpoints, bounds, displacement and cells, while observed path distance
// and antimeridian path crossing include only movement inside each usable segment
// and never bridge an unobserved discontinuity.
//
// Distances use the versioned mean-Earth spherical Haversine model. Geographic
// cells are decimal-degree analytical buckets, not equal-area physical cells.
package geographicalbuilder
