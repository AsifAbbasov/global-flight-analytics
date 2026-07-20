package postgres

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestValidateTrajectoryRelationalIntegrityAcceptsCanonicalAggregate(
	t *testing.T,
) {
	item := canonicalTrajectoryIntegrityItem()

	if err := validateTrajectoryRelationalIntegrity(item); err != nil {
		t.Fatalf("validate canonical trajectory: %v", err)
	}
}

func TestValidateTrajectoryRelationalIntegrityRejectsStoredCountMismatch(
	t *testing.T,
) {
	item := canonicalTrajectoryIntegrityItem()
	item.SegmentCount = 1

	err := validateTrajectoryRelationalIntegrity(item)
	if !errors.Is(err, ErrTrajectoryRelationalIntegrity) {
		t.Fatalf("expected relational integrity error, got %v", err)
	}
}

func TestValidateTrajectoryRelationalIntegrityRejectsSequenceGap(
	t *testing.T,
) {
	item := canonicalTrajectoryIntegrityItem()
	item.Segments[1].SequenceNumber = 3

	err := validateTrajectoryRelationalIntegrity(item)
	if !errors.Is(err, ErrTrajectoryRelationalIntegrity) {
		t.Fatalf("expected relational integrity error, got %v", err)
	}
}

func TestValidateTrajectoryRelationalIntegrityRejectsChildIdentityMismatch(
	t *testing.T,
) {
	item := canonicalTrajectoryIntegrityItem()
	item.CoverageGaps[0].ICAO24 = "FFFFFF"

	err := validateTrajectoryRelationalIntegrity(item)
	if !errors.Is(err, ErrTrajectoryRelationalIntegrity) {
		t.Fatalf("expected relational integrity error, got %v", err)
	}
}

func TestValidateTrajectoryRelationalIntegrityRejectsPointTotalMismatch(
	t *testing.T,
) {
	item := canonicalTrajectoryIntegrityItem()
	item.PointCount = 5

	err := validateTrajectoryRelationalIntegrity(item)
	if !errors.Is(err, ErrTrajectoryRelationalIntegrity) {
		t.Fatalf("expected relational integrity error, got %v", err)
	}
}

func TestSaveTrajectoryValidatesRelationalIntegrityBeforeBeginningTransaction(
	t *testing.T,
) {
	sourceBytes, err := os.ReadFile("trajectory_write_repository.go")
	if err != nil {
		t.Fatalf("read trajectory_write_repository.go: %v", err)
	}

	source := string(sourceBytes)
	validationIndex := strings.Index(
		source,
		"validateTrajectoryRelationalIntegrity(item)",
	)
	beginIndex := strings.Index(source, "repository.db.BeginTx(")

	if validationIndex < 0 {
		t.Fatal("save trajectory does not call relational integrity validation")
	}
	if beginIndex < 0 {
		t.Fatal("save trajectory transaction boundary not found")
	}
	if validationIndex > beginIndex {
		t.Fatal("relational integrity validation occurs after transaction begins")
	}
}

func canonicalTrajectoryIntegrityItem() trajectory.FlightTrajectory {
	return trajectory.FlightTrajectory{
		ICAO24:           "ABC123",
		SegmentCount:     2,
		PointCount:       4,
		CoverageGapCount: 1,
		Points: []trajectory.TrackPoint4D{
			{}, {}, {}, {},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ICAO24:         "ABC123",
				SequenceNumber: 1,
				PointCount:     2,
			},
			{
				ICAO24:         "ABC123",
				SequenceNumber: 2,
				PointCount:     2,
			},
		},
		CoverageGaps: []trajectory.CoverageGap{
			{
				ICAO24: "ABC123",
			},
		},
	}
}
