package historicalroute

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func Build(
	request Request,
) (historicalcontract.Result, error) {
	if request.Snapshot.Version != historicalread.Version {
		return historicalcontract.Result{}, ErrSnapshotVersionInvalid
	}

	canonicalPlan, err := historicalwindow.CanonicalizePlan(request.Plan)
	if err != nil {
		return historicalcontract.Result{}, err
	}
	request.Plan = canonicalPlan

	scope, originICAOCode, destinationICAOCode, err := normalizeScope(
		request.OriginICAOCode,
		request.DestinationICAOCode,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	definition, ok := historicalcontract.MetricSpecFor(request.MetricName)
	if !ok || definition.Family != historicalcontract.MetricFamilyRoute {
		return historicalcontract.Result{}, ErrMetricUnsupported
	}
	if !definition.AllowsScope(scope.Type) {
		return historicalcontract.Result{}, ErrMetricScopeUnsupported
	}
	if err := validateSnapshotWindow(request.Snapshot, request.Plan); err != nil {
		return historicalcontract.Result{}, err
	}

	selectedRecords, err := latestRouteRecords(
		request.Snapshot.Routes,
		request.Plan.AsOfTime,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	allEvidence, err := decodeRouteEvidenceSet(
		selectedRecords,
		request.Plan.AsOfTime,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	scopedEvidence := filterRouteEvidence(
		allEvidence,
		originICAOCode,
		destinationICAOCode,
	)
	if len(scopedEvidence) == 0 {
		return historicalcontract.Result{}, ErrRouteSourceEvidenceUnavailable
	}

	values, err := routeValues(
		request.Plan.Buckets,
		scopedEvidence,
		request.MetricName,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	coverage, err := routeDatasetCoverage(
		request.Snapshot,
		scope.Type,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}
	values, err = historicalseries.BindDatasetCoverage(values, coverage)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	return historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: historicalcontract.Metric{
				Name:        request.MetricName,
				Unit:        definition.Unit,
				Aggregation: definition.Aggregation,
			},
			Scope:                 scope,
			Plan:                  request.Plan,
			Values:                values,
			BuilderVersion:        Version,
			InputFingerprint:      routeFingerprint(request, scopedEvidence, originICAOCode, destinationICAOCode),
			SourceNames:           routeSourceNames(scopedEvidence),
			LatestSourceUpdatedAt: latestRouteUpdate(scopedEvidence),
			GeneratedAt:           request.GeneratedAt,
			Limitations:           routeLimitations(request.MetricName, request.Snapshot),
		},
	)
}
