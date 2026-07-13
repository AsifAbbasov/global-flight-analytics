package confidencereport

import (
	"fmt"
	"math"
	"sort"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
)

type Evaluator struct {
	config Config
}

func New(
	config Config,
) (*Evaluator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate confidence report config: %w",
			err,
		)
	}

	return &Evaluator{
		config: config,
	}, nil
}

func NewDefault() *Evaluator {
	evaluator, err := New(
		DefaultConfig(),
	)
	if err != nil {
		panic(
			fmt.Sprintf(
				"default confidence report config is invalid: %v",
				err,
			),
		)
	}

	return evaluator
}

func (
	evaluator *Evaluator,
) Evaluate(
	request Request,
) (Report, error) {
	if err := request.Validate(); err != nil {
		return Report{}, fmt.Errorf(
			"validate confidence report request: %w",
			err,
		)
	}

	factors := append(
		[]Factor(nil),
		request.Factors...,
	)

	sort.SliceStable(
		factors,
		func(
			left int,
			right int,
		) bool {
			leftRank := factorKindRank(
				factors[left].Kind,
			)
			rightRank := factorKindRank(
				factors[right].Kind,
			)

			if leftRank != rightRank {
				return leftRank < rightRank
			}

			return factors[left].Code <
				factors[right].Code
		},
	)

	totalEvidenceWeight := 0.0
	evidenceWeightedScore := 0.0
	rawPenaltyScore := 0.0

	for _, factor := range factors {
		switch factor.Kind {
		case FactorKindEvidence:
			totalEvidenceWeight += factor.Weight
			evidenceWeightedScore +=
				factor.Weight *
					factor.Value

		case FactorKindPenalty:
			rawPenaltyScore +=
				factor.Weight *
					factor.Value
		}
	}

	baseScore := round(
		clampUnit(
			evidenceWeightedScore/
				totalEvidenceWeight,
		),
		evaluator.config.DecimalPrecision,
	)

	penaltyScore := round(
		math.Min(
			rawPenaltyScore,
			evaluator.config.MaximumPenalty,
		),
		evaluator.config.DecimalPrecision,
	)

	penaltyScale := 0.0
	if rawPenaltyScore > 0 {
		penaltyScale =
			penaltyScore /
				rawPenaltyScore
	}

	contributions := make(
		[]Contribution,
		0,
		len(factors),
	)
	reasons := make(
		[]analyticalresult.Notice,
		0,
		len(factors),
	)

	for _, factor := range factors {
		impact := 0.0
		reasonPrefix := "confidence_evidence_"

		switch factor.Kind {
		case FactorKindEvidence:
			impact =
				factor.Weight /
					totalEvidenceWeight *
					factor.Value

		case FactorKindPenalty:
			impact =
				-factor.Weight *
					factor.Value *
					penaltyScale
			reasonPrefix =
				"confidence_penalty_"
		}

		impact = round(
			impact,
			evaluator.config.DecimalPrecision,
		)

		contributions = append(
			contributions,
			Contribution{
				Code:    factor.Code,
				Kind:    factor.Kind,
				Weight:  factor.Weight,
				Value:   factor.Value,
				Impact:  impact,
				Message: factor.Message,
			},
		)

		reasons = append(
			reasons,
			analyticalresult.Notice{
				Code: reasonPrefix +
					factor.Code,
				Message: factor.Message,
			},
		)
	}

	score := round(
		clampUnit(
			baseScore-
				penaltyScore,
		),
		evaluator.config.DecimalPrecision,
	)

	report := Report{
		BaseScore:    baseScore,
		PenaltyScore: penaltyScore,
		Score:        score,
		Level: evaluator.confidenceLevel(
			score,
		),
		Factors: contributions,
		Reasons: reasons,
		Warnings: cloneAndSortNotices(
			request.Warnings,
		),
		Limitations: cloneAndSortNotices(
			request.Limitations,
		),
		EvaluatedAt: request.EvaluatedAt.UTC(),
	}

	if err := report.AnalyticalConfidence().
		Validate(); err != nil {
		return Report{}, fmt.Errorf(
			"validate generated analytical confidence: %w",
			err,
		)
	}

	return report, nil
}

func (
	evaluator *Evaluator,
) confidenceLevel(
	score float64,
) analyticalresult.ConfidenceLevel {
	switch {
	case score == 0:
		return analyticalresult.
			ConfidenceLevelNone

	case score >= evaluator.config.HighThreshold:
		return analyticalresult.
			ConfidenceLevelHigh

	case score >= evaluator.config.MediumThreshold:
		return analyticalresult.
			ConfidenceLevelMedium

	default:
		return analyticalresult.
			ConfidenceLevelLow
	}
}

func factorKindRank(
	kind FactorKind,
) int {
	switch kind {
	case FactorKindEvidence:
		return 0

	case FactorKindPenalty:
		return 1

	default:
		return 2
	}
}

func cloneAndSortNotices(
	notices []analyticalresult.Notice,
) []analyticalresult.Notice {
	result := append(
		[]analyticalresult.Notice(nil),
		notices...,
	)

	sort.SliceStable(
		result,
		func(
			left int,
			right int,
		) bool {
			return result[left].Code <
				result[right].Code
		},
	)

	return result
}

func round(
	value float64,
	precision int,
) float64 {
	if precision == 0 {
		return math.Round(value)
	}

	scale := math.Pow10(
		precision,
	)

	return math.Round(
		value*scale,
	) / scale
}
