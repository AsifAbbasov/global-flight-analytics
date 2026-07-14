// Package routeresolver composes origin and destination endpoint evidence into
// the versioned Route Intelligence result contract.
//
// The resolver is deliberately conservative: endpoint selection remains the
// responsibility of endpointevidence, while this package determines route
// completeness, overall confidence, route-level limitations, distance,
// provenance, and final contract validity.
package routeresolver
