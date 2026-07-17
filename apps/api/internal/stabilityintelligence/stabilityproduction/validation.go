package stabilityproduction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/confidencepropagation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/failureexplanation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecastanalysis"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecaststability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/scopeenforcement"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/unknownintervention"
)

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version ||
		strings.TrimSpace(
			result.TrajectoryID,
		) == "" ||
		len(result.AsOfTimes) <
			MinimumAsOfTimeCount ||
		len(result.AsOfTimes) >
			MaximumAsOfTimeCount ||
		len(result.Projections) !=
			len(result.AsOfTimes) ||
		len(result.ForecastVersions) !=
			len(result.AsOfTimes) ||
		len(result.Transitions) !=
			len(result.AsOfTimes)-1 ||
		result.GeneratedAt.IsZero() ||
		!fingerprintPattern.MatchString(
			result.InputFingerprint,
		) {
		return fmt.Errorf(
			"production Stability Intelligence identity is invalid",
		)
	}

	var previous time.Time
	for index, asOfTime := range result.AsOfTimes {
		if asOfTime.IsZero() ||
			(!previous.IsZero() &&
				!asOfTime.After(previous)) ||
			!result.Projections[index].
				Projection.Horizon.
				AsOfTime.Equal(asOfTime) {
			return fmt.Errorf(
				"production Stability Intelligence as-of sequence is invalid",
			)
		}
		previous = asOfTime

		if err := result.Projections[index].
			Validate(); err != nil {
			return fmt.Errorf(
				"production projection %d is invalid: %w",
				index,
				err,
			)
		}
		if err := forecaststability.
			ValidateVersionRecord(
				result.ForecastVersions[index],
				forecaststability.
					DefaultVersionPolicy(),
			); err != nil {
			return fmt.Errorf(
				"forecast version %d is invalid: %w",
				index,
				err,
			)
		}
		if result.ForecastVersions[index].
			TrajectoryID !=
			result.TrajectoryID {
			return fmt.Errorf(
				"forecast version trajectory identity mismatch",
			)
		}
	}

	for index, transition := range result.Transitions {
		if err := forecaststability.
			ValidateStabilityResult(
				transition,
				forecaststability.
					DefaultStabilityPolicy(),
			); err != nil {
			return fmt.Errorf(
				"stability transition %d is invalid: %w",
				index,
				err,
			)
		}
		if transition.BaselineVersionID !=
			result.ForecastVersions[index].
				VersionID ||
			transition.CandidateVersionID !=
				result.ForecastVersions[index+1].
					VersionID {
			return fmt.Errorf(
				"stability transition lineage mismatch",
			)
		}
	}

	if err := forecastanalysis.ValidateResult(
		result.ForecastAnalysis,
		forecastanalysis.DefaultPolicy(),
	); err != nil {
		return fmt.Errorf(
			"forecast analysis is invalid: %w",
			err,
		)
	}
	if err := confidencepropagation.ValidateResult(
		result.PropagatedConfidence,
		confidencepropagation.DefaultPolicy(),
	); err != nil {
		return fmt.Errorf(
			"propagated confidence is invalid: %w",
			err,
		)
	}
	if err := failureexplanation.ValidateResult(
		result.FailureExplanation,
		failureexplanation.DefaultPolicy(),
	); err != nil {
		return fmt.Errorf(
			"failure explanation is invalid: %w",
			err,
		)
	}
	if err := unknownintervention.ValidateResult(
		result.UnknownIntervention,
		unknownintervention.DefaultPolicy(),
	); err != nil {
		return fmt.Errorf(
			"unknown intervention result is invalid: %w",
			err,
		)
	}
	if err := scopeenforcement.ValidateResult(
		result.ScopeEnforcement,
		scopeenforcement.DefaultPolicy(),
	); err != nil {
		return fmt.Errorf(
			"scope enforcement result is invalid: %w",
			err,
		)
	}

	if result.ForecastAnalysis.TrajectoryID !=
		result.TrajectoryID ||
		result.ForecastAnalysis.Metrics.
			VersionCount !=
			len(result.ForecastVersions) ||
		len(result.ForecastAnalysis.
			Transitions) !=
			len(result.Transitions) ||
		result.PropagatedConfidence.
			TargetNodeID !=
			"stability_intelligence_output" ||
		result.FailureExplanation.
			SubjectID !=
			result.TrajectoryID ||
		result.UnknownIntervention.
			SubjectID !=
			result.TrajectoryID ||
		result.ScopeEnforcement.
			SubjectID !=
			result.TrajectoryID ||
		result.ScopeEnforcement.Decision ==
			scopeenforcement.DecisionBlocked {
		return fmt.Errorf(
			"production Stability Intelligence composition is invalid",
		)
	}

	if !sortedUniqueNonEmpty(
		result.ScopeGuards,
	) {
		return fmt.Errorf(
			"production Stability Intelligence scope guards are invalid",
		)
	}
	requiredGuard := false
	for _, guard := range result.ScopeGuards {
		if guard == ScopeGuardResearchOnly {
			requiredGuard = true
			break
		}
	}
	if !requiredGuard {
		return fmt.Errorf(
			"production Stability Intelligence scope guard is absent",
		)
	}

	if result.InputFingerprint !=
		inputFingerprint(result) {
		return fmt.Errorf(
			"production Stability Intelligence fingerprint mismatch",
		)
	}

	return nil
}

func inputFingerprint(
	result Result,
) string {
	versionIDs := make(
		[]string,
		0,
		len(result.ForecastVersions),
	)
	projectionFingerprints := make(
		[]string,
		0,
		len(result.Projections),
	)
	for _, version := range result.ForecastVersions {
		versionIDs = append(
			versionIDs,
			version.VersionID,
		)
	}
	for _, projection := range result.Projections {
		projectionFingerprints = append(
			projectionFingerprints,
			projection.InputFingerprint,
		)
	}

	payload := struct {
		Version                 string
		TrajectoryID            string
		AsOfTimes               []time.Time
		VersionIDs              []string
		ProjectionFingerprints  []string
		AnalysisFingerprint     string
		ConfidenceFingerprint   string
		FailureFingerprint      string
		InterventionFingerprint string
		ScopeFingerprint        string
		ScopeGuards             []string
	}{
		Version:                Version,
		TrajectoryID:           result.TrajectoryID,
		AsOfTimes:              append([]time.Time(nil), result.AsOfTimes...),
		VersionIDs:             versionIDs,
		ProjectionFingerprints: projectionFingerprints,
		AnalysisFingerprint: result.
			ForecastAnalysis.
			Provenance.InputFingerprint,
		ConfidenceFingerprint: result.
			PropagatedConfidence.
			Provenance.InputFingerprint,
		FailureFingerprint: result.
			FailureExplanation.
			Provenance.InputFingerprint,
		InterventionFingerprint: result.
			UnknownIntervention.
			Provenance.InputFingerprint,
		ScopeFingerprint: result.
			ScopeEnforcement.
			Provenance.InputFingerprint,
		ScopeGuards: append(
			[]string(nil),
			result.ScopeGuards...,
		),
	}

	encoded, _ := json.Marshal(payload)
	digest := sha256.Sum256(encoded)
	return "sha256:" +
		hex.EncodeToString(digest[:])
}

func sortedUniqueNonEmpty(
	values []string,
) bool {
	if len(values) == 0 {
		return false
	}
	copy := append(
		[]string(nil),
		values...,
	)
	sort.Strings(copy)
	for index, value := range values {
		if strings.TrimSpace(value) == "" ||
			value != copy[index] ||
			(index > 0 &&
				values[index-1] == value) {
			return false
		}
	}
	return true
}
