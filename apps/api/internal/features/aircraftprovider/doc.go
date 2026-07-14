// Package aircraftprovider adapts the existing aircraft domain lookup into
// the AircraftFeatureProvider contract used by the feature extractor.
//
// Missing aircraft metadata is a normal analytical limitation, not a
// pipeline failure. The provider therefore returns unavailable evidence for
// a missing aircraft while preserving genuine repository failures as
// errors. Positive, partial and negative lookup results are cached with
// separate time-to-live values, and concurrent requests for the same ICAO24
// are coalesced.
package aircraftprovider
