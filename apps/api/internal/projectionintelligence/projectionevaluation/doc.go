// Package projectionevaluation measures Projection Intelligence outputs
// against later observed trajectory evidence.
//
// Evaluation is explicitly separated from forecast generation: future
// observations are accepted only as truth data after the projection has
// already been produced. The package never mutates a projection and never
// feeds replay truth back into a forecast result.
package projectionevaluation
