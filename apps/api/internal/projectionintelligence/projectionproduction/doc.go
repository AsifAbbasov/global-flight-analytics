// Package projectionproduction composes Projection Intelligence components
// against one immutable request and Horizon Plan snapshot.
//
// The composer binds route, neighbor, pattern, freshness, route-frequency,
// projection, and Estimated Arrival evidence to the same trajectory and time
// boundary. Every production decision is returned with separate request and
// composition fingerprints and an auditable fallback trace.
package projectionproduction
