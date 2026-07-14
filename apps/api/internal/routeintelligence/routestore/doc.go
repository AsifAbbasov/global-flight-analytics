// Package routestore persists versioned Route Intelligence results.
//
// A stored result is keyed by trajectory identifier, Route Intelligence schema
// version, and the exact analytical as-of time. Replaying the same key with the
// same input fingerprint is idempotent. Reusing the key with different evidence
// is a conflict.
package routestore
