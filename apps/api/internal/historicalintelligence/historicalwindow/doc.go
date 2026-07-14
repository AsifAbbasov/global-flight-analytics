// Package historicalwindow creates deterministic, UTC-normalized historical
// analysis windows and closed calendar buckets.
//
// Standard hour, day, and week plans contain only complete buckets. Partial
// leading and trailing intervals, together with time after the analytical
// as-of cutoff, are represented explicitly as exclusions instead of being
// silently mixed into complete historical periods.
package historicalwindow
