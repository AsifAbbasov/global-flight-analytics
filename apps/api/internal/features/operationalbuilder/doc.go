// Package operationalbuilder derives deterministic operational features
// from trajectory point telemetry.
//
// Every operational signal is evaluated independently. Ground and airborne
// observation shares use all trajectory points, while altitude, velocity,
// vertical rate and heading metrics use only values that are finite and
// semantically usable. Missing signals therefore produce partial evidence
// rather than fabricated zero-valued availability.
package operationalbuilder
