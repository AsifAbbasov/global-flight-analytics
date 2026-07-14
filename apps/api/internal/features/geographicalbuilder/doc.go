// Package geographicalbuilder derives deterministic geographic features
// from trajectory coordinates without external services or dependencies.
//
// Valid trajectory points are the preferred evidence. When no usable point
// coordinate exists, ordered endpoints from non-invalid trajectory segments
// may provide a limited fallback. Invalid and non-finite coordinates are
// excluded and surfaced as explicit analytical limitations.
package geographicalbuilder
