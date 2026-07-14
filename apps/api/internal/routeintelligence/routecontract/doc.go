// Package routecontract defines the versioned, algorithm-independent
// contract produced by Route Intelligence.
//
// The contract deliberately separates route inference results from the
// current endpoint-proximity implementation. Future resolvers can combine
// trajectory geometry, ground cycles, callsign evidence, source flight
// identity, airport activity, and external references without changing
// downstream consumers.
package routecontract
