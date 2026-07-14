// Package endpointevidence converts ranked airport candidates and persisted
// trajectory endpoint quality into a contract-safe Route Intelligence endpoint.
//
// The builder refuses premature selection. A probable endpoint is returned only
// when the leading candidate passes the configured confidence threshold and is
// sufficiently separated from the runner-up candidate.
package endpointevidence
