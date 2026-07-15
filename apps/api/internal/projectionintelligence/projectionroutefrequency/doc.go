// Package projectionroutefrequency applies an explicit low-frequency route
// safety gate before historical continuation is allowed.
//
// The guard consumes a validated Route Intelligence result and a bounded,
// caller-supplied route-history summary. It never queries storage directly and
// never converts sparse route history into an unsupported prediction claim.
package projectionroutefrequency
