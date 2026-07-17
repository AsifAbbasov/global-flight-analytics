// Package sourceconstraints defines the immutable data-source and evidence
// boundaries of Global Flight Analytics.
//
// The policy is deliberately independent from any one provider. It prevents
// free, externally collected, incomplete observations from being published as
// first-party sensor evidence, satellite coverage, commercial aviation data,
// official operational data, or safety-critical guidance. It also preserves
// provider attribution, usage-scope, and deployment-availability obligations.
package sourceconstraints

const ContractVersion = "source-constraints-v2"
