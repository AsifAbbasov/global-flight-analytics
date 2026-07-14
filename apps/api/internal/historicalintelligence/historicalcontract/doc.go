// Package historicalcontract defines the versioned data contract used by
// Historical Intelligence calculations, persistence, HTTP delivery, and
// frontend visualization.
//
// The contract distinguishes analytical time coverage from data quality,
// prevents future evidence, requires deterministic bucket ordering, and keeps
// comparison, confidence, limitation, and provenance metadata explicit.
package historicalcontract
