// Package datasetprofiler produces deterministic, in-memory profiles for
// collections of FlightFeatures snapshots.
//
// Records using the target schema and carrying valid or limited validation
// status contribute to dataset statistics. Invalid, unvalidated, and
// unsupported-schema records remain visible through validation and
// rejection profiles but do not contaminate accepted-record aggregates.
package datasetprofiler
