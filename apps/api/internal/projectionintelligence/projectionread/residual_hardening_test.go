package projectionread

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestLoadHistoricalCandidatesBackfillsRejectedIdentifiers(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(current, asOfTime)
	rejectedID := "93aa02ab-7061-4e9e-a238-d32710371ee3"
	candidate := projectionReadTrajectory(
		"83aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime.Add(-24*time.Hour),
	)
	candidate.Points = nil
	candidate.PointCount = 3

	client := &scriptedClient{
		rowsQueue: []*scriptedRows{
			{
				values: [][]any{
					{rejectedID},
					{candidate.ID},
				},
			},
			{
				values: [][]any{
					projectionReadPointRow(
						"candidate-state-a",
						candidate.StartTime,
						40.10,
						49.20,
					),
					projectionReadPointRow(
						"candidate-state-b",
						candidate.EndTime,
						40.20,
						49.30,
					),
				},
			},
		},
	}
	repository := &trajectoryRepositoryStub{
		items: map[string]trajectory.FlightTrajectory{
			candidate.ID: candidate,
		},
		errs: map[string]error{
			rejectedID: trajectory.ErrNotFound,
		},
	}
	policy := DefaultPolicy().DataSource
	policy.MaximumHistoricalCandidateCount = 1
	source, err := newPostgresDataSource(
		client,
		repository,
		policy,
	)
	if err != nil {
		t.Fatalf("newPostgresDataSource() error = %v", err)
	}

	result, err := source.LoadHistoricalCandidates(
		context.Background(),
		current,
		route,
		asOfTime,
	)
	if err != nil {
		t.Fatalf("LoadHistoricalCandidates() error = %v", err)
	}
	if len(result) != 1 || result[0].ID != candidate.ID {
		t.Fatalf("unexpected backfilled candidates: %#v", result)
	}
	if len(client.queryCalls) != 2 {
		t.Fatalf("query calls = %d, want 2", len(client.queryCalls))
	}
	selectionLimit, ok := client.queryCalls[0].args[7].(int)
	if !ok || selectionLimit != historicalCandidateScanMultiplier {
		t.Fatalf(
			"candidate identifier scan limit = %#v, want %d",
			client.queryCalls[0].args[7],
			historicalCandidateScanMultiplier,
		)
	}
}

func TestRouteHistoryFingerprintBindsContributingRecords(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	summary := projectionReadHistory(asOfTime)
	first := []routeHistoryEvidence{
		projectionReadRouteHistoryEvidence(1, asOfTime.Add(-time.Hour)),
		projectionReadRouteHistoryEvidence(2, asOfTime.Add(-2*time.Hour)),
	}
	second := append([]routeHistoryEvidence(nil), first...)
	second[0].RouteRecordID = "route-record-changed"

	firstFingerprint := routeHistoryFingerprint(
		summary,
		routeHistoryEvidenceFingerprint(first),
	)
	secondFingerprint := routeHistoryFingerprint(
		summary,
		routeHistoryEvidenceFingerprint(second),
	)
	if firstFingerprint == secondFingerprint {
		t.Fatal(
			"route-history fingerprint ignored contributing record lineage",
		)
	}

	reversed := []routeHistoryEvidence{first[1], first[0]}
	if routeHistoryEvidenceFingerprint(first) !=
		routeHistoryEvidenceFingerprint(reversed) {
		t.Fatal("route-history evidence fingerprint depends on row order")
	}
}

func projectionReadRouteHistoryEvidencePayload(
	t *testing.T,
	lastObservedAt time.Time,
) []byte {
	t.Helper()
	offsets := []time.Duration{
		0,
		24 * time.Hour,
		2 * 24 * time.Hour,
		3 * 24 * time.Hour,
		31 * 24 * time.Hour,
		32 * 24 * time.Hour,
		33 * 24 * time.Hour,
		33*24*time.Hour + time.Hour,
	}
	items := make([]routeHistoryEvidence, 0, len(offsets))
	for index, offset := range offsets {
		items = append(
			items,
			projectionReadRouteHistoryEvidence(
				index+1,
				lastObservedAt.Add(-offset),
			),
		)
	}
	payload, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal route-history evidence: %v", err)
	}
	return payload
}

func projectionReadRouteHistoryEvidence(
	sequence int,
	asOfTime time.Time,
) routeHistoryEvidence {
	return routeHistoryEvidence{
		EvidenceID: fmt.Sprintf(
			"evidence-%02d",
			sequence,
		),
		TrajectoryID: fmt.Sprintf(
			"trajectory-%02d",
			sequence,
		),
		RouteRecordID: fmt.Sprintf(
			"route-record-%02d",
			sequence,
		),
		InputFingerprint: fmt.Sprintf(
			"sha256:%064x",
			sequence,
		),
		AsOfTimeUnixNano: asOfTime.UTC().UnixNano(),
	}
}
