package temporalbuilder

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestBuilderBuildsCompleteSameDayUTCFeatures(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		15,
		30,
		0,
		time.UTC,
	)
	endTime := time.Date(
		2026,
		time.July,
		14,
		10,
		45,
		45,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		StartTime: startTime,
		EndTime:   endTime,
		DurationSeconds: int64(
			endTime.Sub(startTime) / time.Second,
		),
		Points: []trajectory.TrackPoint4D{
			{ObservedAt: startTime},
			{ObservedAt: startTime.Add(time.Hour)},
			{ObservedAt: endTime},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	want := flightfeatures.TemporalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:               flightfeatures.AvailabilityStatusAvailable,
			AvailableFieldCount:  TemporalFeatureFieldCount,
			TotalFieldCount:      TemporalFeatureFieldCount,
			SupportingPointCount: 3,
		},
		DurationSeconds: int64(
			endTime.Sub(startTime) / time.Second,
		),
		StartHourUTC:        8,
		EndHourUTC:          10,
		StartWeekday:        time.Tuesday,
		EndWeekday:          time.Tuesday,
		StartMinuteOfDayUTC: 8*60 + 15,
		EndMinuteOfDayUTC:   10*60 + 45,
		CrossesUTCMidnight:  false,
	}

	if !reflect.DeepEqual(features, want) {
		t.Fatalf(
			"features = %#v, want %#v",
			features,
			want,
		)
	}
}

func TestBuilderNormalizesWindowAndPointsToUTC(t *testing.T) {
	location := time.FixedZone("UTC+04", 4*60*60)
	startTime := time.Date(
		2026,
		time.July,
		14,
		23,
		30,
		0,
		0,
		location,
	)
	endTime := startTime.Add(2 * time.Hour)
	item := trajectory.FlightTrajectory{
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: 7200,
		Points: []trajectory.TrackPoint4D{
			{ObservedAt: startTime},
			{ObservedAt: endTime},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.StartHourUTC != 19 ||
		features.EndHourUTC != 21 ||
		features.StartMinuteOfDayUTC != 19*60+30 ||
		features.EndMinuteOfDayUTC != 21*60+30 {
		t.Fatalf(
			"unexpected UTC features: %#v",
			features,
		)
	}
	if features.CrossesUTCMidnight {
		t.Fatal(
			"local midnight crossing must not be treated as UTC midnight crossing",
		)
	}
	if features.Evidence.SupportingPointCount != 2 {
		t.Fatalf(
			"supporting points = %d, want 2",
			features.Evidence.SupportingPointCount,
		)
	}
}

func TestBuilderDetectsUTCCalendarBoundary(t *testing.T) {
	startTime := time.Date(
		2026,
		time.December,
		31,
		23,
		59,
		30,
		0,
		time.UTC,
	)
	endTime := startTime.Add(time.Minute)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 60,
			Points: []trajectory.TrackPoint4D{
				{ObservedAt: startTime},
				{ObservedAt: endTime},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !features.CrossesUTCMidnight {
		t.Fatal("expected UTC calendar boundary crossing")
	}
	if features.StartWeekday != time.Thursday ||
		features.EndWeekday != time.Friday ||
		features.StartMinuteOfDayUTC != 1439 ||
		features.EndMinuteOfDayUTC != 0 {
		t.Fatalf(
			"unexpected boundary features: %#v",
			features,
		)
	}
}

func TestBuilderSupportsZeroDurationWindow(t *testing.T) {
	instant := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime: instant,
			EndTime:   instant,
			Points: []trajectory.TrackPoint4D{
				{ObservedAt: instant},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.DurationSeconds != 0 ||
		features.CrossesUTCMidnight ||
		features.Evidence.SupportingPointCount != 1 {
		t.Fatalf(
			"unexpected zero-duration features: %#v",
			features,
		)
	}
}

func TestBuilderRejectsInvalidWindow(t *testing.T) {
	validTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	tests := []struct {
		name    string
		item    trajectory.FlightTrajectory
		wantErr error
	}{
		{
			name: "missing start",
			item: trajectory.FlightTrajectory{
				EndTime: validTime,
			},
			wantErr: ErrTrajectoryStartTimeRequired,
		},
		{
			name: "missing end",
			item: trajectory.FlightTrajectory{
				StartTime: validTime,
			},
			wantErr: ErrTrajectoryEndTimeRequired,
		},
		{
			name: "end before start",
			item: trajectory.FlightTrajectory{
				StartTime: validTime,
				EndTime:   validTime.Add(-time.Second),
			},
			wantErr: ErrInvalidTrajectoryWindow,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New().Build(
				context.Background(),
				test.item,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"Build() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}

func TestBuilderEvaluatesPointEvidence(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(time.Hour)
	item := trajectory.FlightTrajectory{
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: 3600,
		Points: []trajectory.TrackPoint4D{
			{ObservedAt: startTime},
			{ObservedAt: startTime.Add(30 * time.Minute)},
			{ObservedAt: endTime},
			{ObservedAt: time.Time{}},
			{ObservedAt: startTime.Add(-time.Second)},
			{ObservedAt: endTime.Add(time.Second)},
		},
	}

	features, err := New().Build(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.Evidence.SupportingPointCount != 3 {
		t.Fatalf(
			"supporting points = %d, want 3",
			features.Evidence.SupportingPointCount,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"temporal_point_timestamp_missing",
	) {
		t.Fatalf(
			"missing zero timestamp limitation: %#v",
			features.Evidence.Limitations,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"temporal_point_outside_window",
	) {
		t.Fatalf(
			"missing outside-window limitation: %#v",
			features.Evidence.Limitations,
		)
	}
	if hasLimitation(
		features.Evidence.Limitations,
		"temporal_point_evidence_unusable",
	) {
		t.Fatal("usable point evidence was incorrectly rejected")
	}
}

func TestBuilderMarksMissingAndUnusablePointEvidence(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(time.Hour)

	tests := []struct {
		name      string
		points    []trajectory.TrackPoint4D
		wantCodes []string
	}{
		{
			name:   "missing points",
			points: nil,
			wantCodes: []string{
				"temporal_point_evidence_unavailable",
			},
		},
		{
			name: "all points unusable",
			points: []trajectory.TrackPoint4D{
				{ObservedAt: time.Time{}},
				{ObservedAt: startTime.Add(-time.Second)},
			},
			wantCodes: []string{
				"temporal_point_timestamp_missing",
				"temporal_point_outside_window",
				"temporal_point_evidence_unusable",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			features, err := New().Build(
				context.Background(),
				trajectory.FlightTrajectory{
					StartTime:       startTime,
					EndTime:         endTime,
					DurationSeconds: 3600,
					Points:          test.points,
				},
			)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			if features.Evidence.Status !=
				flightfeatures.AvailabilityStatusAvailable ||
				features.Evidence.AvailableFieldCount !=
					TemporalFeatureFieldCount ||
				features.Evidence.SupportingPointCount != 0 {
				t.Fatalf(
					"unexpected evidence: %#v",
					features.Evidence,
				)
			}
			for _, code := range test.wantCodes {
				if !hasLimitation(
					features.Evidence.Limitations,
					code,
				) {
					t.Fatalf(
						"missing limitation %q in %#v",
						code,
						features.Evidence.Limitations,
					)
				}
			}
		})
	}
}

func TestBuilderReportsDurationMetadataMismatch(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(time.Hour)

	features, err := New().Build(
		context.Background(),
		trajectory.FlightTrajectory{
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 3599,
			Points: []trajectory.TrackPoint4D{
				{ObservedAt: startTime},
				{ObservedAt: endTime},
			},
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if features.DurationSeconds != 3600 {
		t.Fatalf(
			"duration = %d, want 3600",
			features.DurationSeconds,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"trajectory_duration_metadata_mismatch",
	) {
		t.Fatalf(
			"missing duration mismatch limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestBuilderDoesNotMutateInput(t *testing.T) {
	location := time.FixedZone("UTC+04", 4*60*60)
	startTime := time.Date(
		2026,
		time.July,
		14,
		12,
		0,
		0,
		0,
		location,
	)
	item := trajectory.FlightTrajectory{
		StartTime:       startTime,
		EndTime:         startTime.Add(time.Hour),
		DurationSeconds: 3600,
		Points: []trajectory.TrackPoint4D{
			{ObservedAt: startTime},
		},
	}
	original := item
	original.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)

	if _, err := New().Build(
		context.Background(),
		item,
	); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !reflect.DeepEqual(item, original) {
		t.Fatalf(
			"input was mutated\ninput=%#v\noriginal=%#v",
			item,
			original,
		)
	}
}

func TestBuilderPreservesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New().Build(
		ctx,
		trajectory.FlightTrajectory{},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Build() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestCloneFeaturesDoesNotShareLimitations(t *testing.T) {
	features := flightfeatures.TemporalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Limitations: []flightfeatures.FeatureLimitation{
				{Code: "original"},
			},
		},
	}

	cloned := cloneFeatures(features)
	cloned.Evidence.Limitations[0].Code = "changed"

	if features.Evidence.Limitations[0].Code != "original" {
		t.Fatal("cloneFeatures() shared limitations")
	}
}

func TestTemporalBuilderContractConstantsRemainStable(
	t *testing.T,
) {
	if Version != "temporal-feature-builder-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if TemporalFeatureFieldCount != 8 {
		t.Fatalf(
			"TemporalFeatureFieldCount = %d",
			TemporalFeatureFieldCount,
		)
	}
}

func hasLimitation(
	limitations []flightfeatures.FeatureLimitation,
	code string,
) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}

	return false
}
