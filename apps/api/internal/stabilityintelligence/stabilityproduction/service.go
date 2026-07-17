package stabilityproduction

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/confidencepropagation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/failureexplanation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecastanalysis"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecaststability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/scopeenforcement"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/unknownintervention"
)

type Config struct {
	ProjectionReader ProjectionReader
	Now              func() time.Time
}

type Service struct {
	projectionReader ProjectionReader
	now              func() time.Time
}

func New(config Config) (*Service, error) {
	if config.ProjectionReader == nil {
		return nil, ErrProjectionReaderRequired
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		projectionReader: config.ProjectionReader,
		now:              now,
	}, nil
}

func (
	service *Service,
) Get(
	ctx context.Context,
	request Request,
) (Result, error) {
	if service == nil ||
		service.projectionReader == nil {
		return Result{}, ErrServiceUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	normalized, validationTime, err :=
		normalizeRequest(
			request,
			service.now().UTC(),
		)
	if err != nil {
		return Result{}, err
	}

	projections := make(
		[]projectionproduction.Result,
		0,
		len(normalized.AsOfTimes),
	)
	for _, asOfTime := range normalized.AsOfTimes {
		projection, loadErr :=
			service.projectionReader.ReadProjection(
				ctx,
				ProjectionRequest{
					TrajectoryID: normalized.
						TrajectoryID,
					AsOfTime: asOfTime,
					RequestedDuration: normalized.
						RequestedDuration,
				},
			)
		if loadErr != nil {
			return Result{},
				classifyProjectionLoadError(
					loadErr,
				)
		}
		if validateErr := projection.Validate(); validateErr != nil {
			return Result{},
				fmt.Errorf(
					"%w: production projection contract: %v",
					ErrProjectionLoadFailed,
					validateErr,
				)
		}
		projections = append(
			projections,
			projection.Clone(),
		)
	}

	versions, evaluationTime, err :=
		buildForecastVersions(
			projections,
			validationTime,
		)
	if err != nil {
		return Result{}, err
	}

	analysis, err :=
		forecastanalysis.AnalyzeForecastHistory(
			forecastanalysis.Request{
				Versions:    versions,
				EvaluatedAt: evaluationTime,
			},
			forecastanalysis.DefaultPolicy(),
			forecaststability.DefaultStabilityPolicy(),
		)
	if err != nil {
		return Result{},
			fmt.Errorf(
				"analyze production forecast history: %w",
				err,
			)
	}

	propagatedConfidence, err :=
		buildPropagatedConfidence(
			projections,
			analysis,
			evaluationTime,
		)
	if err != nil {
		return Result{}, err
	}

	failureResult, err :=
		buildFailureExplanation(
			normalized.TrajectoryID,
			analysis,
			propagatedConfidence,
			evaluationTime,
		)
	if err != nil {
		return Result{}, err
	}

	interventionResult, err :=
		buildUnknownInterventionDecision(
			normalized.TrajectoryID,
			projections,
			analysis,
			propagatedConfidence,
			evaluationTime,
		)
	if err != nil {
		return Result{}, err
	}

	scopeResult, guards, err :=
		buildScopeEnforcement(
			normalized.TrajectoryID,
			projections,
			analysis,
			failureResult,
			interventionResult,
			evaluationTime,
		)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Version:              Version,
		TrajectoryID:         normalized.TrajectoryID,
		AsOfTimes:            append([]time.Time(nil), normalized.AsOfTimes...),
		Projections:          projections,
		ForecastVersions:     versions,
		Transitions:          cloneTransitions(analysis.Transitions),
		ForecastAnalysis:     analysis.Clone(),
		PropagatedConfidence: propagatedConfidence.Clone(),
		FailureExplanation:   failureResult.Clone(),
		UnknownIntervention:  interventionResult.Clone(),
		ScopeEnforcement:     scopeResult.Clone(),
		ScopeGuards:          guards,
		GeneratedAt:          evaluationTime,
	}
	result.InputFingerprint =
		inputFingerprint(result)

	if err := result.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"validate production Stability Intelligence result: %w",
				err,
			)
	}

	return result.Clone(), nil
}

func normalizeRequest(
	request Request,
	now time.Time,
) (Request, time.Time, error) {
	trajectoryID := strings.TrimSpace(
		request.TrajectoryID,
	)
	if trajectoryID == "" ||
		request.RequestedDuration <= 0 ||
		now.IsZero() ||
		len(request.AsOfTimes) <
			MinimumAsOfTimeCount ||
		len(request.AsOfTimes) >
			MaximumAsOfTimeCount {
		return Request{},
			time.Time{},
			ErrInvalidRequest
	}

	asOfTimes := make(
		[]time.Time,
		0,
		len(request.AsOfTimes),
	)
	var previous time.Time
	for _, value := range request.AsOfTimes {
		asOfTime := value.UTC()
		if asOfTime.IsZero() ||
			asOfTime.After(now) ||
			(!previous.IsZero() &&
				!asOfTime.After(previous)) {
			return Request{},
				time.Time{},
				ErrInvalidRequest
		}
		asOfTimes = append(
			asOfTimes,
			asOfTime,
		)
		previous = asOfTime
	}

	return Request{
			TrajectoryID:      trajectoryID,
			AsOfTimes:         asOfTimes,
			RequestedDuration: request.RequestedDuration,
		},
		now,
		nil
}

func classifyProjectionLoadError(
	err error,
) error {
	switch {
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return err
	case errors.Is(err, ErrTrajectoryNotFound):
		return ErrTrajectoryNotFound
	case errors.Is(err, ErrServiceUnavailable):
		return ErrServiceUnavailable
	case errors.Is(err, ErrInvalidRequest):
		return ErrInvalidRequest
	default:
		return fmt.Errorf(
			"%w: %v",
			ErrProjectionLoadFailed,
			err,
		)
	}
}

func buildForecastVersions(
	projections []projectionproduction.Result,
	evaluationTime time.Time,
) (
	[]forecaststability.ForecastVersionRecord,
	time.Time,
	error,
) {
	versionPolicy :=
		forecaststability.DefaultVersionPolicy()
	versions := make(
		[]forecaststability.ForecastVersionRecord,
		0,
		len(projections),
	)

	var previous *forecaststability.ForecastVersionRecord
	for index, productionResult := range projections {
		registeredAt :=
			productionResult.GeneratedAt.UTC().
				Add(
					time.Duration(index+1) *
						time.Nanosecond,
				)
		if previous != nil &&
			!registeredAt.After(
				previous.CreatedAt,
			) {
			registeredAt =
				previous.CreatedAt.Add(
					time.Nanosecond,
				)
		}

		registration, err :=
			forecaststability.RegisterVersion(
				forecaststability.ForecastVersionRequest{
					Projection: productionResult.
						Projection.Clone(),
					PolicyVersion: strings.Join(
						[]string{
							productionResult.Version,
							string(
								productionResult.
									Strategy,
							),
						},
						":",
					),
					ImplementationVersion: strings.Join(
						[]string{
							productionResult.
								Projection.
								Method.Name,
							productionResult.
								Projection.
								Method.Version,
						},
						":",
					),
					Previous:     previous,
					RegisteredAt: registeredAt,
				},
				versionPolicy,
			)
		if err != nil {
			return nil,
				time.Time{},
				fmt.Errorf(
					"register production forecast version %d: %w",
					index+1,
					err,
				)
		}

		record := registration.Record.Clone()
		versions = append(
			versions,
			record,
		)
		previous = &record

		if evaluationTime.Before(
			record.CreatedAt,
		) {
			evaluationTime =
				record.CreatedAt
		}
	}

	evaluationTime =
		evaluationTime.UTC().
			Add(time.Second)

	return versions,
		evaluationTime,
		nil
}

func buildPropagatedConfidence(
	projections []projectionproduction.Result,
	analysis forecastanalysis.Result,
	evaluatedAt time.Time,
) (
	confidencepropagation.Result,
	error,
) {
	nodes := make(
		[]confidencepropagation.Node,
		0,
		len(projections)+2,
	)
	historyDependencies := make(
		[]confidencepropagation.Dependency,
		0,
		len(projections),
	)
	weight := 1.0 / float64(
		len(projections),
	)

	for index, projection := range projections {
		nodeID := fmt.Sprintf(
			"projection_%02d",
			index+1,
		)
		nodes = append(
			nodes,
			confidencepropagation.Node{
				ID: nodeID,
				Label: "Projection confidence at " +
					projection.Projection.
						Horizon.AsOfTime.
						Format(time.RFC3339Nano),
				Kind: confidencepropagation.
					NodeKindEvidence,
				Classification: confidencepropagation.
					ClassificationEstimated,
				LocalScore: projection.Projection.
					Confidence.Score,
				SourceFingerprint: projection.
					InputFingerprint,
			},
		)
		historyDependencies = append(
			historyDependencies,
			confidencepropagation.Dependency{
				NodeID:   nodeID,
				Weight:   weight,
				Required: true,
			},
		)
	}

	nodes = append(
		nodes,
		confidencepropagation.Node{
			ID:             "forecast_history",
			Label:          "Forecast history stability",
			Kind:           confidencepropagation.NodeKindDecision,
			Classification: confidencepropagation.ClassificationDerived,
			LocalScore:     analysis.Confidence.Score,
			Dependencies:   historyDependencies,
			SourceFingerprint: analysis.
				Provenance.InputFingerprint,
		},
	)

	latestNodeID := fmt.Sprintf(
		"projection_%02d",
		len(projections),
	)
	nodes = append(
		nodes,
		confidencepropagation.Node{
			ID:             "stability_intelligence_output",
			Label:          "Stability Intelligence output",
			Kind:           confidencepropagation.NodeKindOutput,
			Classification: confidencepropagation.ClassificationDerived,
			LocalScore:     analysis.Confidence.Score,
			Dependencies: []confidencepropagation.Dependency{
				{
					NodeID:   "forecast_history",
					Weight:   0.60,
					Required: true,
				},
				{
					NodeID:   latestNodeID,
					Weight:   0.40,
					Required: true,
				},
			},
			SourceFingerprint: analysis.
				Provenance.InputFingerprint,
		},
	)

	result, err :=
		confidencepropagation.Propagate(
			confidencepropagation.Request{
				TargetNodeID: "stability_intelligence_output",
				Nodes:        nodes,
				EvaluatedAt:  evaluatedAt,
			},
			confidencepropagation.DefaultPolicy(),
		)
	if err != nil {
		return confidencepropagation.Result{},
			fmt.Errorf(
				"propagate production confidence: %w",
				err,
			)
	}

	return result.Clone(), nil
}

func buildFailureExplanation(
	trajectoryID string,
	analysis forecastanalysis.Result,
	confidence confidencepropagation.Result,
	evaluatedAt time.Time,
) (
	failureexplanation.Result,
	error,
) {
	signals :=
		failureexplanation.
			SignalsFromForecastAnalysis(
				analysis,
			)
	signals = append(
		signals,
		failureexplanation.
			SignalsFromConfidencePropagation(
				confidence,
			)...,
	)

	if len(analysis.Transitions) > 0 {
		latest := analysis.Transitions[len(analysis.Transitions)-1]
		signals = append(
			signals,
			failureexplanation.
				SignalsFromDecisionStability(
					latest,
				)...,
		)
	}

	result, err :=
		failureexplanation.Explain(
			failureexplanation.Request{
				SubjectID:   trajectoryID,
				SubjectType: "forecast_history",
				Signals:     signals,
				EvaluatedAt: evaluatedAt,
			},
			failureexplanation.DefaultPolicy(),
		)
	if err != nil {
		return failureexplanation.Result{},
			fmt.Errorf(
				"build production failure explanation: %w",
				err,
			)
	}

	return result.Clone(), nil
}

func buildUnknownInterventionDecision(
	trajectoryID string,
	projections []projectionproduction.Result,
	analysis forecastanalysis.Result,
	confidence confidencepropagation.Result,
	evaluatedAt time.Time,
) (
	unknownintervention.Result,
	error,
) {
	latest := projections[len(projections)-1]
	completeness :=
		analysis.Metrics.
			ComparableTransitionShare
	if completeness < 0 {
		completeness = 0
	}
	if completeness > 1 {
		completeness = 1
	}

	evidence := []unknownintervention.Evidence{
		{
			ID:       "forecast_history",
			Label:    "Forecast history stability",
			Class:    unknownintervention.EvidenceDerived,
			Score:    analysis.Confidence.Score,
			Required: true,
			Source:   forecastanalysis.Version,
			Fingerprint: analysis.
				Provenance.InputFingerprint,
		},
		{
			ID:       "propagated_confidence",
			Label:    "Propagated confidence",
			Class:    unknownintervention.EvidenceDerived,
			Score:    confidence.Score,
			Required: true,
			Source:   confidencepropagation.Version,
			Fingerprint: confidence.
				Provenance.InputFingerprint,
		},
		{
			ID:       "latest_projection",
			Label:    "Latest estimated projection",
			Class:    unknownintervention.EvidenceEstimated,
			Score:    latest.Projection.Confidence.Score,
			Required: true,
			Source:   latest.Version,
			Fingerprint: latest.
				InputFingerprint,
			Limitation: "The projection is estimated and does not reveal pilot intent, air traffic control instruction, or exact operational cause.",
		},
	}

	result, err :=
		unknownintervention.Evaluate(
			unknownintervention.Request{
				SubjectID: trajectoryID,
				ClaimKind: unknownintervention.
					ClaimKindContextualAssociation,
				ClaimText: fmt.Sprintf(
					"Forecast history is %s with trend %s.",
					analysis.Health,
					analysis.Trend,
				),
				Evidence:             evidence,
				EvidenceCompleteness: completeness,
				EvaluatedAt:          evaluatedAt,
			},
			unknownintervention.DefaultPolicy(),
		)
	if err != nil {
		return unknownintervention.Result{},
			fmt.Errorf(
				"evaluate production intervention boundary: %w",
				err,
			)
	}

	return result.Clone(), nil
}

func buildScopeEnforcement(
	trajectoryID string,
	projections []projectionproduction.Result,
	analysis forecastanalysis.Result,
	failure failureexplanation.Result,
	intervention unknownintervention.Result,
	evaluatedAt time.Time,
) (
	scopeenforcement.Result,
	[]string,
	error,
) {
	guards := []string{
		string(
			forecaststability.ScopeGuardResearchOnly,
		),
		forecastanalysis.ScopeGuardResearchOnly,
		confidencepropagation.ScopeGuardResearchOnly,
		failureexplanation.ScopeGuardResearchOnly,
		unknownintervention.ScopeGuardResearchOnly,
		scopeenforcement.ScopeGuardResearchOnly,
		ScopeGuardResearchOnly,
	}
	for _, projection := range projections {
		guards = append(
			guards,
			string(
				projection.
					Projection.
					ScopeGuard,
			),
		)
	}
	sort.Strings(guards)
	guards = uniqueStrings(guards)

	claims := []scopeenforcement.Claim{
		{
			Code: "forecast_stability_history",
			Text: fmt.Sprintf(
				"Forecast history health is %s and trend is %s.",
				analysis.Health,
				analysis.Trend,
			),
			Capability: forecastanalysis.Version,
			Scope: scopeenforcement.
				ScopeResearchAnalysis,
			Strength: scopeenforcement.
				StrengthAnalytical,
			SourceGuard: forecastanalysis.
				ScopeGuardResearchOnly,
		},
		{
			Code: "failure_condition_explanation",
			Text: fmt.Sprintf(
				"Primary bounded failure condition is %s.",
				failure.PrimaryCode,
			),
			Capability: failureexplanation.Version,
			Scope: scopeenforcement.
				ScopeResearchAnalysis,
			Strength: scopeenforcement.
				StrengthAnalytical,
			SourceGuard: failureexplanation.
				ScopeGuardResearchOnly,
		},
		{
			Code: "unknown_intervention_boundary",
			Text: fmt.Sprintf(
				"Unknown intervention decision is %s and no pilot intent, air traffic control instruction, or exact cause is claimed.",
				intervention.Decision,
			),
			Capability: unknownintervention.Version,
			Scope: scopeenforcement.
				ScopeResearchAnalysis,
			Strength: scopeenforcement.
				StrengthDescriptive,
			SourceGuard: unknownintervention.
				ScopeGuardResearchOnly,
		},
	}

	result, err :=
		scopeenforcement.Enforce(
			scopeenforcement.Request{
				SubjectID:      trajectoryID,
				DeclaredGuards: guards,
				Claims:         claims,
				EvaluatedAt:    evaluatedAt,
			},
			scopeenforcement.DefaultPolicy(),
		)
	if err != nil {
		return scopeenforcement.Result{},
			nil,
			fmt.Errorf(
				"enforce production publication scope: %w",
				err,
			)
	}

	return result.Clone(),
		guards,
		nil
}

func cloneTransitions(
	items []forecaststability.StabilityResult,
) []forecaststability.StabilityResult {
	result := make(
		[]forecaststability.StabilityResult,
		0,
		len(items),
	)
	for _, item := range items {
		result = append(
			result,
			item.Clone(),
		)
	}
	return result
}

func uniqueStrings(
	values []string,
) []string {
	result := make(
		[]string,
		0,
		len(values),
	)
	for _, value := range values {
		if value == "" {
			continue
		}
		if len(result) == 0 ||
			result[len(result)-1] != value {
			result = append(
				result,
				value,
			)
		}
	}
	return result
}
