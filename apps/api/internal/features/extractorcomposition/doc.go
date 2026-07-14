// Package extractorcomposition constructs the production FlightFeatures
// extractor from the concrete feature builders and aircraft metadata
// provider.
//
// This package is intentionally separate from extractor and the builder
// packages so dependency direction remains acyclic: builders implement
// extractor contracts, while composition imports both sides only at the
// application wiring boundary.
package extractorcomposition
