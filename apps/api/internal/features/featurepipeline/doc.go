// Package featurepipeline coordinates extraction, validation, and feature
// snapshot persistence as one ordered processing boundary.
//
// Extraction and validation complete before the single store write is
// attempted. Invalid, unvalidated, inconsistent, or unknown validation
// outcomes are never passed to the Store.
package featurepipeline
