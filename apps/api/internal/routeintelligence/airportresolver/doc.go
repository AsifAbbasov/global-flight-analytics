// Package airportresolver builds a deterministic airport catalog and
// resolves geographically plausible airport candidates for Route
// Intelligence endpoint evidence.
//
// The package intentionally does not choose the final origin or
// destination. It ranks proximity candidates so later evidence builders
// can combine geometry with ground cycles, callsign tokens, source flight
// identity, airport activity, and external references.
package airportresolver
