package confidencereport

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
)

var (
	ErrEvaluationTimeMissing = errors.New(
		"confidence evaluation time is required",
	)
	ErrFactorsMissing = errors.New(
		"confidence factors are required",
	)
	ErrEvidenceFactorsMissing = errors.New(
		"at least one evidence confidence factor is required",
	)
	ErrFactorCodeInvalid = errors.New(
		"confidence factor code is invalid",
	)
	ErrFactorKindInvalid = errors.New(
		"confidence factor kind is invalid",
	)
	ErrFactorWeightInvalid = errors.New(
		"confidence factor weight must be finite and greater than zero",
	)
	ErrFactorValueInvalid = errors.New(
		"confidence factor value must be finite and between zero and one",
	)
	ErrFactorMessageInvalid = errors.New(
		"confidence factor message is required and must not have surrounding whitespace",
	)
	ErrDuplicateFactorCode = errors.New(
		"confidence factor codes must be unique",
	)
	ErrNoticeCodeInvalid = errors.New(
		"confidence notice code is invalid",
	)
	ErrNoticeMessageInvalid = errors.New(
		"confidence notice message is required and must not have surrounding whitespace",
	)
	ErrDuplicateNoticeCode = errors.New(
		"confidence notice codes must be unique within a collection",
	)
)

func (
	request Request,
) Validate() error {
	if request.EvaluatedAt.IsZero() {
		return ErrEvaluationTimeMissing
	}

	if len(request.Factors) == 0 {
		return ErrFactorsMissing
	}

	factorCodes := make(
		map[string]struct{},
		len(request.Factors),
	)
	evidenceCount := 0

	for index, factor := range request.Factors {
		if err := factor.Validate(); err != nil {
			return fmt.Errorf(
				"validate confidence factor at index %d: %w",
				index,
				err,
			)
		}

		if _, exists := factorCodes[factor.Code]; exists {
			return fmt.Errorf(
				"%w: %q",
				ErrDuplicateFactorCode,
				factor.Code,
			)
		}

		factorCodes[factor.Code] = struct{}{}

		if factor.Kind == FactorKindEvidence {
			evidenceCount++
		}
	}

	if evidenceCount == 0 {
		return ErrEvidenceFactorsMissing
	}

	if err := validateNotices(
		"warnings",
		request.Warnings,
	); err != nil {
		return err
	}

	if err := validateNotices(
		"limitations",
		request.Limitations,
	); err != nil {
		return err
	}

	return nil
}

func (
	factor Factor,
) Validate() error {
	if err := validateMachineCode(
		factor.Code,
		ErrFactorCodeInvalid,
	); err != nil {
		return err
	}

	if !factor.Kind.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrFactorKindInvalid,
			factor.Kind,
		)
	}

	if !isFinite(factor.Weight) ||
		factor.Weight <= 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrFactorWeightInvalid,
			factor.Weight,
		)
	}

	if !isFinite(factor.Value) ||
		factor.Value < 0 ||
		factor.Value > 1 {
		return fmt.Errorf(
			"%w: %f",
			ErrFactorValueInvalid,
			factor.Value,
		)
	}

	if strings.TrimSpace(factor.Message) == "" ||
		factor.Message != strings.TrimSpace(
			factor.Message,
		) {
		return ErrFactorMessageInvalid
	}

	return nil
}

func validateNotices(
	collectionName string,
	notices []analyticalresult.Notice,
) error {
	seen := make(
		map[string]struct{},
		len(notices),
	)

	for index, notice := range notices {
		if err := validateMachineCode(
			notice.Code,
			ErrNoticeCodeInvalid,
		); err != nil {
			return fmt.Errorf(
				"validate %s notice at index %d: %w",
				collectionName,
				index,
				err,
			)
		}

		if strings.TrimSpace(notice.Message) == "" ||
			notice.Message != strings.TrimSpace(
				notice.Message,
			) {
			return fmt.Errorf(
				"validate %s notice at index %d: %w",
				collectionName,
				index,
				ErrNoticeMessageInvalid,
			)
		}

		if _, exists := seen[notice.Code]; exists {
			return fmt.Errorf(
				"%w: collection=%s code=%q",
				ErrDuplicateNoticeCode,
				collectionName,
				notice.Code,
			)
		}

		seen[notice.Code] = struct{}{}
	}

	return nil
}

func validateMachineCode(
	value string,
	sentinel error,
) error {
	if value == "" ||
		value != strings.TrimSpace(value) {
		return fmt.Errorf(
			"%w: %q",
			sentinel,
			value,
		)
	}

	for index, character := range value {
		if index == 0 &&
			!unicode.IsLower(character) {
			return fmt.Errorf(
				"%w: %q",
				sentinel,
				value,
			)
		}

		if unicode.IsLower(character) ||
			unicode.IsDigit(character) ||
			character == '_' ||
			character == '-' ||
			character == '.' {
			continue
		}

		return fmt.Errorf(
			"%w: %q",
			sentinel,
			value,
		)
	}

	return nil
}

func clampUnit(
	value float64,
) float64 {
	return math.Max(
		0,
		math.Min(
			1,
			value,
		),
	)
}
