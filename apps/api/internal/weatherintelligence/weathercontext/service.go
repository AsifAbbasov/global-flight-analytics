// Package weathercontext composes the production Weather Context read model
// from bounded trajectory, weather, projection, and uncertainty evidence.
package weathercontext

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheradapter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
)

const (
	Version            = "weather-context-production-composition-v1"
	FingerprintVersion = "weather-context-production-composition-fingerprint-v1"
)

var (
	ErrTrajectoryReaderRequired = errors.New(
		"Weather Context trajectory reader is required",
	)
	ErrWeatherSnapshotReaderRequired = errors.New(
		"Weather Context weather snapshot reader is required",
	)
	ErrProjectionReaderRequired = errors.New(
		"Weather Context projection reader is required",
	)
	ErrServiceUnavailable = errors.New(
		"Weather Context production service is unavailable",
	)
	ErrInvalidRequest = errors.New(
		"Weather Context production request is invalid",
	)
	ErrTrajectoryNotFound = errors.New(
		"Weather Context trajectory was not found",
	)
	ErrWeatherNotFound = errors.New(
		"Weather Context weather snapshot was not found",
	)
	ErrProjectionNotFound = errors.New(
		"Weather Context projection was not found",
	)
	ErrTrajectoryInvalid = errors.New(
		"Weather Context trajectory evidence is invalid",
	)
	ErrProjectionInvalid = errors.New(
		"Weather Context projection evidence is invalid",
	)
	ErrGeneratedAtInvalid = errors.New(
		"Weather Context generated-at time is invalid",
	)
	ErrResultInvalid = errors.New(
		"Weather Context production result is invalid",
	)
)

type Request struct {
	TrajectoryID      string
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type WeatherSnapshotRequest struct {
	Latitude  float64
	Longitude float64
	AsOfTime  time.Time
}

type ProjectionRequest struct {
	TrajectoryID      string
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type TrajectoryReader interface {
	GetTrajectoryByID(
		context.Context,
		string,
	) (trajectory.FlightTrajectory, error)
}

type WeatherSnapshotReader interface {
	GetLatestSnapshot(
		context.Context,
		WeatherSnapshotRequest,
	) (domainweather.CurrentSnapshot, error)
}

type ProjectionReader interface {
	GetProjection(
		context.Context,
		ProjectionRequest,
	) (projectionproduction.Result, error)
}

type Config struct {
	TrajectoryReader      TrajectoryReader
	WeatherSnapshotReader WeatherSnapshotReader
	ProjectionReader      ProjectionReader

	TrustPolicy       weathertrust.Policy
	AlignmentPolicy   weatheralignment.Policy
	EncounterPolicy   weatherencounter.Policy
	UncertaintyPolicy weatheruncertainty.Policy

	Now func() time.Time
}

type Service struct {
	trajectoryReader      TrajectoryReader
	weatherSnapshotReader WeatherSnapshotReader
	projectionReader      ProjectionReader

	trustPolicy       weathertrust.Policy
	alignmentPolicy   weatheralignment.Policy
	encounterPolicy   weatherencounter.Policy
	uncertaintyPolicy weatheruncertainty.Policy

	now func() time.Time
}

func NewService(
	config Config,
) (*Service, error) {
	if config.TrajectoryReader == nil {
		return nil, ErrTrajectoryReaderRequired
	}
	if config.WeatherSnapshotReader == nil {
		return nil, ErrWeatherSnapshotReaderRequired
	}
	if config.ProjectionReader == nil {
		return nil, ErrProjectionReaderRequired
	}
	if err := config.TrustPolicy.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate Weather Context trust policy: %w",
			err,
		)
	}
	if err := config.AlignmentPolicy.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate Weather Context alignment policy: %w",
			err,
		)
	}
	if err := config.EncounterPolicy.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate Weather Context encounter policy: %w",
			err,
		)
	}
	if err := config.UncertaintyPolicy.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate Weather Context uncertainty policy: %w",
			err,
		)
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		trajectoryReader:      config.TrajectoryReader,
		weatherSnapshotReader: config.WeatherSnapshotReader,
		projectionReader:      config.ProjectionReader,
		trustPolicy:           config.TrustPolicy,
		alignmentPolicy:       config.AlignmentPolicy,
		encounterPolicy:       config.EncounterPolicy,
		uncertaintyPolicy:     config.UncertaintyPolicy,
		now:                   now,
	}, nil
}

type Result struct {
	Version string

	Weather     weathercontract.Result
	Trust       weathertrust.Result
	Alignment   weatheralignment.Result
	Encounter   weatherencounter.Result
	Uncertainty weatheruncertainty.Result

	InputFingerprint string
	GeneratedAt      time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Weather = result.Weather.Clone()
	cloned.Trust = result.Trust.Clone()
	cloned.Alignment = result.Alignment.Clone()
	cloned.Encounter = result.Encounter.Clone()
	cloned.Uncertainty = result.Uncertainty.Clone()
	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version {
		return fmt.Errorf(
			"Weather Context production version is invalid",
		)
	}
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return fmt.Errorf(
			"Weather Context production fingerprint is invalid",
		)
	}
	if result.GeneratedAt.IsZero() {
		return fmt.Errorf(
			"Weather Context production generated-at time is required",
		)
	}

	weatherReport := weathercontract.Validate(
		result.Weather,
	)
	if weatherReport.Status !=
		weathercontract.ValidationStatusValid {
		return fmt.Errorf(
			"Weather Context weather contract is invalid: %v",
			weatherReport.Issues,
		)
	}
	if err := result.Trust.Validate(); err != nil {
		return fmt.Errorf(
			"Weather Context trust result is invalid: %w",
			err,
		)
	}
	if err := result.Alignment.Validate(); err != nil {
		return fmt.Errorf(
			"Weather Context alignment result is invalid: %w",
			err,
		)
	}
	if err := result.Encounter.Validate(); err != nil {
		return fmt.Errorf(
			"Weather Context encounter result is invalid: %w",
			err,
		)
	}
	if err := result.Uncertainty.Validate(); err != nil {
		return fmt.Errorf(
			"Weather Context uncertainty result is invalid: %w",
			err,
		)
	}

	trajectoryID := strings.TrimSpace(
		result.Weather.TrajectoryID,
	)
	if trajectoryID == "" ||
		strings.TrimSpace(result.Alignment.TrajectoryID) != trajectoryID ||
		strings.TrimSpace(result.Encounter.TrajectoryID) != trajectoryID ||
		strings.TrimSpace(result.Uncertainty.TrajectoryID) != trajectoryID {
		return fmt.Errorf(
			"Weather Context production trajectory identifiers are inconsistent",
		)
	}

	asOfTime := result.Weather.AsOfTime.UTC()
	if !result.Trust.AsOfTime.UTC().Equal(asOfTime) ||
		!result.Alignment.AsOfTime.UTC().Equal(asOfTime) ||
		!result.Encounter.AsOfTime.UTC().Equal(asOfTime) ||
		!result.Uncertainty.AsOfTime.UTC().Equal(asOfTime) {
		return fmt.Errorf(
			"Weather Context production as-of times are inconsistent",
		)
	}

	for _, generatedAt := range []time.Time{
		result.Weather.GeneratedAt,
		result.Alignment.GeneratedAt,
		result.Encounter.GeneratedAt,
		result.Uncertainty.GeneratedAt,
	} {
		if !generatedAt.UTC().Equal(
			result.GeneratedAt.UTC(),
		) {
			return fmt.Errorf(
				"Weather Context production generated-at times are inconsistent",
			)
		}
	}

	return nil
}

func (
	service *Service,
) Get(
	ctx context.Context,
	request Request,
) (Result, error) {
	if service == nil ||
		service.trajectoryReader == nil ||
		service.weatherSnapshotReader == nil ||
		service.projectionReader == nil ||
		service.now == nil {
		return Result{}, ErrServiceUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	normalizedRequest, err := normalizeRequest(
		request,
	)
	if err != nil {
		return Result{}, err
	}

	loadedTrajectory, err := service.trajectoryReader.
		GetTrajectoryByID(
			ctx,
			normalizedRequest.TrajectoryID,
		)
	if err != nil {
		return Result{}, classifyDependencyError(
			"load Weather Context trajectory",
			err,
		)
	}

	boundedTrajectory, err := trajectoryAt(
		loadedTrajectory,
		normalizedRequest.TrajectoryID,
		normalizedRequest.AsOfTime,
	)
	if err != nil {
		return Result{}, err
	}
	latestPoint := boundedTrajectory.Points[len(boundedTrajectory.Points)-1]

	snapshot, err := service.weatherSnapshotReader.
		GetLatestSnapshot(
			ctx,
			WeatherSnapshotRequest{
				Latitude:  latestPoint.Latitude,
				Longitude: latestPoint.Longitude,
				AsOfTime:  normalizedRequest.AsOfTime,
			},
		)
	if err != nil {
		return Result{}, classifyDependencyError(
			"load Weather Context weather snapshot",
			err,
		)
	}

	projection, err := service.projectionReader.
		GetProjection(
			ctx,
			ProjectionRequest{
				TrajectoryID: normalizedRequest.
					TrajectoryID,
				AsOfTime: normalizedRequest.
					AsOfTime,
				RequestedDuration: normalizedRequest.
					RequestedDuration,
			},
		)
	if err != nil {
		return Result{}, classifyDependencyError(
			"load Weather Context projection",
			err,
		)
	}
	if err := projection.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrProjectionInvalid,
			err,
		)
	}
	if strings.TrimSpace(
		projection.Projection.TrajectoryID,
	) != normalizedRequest.TrajectoryID ||
		!projection.Projection.Horizon.AsOfTime.UTC().Equal(
			normalizedRequest.AsOfTime,
		) {
		return Result{}, fmt.Errorf(
			"%w: projection identity does not match request",
			ErrProjectionInvalid,
		)
	}

	generatedAt := service.now().UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(normalizedRequest.AsOfTime) ||
		generatedAt.Before(projection.GeneratedAt.UTC()) {
		return Result{}, ErrGeneratedAtInvalid
	}

	weather, err := weatheradapter.MapOpenMeteoCurrentSnapshot(
		weatheradapter.Request{
			TrajectoryID: normalizedRequest.TrajectoryID,
			AsOfTime:     normalizedRequest.AsOfTime,
			GeneratedAt:  generatedAt,
			Snapshot:     snapshot,
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"map Weather Context snapshot: %w",
			err,
		)
	}

	trust, err := weathertrust.Evaluate(
		weather,
		service.trustPolicy,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"evaluate Weather Context trust: %w",
			err,
		)
	}

	alignment, err := weatheralignment.Align(
		weatheralignment.Request{
			Trajectory:  boundedTrajectory,
			Weather:     weather,
			Trust:       trust,
			Policy:      service.alignmentPolicy,
			GeneratedAt: generatedAt,
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"align Weather Context evidence: %w",
			err,
		)
	}

	encounter, err := weatherencounter.Build(
		weatherencounter.Request{
			Weather:     weather,
			Alignment:   alignment,
			Policy:      service.encounterPolicy,
			GeneratedAt: generatedAt,
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"build Weather Context encounter: %w",
			err,
		)
	}

	uncertainty, err := weatheruncertainty.Apply(
		weatheruncertainty.Request{
			Projection:  projection.Projection,
			Trust:       trust,
			Encounter:   encounter,
			Policy:      service.uncertaintyPolicy,
			GeneratedAt: generatedAt,
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"apply Weather Context uncertainty: %w",
			err,
		)
	}

	result := Result{
		Version:     Version,
		Weather:     weather.Clone(),
		Trust:       trust.Clone(),
		Alignment:   alignment.Clone(),
		Encounter:   encounter.Clone(),
		Uncertainty: uncertainty.Clone(),
		InputFingerprint: productionFingerprint(
			normalizedRequest,
			weather,
			trust,
			alignment,
			encounter,
			uncertainty,
		),
		GeneratedAt: generatedAt,
	}
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrResultInvalid,
			err,
		)
	}

	return result.Clone(), nil
}

func normalizeRequest(
	request Request,
) (Request, error) {
	normalized := request
	normalized.TrajectoryID = strings.TrimSpace(
		request.TrajectoryID,
	)
	normalized.AsOfTime = request.AsOfTime.UTC()

	if normalized.TrajectoryID == "" ||
		normalized.AsOfTime.IsZero() ||
		normalized.RequestedDuration <= 0 {
		return Request{}, ErrInvalidRequest
	}

	return normalized, nil
}

func trajectoryAt(
	item trajectory.FlightTrajectory,
	expectedID string,
	asOfTime time.Time,
) (trajectory.FlightTrajectory, error) {
	if strings.TrimSpace(item.ID) != expectedID {
		return trajectory.FlightTrajectory{}, fmt.Errorf(
			"%w: loaded trajectory identifier does not match request",
			ErrTrajectoryInvalid,
		)
	}

	points := make(
		[]trajectory.TrackPoint4D,
		0,
		len(item.Points),
	)
	for _, point := range item.Points {
		if point.ObservedAt.IsZero() {
			return trajectory.FlightTrajectory{}, fmt.Errorf(
				"%w: trajectory point time is required",
				ErrTrajectoryInvalid,
			)
		}
		if point.ObservedAt.UTC().After(asOfTime) {
			continue
		}
		copied := point
		copied.ObservedAt = point.ObservedAt.UTC()
		points = append(points, copied)
	}
	if len(points) == 0 {
		return trajectory.FlightTrajectory{}, ErrTrajectoryNotFound
	}

	sort.SliceStable(
		points,
		func(left int, right int) bool {
			leftTime := points[left].ObservedAt.UTC()
			rightTime := points[right].ObservedAt.UTC()
			if leftTime.Equal(rightTime) {
				return strings.TrimSpace(points[left].ID) <
					strings.TrimSpace(points[right].ID)
			}
			return leftTime.Before(rightTime)
		},
	)

	bounded := item
	bounded.Points = points
	bounded.PointCount = len(points)
	bounded.StartTime = points[0].ObservedAt.UTC()
	bounded.EndTime = points[len(points)-1].ObservedAt.UTC()
	bounded.DurationSeconds = int64(
		bounded.EndTime.Sub(bounded.StartTime) /
			time.Second,
	)
	bounded.Segments = nil
	bounded.SegmentCount = 0
	bounded.CoverageGaps = nil
	bounded.CoverageGapCount = 0
	if bounded.UpdatedAt.IsZero() ||
		bounded.UpdatedAt.After(asOfTime) {
		bounded.UpdatedAt = asOfTime
	}

	return bounded, nil
}

func classifyDependencyError(
	operation string,
	err error,
) error {
	if err == nil {
		return nil
	}

	for _, sentinel := range []error{
		ErrTrajectoryNotFound,
		ErrWeatherNotFound,
		ErrProjectionNotFound,
		ErrServiceUnavailable,
		ErrInvalidRequest,
		context.Canceled,
		context.DeadlineExceeded,
	} {
		if errors.Is(err, sentinel) {
			return err
		}
	}

	return fmt.Errorf(
		"%s: %w",
		operation,
		err,
	)
}

func productionFingerprint(
	request Request,
	weather weathercontract.Result,
	trust weathertrust.Result,
	alignment weatheralignment.Result,
	encounter weatherencounter.Result,
	uncertainty weatheruncertainty.Result,
) string {
	hasher := sha256.New()
	parts := []string{
		FingerprintVersion,
		request.TrajectoryID,
		request.AsOfTime.UTC().Format(time.RFC3339Nano),
		strconv.FormatInt(
			int64(request.RequestedDuration),
			10,
		),
		weather.Provenance.InputFingerprint,
		trust.InputFingerprint,
		alignment.InputFingerprint,
		encounter.InputFingerprint,
		uncertainty.InputFingerprint,
	}

	for _, part := range parts {
		_, _ = hasher.Write([]byte(part))
		_, _ = hasher.Write([]byte{0})
	}

	return "sha256:" +
		hex.EncodeToString(hasher.Sum(nil))
}
