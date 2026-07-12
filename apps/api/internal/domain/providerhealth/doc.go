// Package providerhealth defines the provider-independent operational health
// evidence used by ingestion orchestration and explainable aviation metrics.
//
// The package does not call external providers, persist data, or decide which
// provider should be selected. It converts measured request, observation, and
// budget evidence into a deterministic health snapshot.
package providerhealth
