package routepipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
)

type catalogCache struct {
	mutex     sync.Mutex
	catalog   *airportresolver.Catalog
	report    airportresolver.CatalogBuildReport
	expiresAt time.Time
	loaded    bool
}

func (pipeline *Pipeline) loadCatalog(
	ctx context.Context,
) (
	*airportresolver.Catalog,
	airportresolver.CatalogBuildReport,
	error,
) {
	pipeline.catalogCache.mutex.Lock()
	defer pipeline.catalogCache.mutex.Unlock()

	now := pipeline.now().UTC()
	if pipeline.catalogCache.loaded &&
		now.Before(
			pipeline.catalogCache.expiresAt,
		) {
		return pipeline.catalogCache.catalog,
			pipeline.catalogCache.report.Clone(),
			nil
	}

	airports, err := pipeline.airportLister.List(ctx)
	if err != nil {
		return nil,
			airportresolver.CatalogBuildReport{},
			fmt.Errorf(
				"list airports for Route Intelligence: %w",
				err,
			)
	}
	if err := ctx.Err(); err != nil {
		return nil,
			airportresolver.CatalogBuildReport{},
			err
	}

	catalog, report, err := airportresolver.NewCatalog(
		airports,
	)
	if err != nil {
		return nil,
			report.Clone(),
			fmt.Errorf(
				"build Route Intelligence airport catalog: %w",
				err,
			)
	}

	pipeline.catalogCache.catalog = catalog
	pipeline.catalogCache.report = report.Clone()
	pipeline.catalogCache.expiresAt = now.Add(
		pipeline.airportCatalogTTL,
	)
	pipeline.catalogCache.loaded = true

	return catalog,
		report.Clone(),
		nil
}
