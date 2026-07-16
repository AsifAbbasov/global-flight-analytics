package weatheradapter

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
)

func TestMapOpenMeteoCurrentSnapshot(
	t *testing.T,
) {
	t.Parallel()

	request := validRequest()
	result, err :=
		MapOpenMeteoCurrentSnapshot(request)
	if err != nil {
		t.Fatalf(
			"map current snapshot: %v",
			err,
		)
	}

	if result.Status !=
		weathercontract.ResultStatusLimited {
		t.Fatalf(
			"expected limited result, got %q",
			result.Status,
		)
	}
	if result.ScopeGuard !=
		weathercontract.ScopeGuardContextOnly {
		t.Fatalf(
			"unexpected scope guard %q",
			result.ScopeGuard,
		)
	}
	if len(result.Samples) != 1 {
		t.Fatalf(
			"expected one sample, got %d",
			len(result.Samples),
		)
	}

	sample := result.Samples[0]
	if sample.Source.Provider !=
		domainweather.ProviderOpenMeteo ||
		sample.Source.Dataset !=
			CurrentSnapshotDataset ||
		sample.Source.EvidenceKind !=
			weathercontract.
				EvidenceKindAnalysis {
		t.Fatalf(
			"unexpected mapped source %#v",
			sample.Source,
		)
	}
	if sample.Position.VerticalReference !=
		weathercontract.
			VerticalReferenceSurface ||
		sample.Position.AltitudeMeters != nil {
		t.Fatalf(
			"expected surface-only position, got %#v",
			sample.Position,
		)
	}
	if !sample.AvailableAt.Equal(
		request.Snapshot.RetrievedAt,
	) ||
		!sample.RetrievedAt.Equal(
			request.Snapshot.RetrievedAt,
		) {
		t.Fatal(
			"retrieval time was not used as the conservative availability boundary",
		)
	}
	if sample.Features.ConditionCode == nil ||
		*sample.Features.ConditionCode !=
			request.Snapshot.WeatherCode ||
		sample.Features.ConditionCodeScheme !=
			WMOConditionCodeScheme {
		t.Fatalf(
			"weather condition code was not preserved: %#v",
			sample.Features,
		)
	}
	if sample.Features.PresentCount() != 10 {
		t.Fatalf(
			"expected all ten mapped features, got %d",
			sample.Features.PresentCount(),
		)
	}
	if len(result.Limitations) != 3 ||
		len(result.Explanations) != 2 {
		t.Fatalf(
			"unexpected evidence explanation counts: limitations=%d explanations=%d",
			len(result.Limitations),
			len(result.Explanations),
		)
	}
	if !strings.HasPrefix(
		result.Provenance.InputFingerprint,
		"sha256:",
	) {
		t.Fatalf(
			"unexpected fingerprint %q",
			result.Provenance.InputFingerprint,
		)
	}

	report := weathercontract.Validate(result)
	if report.Status !=
		weathercontract.ValidationStatusValid {
		t.Fatalf(
			"mapped result is invalid: %#v",
			report.Issues,
		)
	}
}

func TestMapOpenMeteoCurrentSnapshotFingerprintIsDeterministic(
	t *testing.T,
) {
	t.Parallel()

	request := validRequest()

	first, err :=
		MapOpenMeteoCurrentSnapshot(request)
	if err != nil {
		t.Fatalf(
			"map first result: %v",
			err,
		)
	}
	second, err :=
		MapOpenMeteoCurrentSnapshot(request)
	if err != nil {
		t.Fatalf(
			"map second result: %v",
			err,
		)
	}

	if first.Provenance.InputFingerprint !=
		second.Provenance.InputFingerprint {
		t.Fatal(
			"same normalized input produced different fingerprints",
		)
	}

	changedRequest := request
	changedRequest.Snapshot.WindSpeedMetersPerSecond++
	changed, err :=
		MapOpenMeteoCurrentSnapshot(
			changedRequest,
		)
	if err != nil {
		t.Fatalf(
			"map changed result: %v",
			err,
		)
	}
	if changed.Provenance.InputFingerprint ==
		first.Provenance.InputFingerprint {
		t.Fatal(
			"changed weather input did not change the fingerprint",
		)
	}
}

func TestMapOpenMeteoCurrentSnapshotRejectsInvalidIdentity(
	t *testing.T,
) {
	t.Parallel()

	testCases := []struct {
		name      string
		change    func(*Request)
		targetErr error
	}{
		{
			name: "trajectory id",
			change: func(request *Request) {
				request.TrajectoryID = " "
			},
			targetErr: ErrTrajectoryIDRequired,
		},
		{
			name: "as-of time",
			change: func(request *Request) {
				request.AsOfTime = time.Time{}
			},
			targetErr: ErrAsOfTimeRequired,
		},
		{
			name: "generated-at time",
			change: func(request *Request) {
				request.GeneratedAt =
					request.AsOfTime.Add(
						-time.Second,
					)
			},
			targetErr: ErrGeneratedAtInvalid,
		},
		{
			name: "provider",
			change: func(request *Request) {
				request.Snapshot.Provider =
					"other"
			},
			targetErr: ErrProviderMismatch,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				request := validRequest()
				testCase.change(&request)

				_, err :=
					MapOpenMeteoCurrentSnapshot(
						request,
					)
				if !errors.Is(
					err,
					testCase.targetErr,
				) {
					t.Fatalf(
						"expected %v, got %v",
						testCase.targetErr,
						err,
					)
				}
			},
		)
	}
}

func TestMapOpenMeteoCurrentSnapshotRejectsFutureLeak(
	t *testing.T,
) {
	t.Parallel()

	request := validRequest()
	request.Snapshot.RetrievedAt =
		request.AsOfTime.Add(time.Second)
	request.GeneratedAt =
		request.Snapshot.RetrievedAt

	_, err := MapOpenMeteoCurrentSnapshot(
		request,
	)
	if !errors.Is(
		err,
		ErrFutureSnapshotEvidence,
	) {
		t.Fatalf(
			"expected future snapshot error, got %v",
			err,
		)
	}
}

func TestMapOpenMeteoCurrentSnapshotRejectsInvalidTimes(
	t *testing.T,
) {
	t.Parallel()

	request := validRequest()
	request.Snapshot.ObservedAt =
		request.Snapshot.RetrievedAt.Add(
			time.Second,
		)

	_, err := MapOpenMeteoCurrentSnapshot(
		request,
	)
	if !errors.Is(
		err,
		ErrSnapshotTimeInvalid,
	) {
		t.Fatalf(
			"expected snapshot time error, got %v",
			err,
		)
	}
}

func TestMapOpenMeteoCurrentSnapshotRejectsInvalidWeatherValues(
	t *testing.T,
) {
	t.Parallel()

	request := validRequest()
	request.Snapshot.SurfacePressureHPA =
		math.Inf(1)

	_, err := MapOpenMeteoCurrentSnapshot(
		request,
	)
	if !errors.Is(
		err,
		ErrMappedResultInvalid,
	) {
		t.Fatalf(
			"expected mapped result validation error, got %v",
			err,
		)
	}
}

func TestWeatherConditionCodeRequiresScheme(
	t *testing.T,
) {
	t.Parallel()

	result, err :=
		MapOpenMeteoCurrentSnapshot(
			validRequest(),
		)
	if err != nil {
		t.Fatalf(
			"map current snapshot: %v",
			err,
		)
	}

	result.Samples[0].
		Features.ConditionCodeScheme = ""

	report := weathercontract.Validate(result)
	if !report.HasCode(
		weathercontract.
			IssueConditionCodeInvalid,
	) {
		t.Fatalf(
			"expected condition code issue, got %#v",
			report.Issues,
		)
	}
}

func TestMappedResultDoesNotAliasSnapshotValues(
	t *testing.T,
) {
	t.Parallel()

	request := validRequest()
	result, err :=
		MapOpenMeteoCurrentSnapshot(request)
	if err != nil {
		t.Fatalf(
			"map current snapshot: %v",
			err,
		)
	}

	*result.Samples[0].
		Features.TemperatureCelsius = 99
	*result.Samples[0].
		Features.ConditionCode = 99

	if request.Snapshot.TemperatureCelsius ==
		99 ||
		request.Snapshot.WeatherCode == 99 {
		t.Fatal(
			"mapped result aliases request snapshot values",
		)
	}
}

func validRequest() Request {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	return Request{
		TrajectoryID: "trajectory-1",
		AsOfTime:     asOfTime,
		GeneratedAt:  asOfTime.Add(time.Minute),
		Snapshot: domainweather.CurrentSnapshot{
			Provider: domainweather.
				ProviderOpenMeteo,
			Latitude:  40.4675,
			Longitude: 50.0467,
			ObservedAt: asOfTime.Add(
				-10 * time.Minute,
			),
			TemperatureCelsius:       23.5,
			RelativeHumidityPercent:  54,
			PrecipitationMillimeters: 0,
			RainMillimeters:          0,
			WeatherCode:              1,
			CloudCoverPercent:        18,
			SurfacePressureHPA:       1008.2,
			WindSpeedMetersPerSecond: 7.2,
			WindDirectionDegrees:     245,
			WindGustsMetersPerSecond: 10.1,
			RetrievedAt: asOfTime.Add(
				-5 * time.Minute,
			),
		},
	}
}
