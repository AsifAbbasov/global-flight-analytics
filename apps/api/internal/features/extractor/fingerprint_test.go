package extractor

import (
	"strings"
	"testing"
	"time"
)

func TestFingerprintTrajectoryIsDeterministic(t *testing.T) {
	item := validRequest().Trajectory

	first, err := fingerprintTrajectory(item)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}
	second, err := fingerprintTrajectory(item)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}

	if first != second {
		t.Fatalf(
			"fingerprints differ: first=%q second=%q",
			first,
			second,
		)
	}
	if !strings.HasPrefix(first, fingerprintPrefix) ||
		len(first) != len(fingerprintPrefix)+64 {
		t.Fatalf("unexpected fingerprint format: %q", first)
	}
}

func TestFingerprintTrajectoryNormalizesTimeZones(t *testing.T) {
	firstItem := validRequest().Trajectory
	secondItem := cloneTrajectory(firstItem)
	location := time.FixedZone("UTC+04", 4*60*60)

	secondItem.StartTime = firstItem.StartTime.In(location)
	secondItem.EndTime = firstItem.EndTime.In(location)
	secondItem.CreatedAt = firstItem.CreatedAt.In(location)
	secondItem.UpdatedAt = firstItem.UpdatedAt.In(location)

	for index := range secondItem.Points {
		secondItem.Points[index].ObservedAt =
			firstItem.Points[index].ObservedAt.In(location)
	}
	for index := range secondItem.Segments {
		secondItem.Segments[index].StartTime =
			firstItem.Segments[index].StartTime.In(location)
		secondItem.Segments[index].EndTime =
			firstItem.Segments[index].EndTime.In(location)
		secondItem.Segments[index].CreatedAt =
			firstItem.Segments[index].CreatedAt.In(location)
	}
	for index := range secondItem.CoverageGaps {
		secondItem.CoverageGaps[index].StartTime =
			firstItem.CoverageGaps[index].StartTime.In(location)
		secondItem.CoverageGaps[index].EndTime =
			firstItem.CoverageGaps[index].EndTime.In(location)
		secondItem.CoverageGaps[index].CreatedAt =
			firstItem.CoverageGaps[index].CreatedAt.In(location)
	}

	first, err := fingerprintTrajectory(firstItem)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}
	second, err := fingerprintTrajectory(secondItem)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}

	if first != second {
		t.Fatalf(
			"equivalent instants produced different fingerprints: first=%q second=%q",
			first,
			second,
		)
	}
}

func TestFingerprintTrajectoryChangesWithEvidence(t *testing.T) {
	firstItem := validRequest().Trajectory
	secondItem := cloneTrajectory(firstItem)
	secondItem.Points[0].Latitude += 0.01

	first, err := fingerprintTrajectory(firstItem)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}
	second, err := fingerprintTrajectory(secondItem)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}

	if first == second {
		t.Fatal("changed trajectory evidence produced the same fingerprint")
	}
}

func TestFingerprintTrajectoryPreservesEvidenceOrder(t *testing.T) {
	firstItem := validRequest().Trajectory
	secondItem := cloneTrajectory(firstItem)
	secondItem.Points[0], secondItem.Points[1] =
		secondItem.Points[1], secondItem.Points[0]

	first, err := fingerprintTrajectory(firstItem)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}
	second, err := fingerprintTrajectory(secondItem)
	if err != nil {
		t.Fatalf("fingerprintTrajectory() error = %v", err)
	}

	if first == second {
		t.Fatal("changed point order produced the same fingerprint")
	}
}
