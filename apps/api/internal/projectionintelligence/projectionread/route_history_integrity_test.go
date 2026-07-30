package projectionread

import (
	"strings"
	"testing"
)

func TestRouteHistorySummarySQLExcludesCurrentFlightAndDeduplicatesEvidence(t *testing.T) {
	for _, fragment := range []string{
		"route_result.trajectory_id::text <> $7",
		"trajectory.flight_id::text <> $8",
		"COALESCE(",
		") AS evidence_id",
		"latest_route_per_evidence",
		"SELECT DISTINCT ON (evidence_id)",
		"as_of_time AT TIME ZONE 'UTC'",
		"WHERE as_of_time >= $6",
	} {
		if !strings.Contains(routeHistorySummarySQL, fragment) {
			t.Fatalf("route-history SQL is missing %q", fragment)
		}
	}

	for _, forbidden := range []string{
		"FROM latest_route_per_trajectory;",
		"COUNT(\n\t\t\tDISTINCT COALESCE(",
	} {
		if strings.Contains(routeHistorySummarySQL, forbidden) {
			t.Fatalf("route-history SQL retains forbidden fragment %q", forbidden)
		}
	}
}
