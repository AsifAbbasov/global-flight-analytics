// Package weathercontract defines the canonical, source-aware,
// research-only weather feature boundary used by Weather Intelligence.
//
// The contract separates provider payloads from analytical weather context.
// It records when weather evidence was valid, when it became available,
// where it applies, which fields are present, and why it may or may not be
// trusted by later trajectory-alignment and uncertainty modules.
//
// Weather evidence is context only. It must never be represented as proof of
// pilot intent, controller intent, rerouting reason, or maneuver cause.
package weathercontract
