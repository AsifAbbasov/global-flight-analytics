// Package validator validates versioned FlightFeatures after extraction and
// before persistence or analytical use.
//
// The validator does not recalculate feature values. It checks structural
// identity, time boundaries, schema-aligned evidence counts, numeric ranges,
// cross-field relationships, provenance, and quality metadata. It returns a
// defensive copy whose ValidationStatus is valid, limited, or invalid and
// whose limitations include deterministic validation findings.
package validator
