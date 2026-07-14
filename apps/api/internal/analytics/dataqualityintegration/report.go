package dataqualityintegration

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	DefaultExpectedObservationInterval = 10 * time.Second
	DefaultStaleAfter                  = 5 * time.Minute

	TrajectoryTransformation = "trajectory_analytics_quality_assessment"
	UnavailableIngestionRun  = "unavailable"
)

const (
	NoticeCodeFutureObservationsExcluded = "future_observations_excluded"
	NoticeCodeSourceNameUnavailable      = "source_name_unavailable"

	LimitationCodeIngestionRunUnavailable = "ingestion_run_provenance_unavailable"
	LimitationCodeReceivedAtDerived       = "received_at_derived"
	LimitationCodeSimilarityPending       = "historical_similarity_not_implemented"
)

type TrajectoryReportRequest struct {
	Trajectories []trajectory.FlightTrajectory
	EvaluatedAt  time.Time

	ExpectedObservationInterval time.Duration
	StaleAfter                  time.Duration
}

type trajectoryEvidence struct {
	observationTimes []time.Time
	sourceNames      []string
	sourceRecordTime time.Time
	receivedAt       time.Time
	inputFingerprint string
	missingFields    []string
	warnings         []dataqualitycontract.Notice
	limitations      []dataqualitycontract.Notice
}

func BuildTrajectoryReport(
	request TrajectoryReportRequest,
) (*dataqualitycontract.Report, error) {
	if len(request.Trajectories) == 0 {
		return nil, nil
	}
	if request.EvaluatedAt.IsZero() {
		return nil, ErrEvaluatedAtRequired
	}

	expectedInterval := request.ExpectedObservationInterval
	if expectedInterval == 0 {
		expectedInterval = DefaultExpectedObservationInterval
	}
	staleAfter := request.StaleAfter
	if staleAfter == 0 {
		staleAfter = DefaultStaleAfter
	}

	evaluatedAt := request.EvaluatedAt.UTC()
	evidence, err := collectTrajectoryEvidence(
		request.Trajectories,
		evaluatedAt,
	)
	if err != nil {
		return nil, err
	}

	if !evaluatedAt.After(evidence.sourceRecordTime) {
		evaluatedAt = evidence.sourceRecordTime.Add(time.Nanosecond)
	}

	freshness, err := dataqualitycontract.EvaluateFreshness(
		dataqualitycontract.FreshnessInput{
			ObservedAt:       evidence.sourceRecordTime,
			EvaluatedAt:      evaluatedAt,
			ExpectedInterval: expectedInterval,
			StaleAfter:       staleAfter,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("evaluate trajectory freshness: %w", err)
	}

	windowStart := evidence.observationTimes[0].UTC().Truncate(
		expectedInterval,
	)
	samplingDensity, err := dataqualitycontract.EvaluateSamplingDensity(
		dataqualitycontract.SamplingDensityInput{
			WindowStart:      windowStart,
			WindowEnd:        evaluatedAt,
			ExpectedInterval: expectedInterval,
			ObservationTimes: evidence.observationTimes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"evaluate trajectory sampling density: %w",
			err,
		)
	}

	phaseDetection, phaseLimitations, err :=
		evaluatePhaseDetection(
			request.Trajectories,
		)
	if err != nil {
		return nil, err
	}

	permissions, err := buildAnalyticsPermissions(
		request.Trajectories,
		evaluatedAt,
		phaseDetection,
	)
	if err != nil {
		return nil, err
	}

	evidence.limitations = append(
		evidence.limitations,
		phaseLimitations...,
	)

	report, err := dataqualitycontract.NewReport(
		dataqualitycontract.Provenance{
			SourceName: strings.Join(
				evidence.sourceNames,
				",",
			),
			SourceRecordTime: evidence.sourceRecordTime,
			ReceivedAt:       evidence.receivedAt,
			IngestionRunID:   UnavailableIngestionRun,
			Transformation:   TrajectoryTransformation,
			AlgorithmVersion: dataqualitycontract.ContractVersion,
			InputFingerprint: evidence.inputFingerprint,
		},
		freshness,
		samplingDensity,
		permissions,
		evidence.missingFields,
		evidence.warnings,
		evidence.limitations,
		evaluatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("build trajectory data quality report: %w", err)
	}

	return &report, nil
}

func collectTrajectoryEvidence(
	items []trajectory.FlightTrajectory,
	evaluatedAt time.Time,
) (trajectoryEvidence, error) {
	sourceNames := make(map[string]struct{})
	observationTimes := make([]time.Time, 0)
	futureObservationCount := 0
	missingSourceName := false
	hasCallsign := false
	hasIdentity := false
	hasTrackPoints := false
	receivedAt := time.Time{}

	for _, item := range items {
		sourceName := strings.TrimSpace(item.SourceName)
		if sourceName == "" {
			sourceName = "unknown"
			missingSourceName = true
		}
		sourceNames[sourceName] = struct{}{}

		if strings.TrimSpace(item.Callsign) != "" {
			hasCallsign = true
		}
		if strings.TrimSpace(item.IdentityKey) != "" {
			hasIdentity = true
		}

		receivedAt = latestTime(
			receivedAt,
			item.CreatedAt,
			item.UpdatedAt,
		)

		itemObservationCount := 0
		for _, point := range item.Points {
			if point.ObservedAt.IsZero() {
				continue
			}
			if point.ObservedAt.After(evaluatedAt) {
				futureObservationCount++
				continue
			}

			observationTimes = append(
				observationTimes,
				point.ObservedAt.UTC(),
			)
			itemObservationCount++
			hasTrackPoints = true
		}

		if itemObservationCount == 0 && !item.EndTime.IsZero() {
			if item.EndTime.After(evaluatedAt) {
				futureObservationCount++
			} else {
				observationTimes = append(
					observationTimes,
					item.EndTime.UTC(),
				)
			}
		}
	}

	if len(observationTimes) == 0 {
		return trajectoryEvidence{}, ErrNoUsableObservationTimes
	}

	sort.SliceStable(
		observationTimes,
		func(left int, right int) bool {
			return observationTimes[left].Before(
				observationTimes[right],
			)
		},
	)

	sourceRecordTime := observationTimes[len(observationTimes)-1]
	limitations := []dataqualitycontract.Notice{
		{
			Code:    LimitationCodeIngestionRunUnavailable,
			Message: "The trajectory read model does not expose a single authoritative ingestion run identifier.",
		},
		{
			Code:    LimitationCodeSimilarityPending,
			Message: "Historical-similarity permission remains denied until the similarity policy is implemented.",
		},
	}

	if receivedAt.IsZero() ||
		receivedAt.Before(sourceRecordTime) ||
		receivedAt.After(evaluatedAt) {
		receivedAt = evaluatedAt
		limitations = append(
			limitations,
			dataqualitycontract.Notice{
				Code:    LimitationCodeReceivedAtDerived,
				Message: "Received-at provenance was derived from the analytical evaluation time because a reliable retained receipt timestamp was unavailable.",
			},
		)
	}

	warnings := make([]dataqualitycontract.Notice, 0, 2)
	if futureObservationCount > 0 {
		warnings = append(
			warnings,
			dataqualitycontract.Notice{
				Code: NoticeCodeFutureObservationsExcluded,
				Message: fmt.Sprintf(
					"%d future-dated observations were excluded from data-quality scoring.",
					futureObservationCount,
				),
			},
		)
	}
	if missingSourceName {
		warnings = append(
			warnings,
			dataqualitycontract.Notice{
				Code:    NoticeCodeSourceNameUnavailable,
				Message: "One or more trajectories did not expose a source name and were grouped under unknown.",
			},
		)
	}

	missingFields := make([]string, 0, 3)
	if !hasCallsign {
		missingFields = append(missingFields, "callsign")
	}
	if !hasIdentity {
		missingFields = append(missingFields, "identity_key")
	}
	if !hasTrackPoints {
		missingFields = append(missingFields, "track_points")
	}
	sort.Strings(missingFields)

	names := make([]string, 0, len(sourceNames))
	for name := range sourceNames {
		names = append(names, name)
	}
	sort.Strings(names)

	return trajectoryEvidence{
		observationTimes: observationTimes,
		sourceNames:      names,
		sourceRecordTime: sourceRecordTime.UTC(),
		receivedAt:       receivedAt.UTC(),
		inputFingerprint: trajectoryFingerprint(items),
		missingFields:    missingFields,
		warnings:         warnings,
		limitations:      limitations,
	}, nil
}

func buildAnalyticsPermissions(
	items []trajectory.FlightTrajectory,
	evaluatedAt time.Time,
	phaseDetection dataqualitycontract.Permission,
) (dataqualitycontract.AnalyticsPermissions, error) {
	evaluator := trajectoryeligibility.NewDefault()

	routeInference, err := aggregateCapabilityPermission(
		evaluator,
		items,
		evaluatedAt,
		trajectoryeligibility.CapabilityRouteInference,
	)
	if err != nil {
		return dataqualitycontract.AnalyticsPermissions{}, err
	}
	historicalAnalytics, err := aggregateCapabilityPermission(
		evaluator,
		items,
		evaluatedAt,
		trajectoryeligibility.CapabilityHistoricalAggregation,
	)
	if err != nil {
		return dataqualitycontract.AnalyticsPermissions{}, err
	}
	projection, err := aggregateCapabilityPermission(
		evaluator,
		items,
		evaluatedAt,
		trajectoryeligibility.CapabilityProjection,
	)
	if err != nil {
		return dataqualitycontract.AnalyticsPermissions{}, err
	}

	historicalSimilarity, err := dataqualitycontract.DeniedPermission(
		LimitationCodeSimilarityPending,
	)
	if err != nil {
		return dataqualitycontract.AnalyticsPermissions{}, err
	}

	result := dataqualitycontract.AnalyticsPermissions{
		RouteInference:       routeInference,
		PhaseDetection:       phaseDetection,
		HistoricalAnalytics:  historicalAnalytics,
		HistoricalSimilarity: historicalSimilarity,
		Projection:           projection,
	}
	if err := result.Validate(); err != nil {
		return dataqualitycontract.AnalyticsPermissions{},
			fmt.Errorf("validate integrated analytics permissions: %w", err)
	}

	return result, nil
}

func aggregateCapabilityPermission(
	evaluator *trajectoryeligibility.Evaluator,
	items []trajectory.FlightTrajectory,
	evaluatedAt time.Time,
	capability trajectoryeligibility.Capability,
) (dataqualitycontract.Permission, error) {
	reasons := make(map[string]struct{})

	for _, item := range items {
		evaluation := evaluator.Evaluate(
			item,
			evaluatedAt,
		)
		decision, exists := evaluation.Decision(
			capability,
		)
		if !exists {
			continue
		}
		if decision.Allowed {
			return dataqualitycontract.AllowedPermission(), nil
		}

		for _, reason := range decision.Reasons {
			reasons[string(reason)] = struct{}{}
		}
	}

	if len(reasons) == 0 {
		reasons["no_eligible_trajectory"] = struct{}{}
	}

	values := make([]string, 0, len(reasons))
	for reason := range reasons {
		values = append(values, reason)
	}
	sort.Strings(values)

	permission, err := dataqualitycontract.DeniedPermission(values...)
	if err != nil {
		return dataqualitycontract.Permission{},
			fmt.Errorf(
				"build %s data-quality permission: %w",
				capability,
				err,
			)
	}

	return permission, nil
}

func trajectoryFingerprint(
	items []trajectory.FlightTrajectory,
) string {
	records := make([]string, 0)

	for _, item := range items {
		records = append(
			records,
			fmt.Sprintf(
				"trajectory|%s|%s|%s|%s|%s|%s|%d|%d|%.12f",
				strings.TrimSpace(item.ID),
				strings.TrimSpace(item.IdentityKey),
				strings.TrimSpace(item.ICAO24),
				strings.TrimSpace(item.Callsign),
				item.StartTime.UTC().Format(time.RFC3339Nano),
				item.EndTime.UTC().Format(time.RFC3339Nano),
				item.PointCount,
				item.CoverageGapCount,
				item.QualityScore,
			),
		)

		for _, point := range item.Points {
			records = append(
				records,
				fmt.Sprintf(
					"point|%s|%s|%s|%s",
					strings.TrimSpace(point.ID),
					strings.TrimSpace(point.FlightStateID),
					strings.TrimSpace(point.SourceName),
					point.ObservedAt.UTC().Format(
						time.RFC3339Nano,
					),
				),
			)
		}
	}

	sort.Strings(records)
	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return hex.EncodeToString(sum[:])
}

func latestTime(
	current time.Time,
	candidates ...time.Time,
) time.Time {
	result := current
	for _, candidate := range candidates {
		if candidate.After(result) {
			result = candidate
		}
	}
	return result
}
