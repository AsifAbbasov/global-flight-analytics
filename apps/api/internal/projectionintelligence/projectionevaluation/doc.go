// Package projectionevaluation measures immutable Projection Intelligence
// snapshots against later observed trajectory and arrival evidence.
//
// Replay truth is admitted only through explicit point-availability evidence,
// so event time and system knowledge time remain distinct. Evaluation never
// mutates a projection and never feeds future truth back into forecast
// generation. Results publish canonical projection, truth, policy, endpoint,
// lead-time, confidence-calibration, arrival-availability, and aggregate
// lineage suitable for bounded offline research comparison.
package projectionevaluation
