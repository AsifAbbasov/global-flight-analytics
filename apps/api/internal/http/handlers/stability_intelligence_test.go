package handlers

import (
	"errors"
	"testing"
	"time"
)

func TestParseStabilityIntelligenceReadRequest(
	t *testing.T,
) {
	request, err :=
		parseStabilityIntelligenceReadRequest(
			"2E0DC3A0-4C5E-4BDA-A5AD-5A14DE916A41",
			"2035-01-15T12:00:00Z,2035-01-15T12:00:30Z,2035-01-15T12:01:00Z",
			"300",
		)
	if err != nil {
		t.Fatal(err)
	}
	if request.TrajectoryID !=
		"2e0dc3a0-4c5e-4bda-a5ad-5a14de916a41" ||
		len(request.AsOfTimes) != 3 ||
		request.RequestedDuration !=
			5*time.Minute {
		t.Fatalf(
			"unexpected request: %#v",
			request,
		)
	}
}

func TestParseStabilityAsOfTimesRejectsDuplicate(
	t *testing.T,
) {
	_, err := parseStabilityAsOfTimes(
		"2035-01-15T12:00:00Z,2035-01-15T12:00:00Z",
	)
	if !errors.Is(
		err,
		errStabilityAsOfTimesInvalid,
	) {
		t.Fatalf(
			"error = %v",
			err,
		)
	}
}

func TestParseStabilityAsOfTimesRejectsSingleTimestamp(
	t *testing.T,
) {
	_, err := parseStabilityAsOfTimes(
		"2035-01-15T12:00:00Z",
	)
	if !errors.Is(
		err,
		errStabilityAsOfTimesInvalid,
	) {
		t.Fatalf(
			"error = %v",
			err,
		)
	}
}
