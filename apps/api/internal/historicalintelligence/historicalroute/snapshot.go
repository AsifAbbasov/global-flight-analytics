package historicalroute

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func validateSnapshotWindow(
	snapshot historicalread.Snapshot,
	plan historicalwindow.Plan,
) error {
	queryWindow := snapshot.Query.Window
	if snapshot.IsolationLevel == "" && emptyWindow(queryWindow) {
		// Retain source compatibility for legacy in-memory fixtures. Production
		// repository snapshots always carry both isolation and query evidence.
		return nil
	}
	if emptyWindow(queryWindow) || plan.EffectiveWindow == nil {
		return ErrSnapshotWindowMismatch
	}

	expected := *plan.EffectiveWindow
	if !windowContains(queryWindow, expected) {
		return ErrSnapshotWindowMismatch
	}
	return nil
}

func routeDatasetCoverage(
	snapshot historicalread.Snapshot,
	scopeType historicalcontract.ScopeType,
) (historicalseries.DatasetCoverage, error) {
	incomplete := snapshot.RouteLimitReached || snapshot.RouteByteLimitReached
	if !incomplete {
		return historicalseries.DatasetCoverage{
			State: historicalseries.DatasetReadComplete,
		}, nil
	}
	if scopeType == historicalcontract.ScopeTypeRoute {
		return historicalseries.DatasetCoverage{}, ErrRouteScopeCoverageUnavailable
	}
	if snapshot.RouteMatchedCount <= 0 {
		return historicalseries.DatasetCoverage{}, ErrRouteMatchedCountRequired
	}
	if snapshot.RouteMatchedCount <= int64(len(snapshot.Routes)) {
		return historicalseries.DatasetCoverage{}, ErrRouteMatchedCountInvalid
	}
	return historicalseries.DatasetCoverage{
		State:        historicalseries.DatasetReadIncomplete,
		MatchedCount: snapshot.RouteMatchedCount,
	}, nil
}

func emptyWindow(window historicalcontract.TimeWindow) bool {
	return window.StartTime.IsZero() &&
		window.EndTime.IsZero() &&
		window.AsOfTime.IsZero()
}

func windowContains(
	outer historicalcontract.TimeWindow,
	inner historicalcontract.TimeWindow,
) bool {
	if outer.StartTime.IsZero() || outer.EndTime.IsZero() ||
		outer.AsOfTime.IsZero() || inner.StartTime.IsZero() ||
		inner.EndTime.IsZero() || inner.AsOfTime.IsZero() {
		return false
	}
	return !outer.StartTime.UTC().After(inner.StartTime.UTC()) &&
		!outer.EndTime.UTC().Before(inner.EndTime.UTC()) &&
		equalUTC(outer.AsOfTime, inner.AsOfTime)
}

func equalUTC(left time.Time, right time.Time) bool {
	if left.IsZero() || right.IsZero() {
		return left.IsZero() && right.IsZero()
	}
	return left.UTC().Equal(right.UTC())
}
