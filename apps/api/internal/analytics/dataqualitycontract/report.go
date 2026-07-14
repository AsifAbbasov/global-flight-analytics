package dataqualitycontract

import (
	"fmt"
	"strings"
	"time"
)

func NewReport(
	provenance Provenance,
	freshness Freshness,
	samplingDensity SamplingDensity,
	permissions AnalyticsPermissions,
	missingFields []string,
	warnings []Notice,
	limitations []Notice,
	evaluatedAt time.Time,
) (Report, error) {
	result := Report{
		ContractVersion: ContractVersion,
		Provenance:      provenance,
		Freshness:       freshness,
		SamplingDensity: samplingDensity,
		Permissions:     permissions.Clone(),
		MissingFields:   append([]string(nil), missingFields...),
		Warnings:        cloneNotices(warnings),
		Limitations:     cloneNotices(limitations),
		EvaluatedAt:     evaluatedAt.UTC(),
	}
	if err := result.Validate(); err != nil {
		return Report{}, err
	}
	return result, nil
}

func (value Report) Validate() error {
	if value.ContractVersion != ContractVersion {
		return fmt.Errorf("%w: %q", ErrContractVersionInvalid, value.ContractVersion)
	}
	if value.EvaluatedAt.IsZero() {
		return ErrEvaluatedAtRequired
	}
	if err := value.Provenance.Validate(); err != nil {
		return fmt.Errorf("validate provenance: %w", err)
	}
	if err := value.Freshness.Validate(); err != nil {
		return fmt.Errorf("validate freshness: %w", err)
	}
	if err := value.SamplingDensity.Validate(); err != nil {
		return fmt.Errorf("validate sampling density: %w", err)
	}
	if err := value.Permissions.Validate(); err != nil {
		return fmt.Errorf("validate analytics permissions: %w", err)
	}
	if !value.Freshness.EvaluatedAt.Equal(value.EvaluatedAt) {
		return fmt.Errorf(
			"%w: report=%s freshness=%s",
			ErrEvaluatedAtMismatch,
			value.EvaluatedAt.Format(timeFormat),
			value.Freshness.EvaluatedAt.Format(timeFormat),
		)
	}
	for index, field := range value.MissingFields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("missing field at index %d is blank", index)
		}
	}
	if err := validateNotices("warning", value.Warnings); err != nil {
		return err
	}
	if err := validateNotices("limitation", value.Limitations); err != nil {
		return err
	}
	return nil
}

func (value Report) Clone() Report {
	value.Permissions = value.Permissions.Clone()
	value.MissingFields = append([]string(nil), value.MissingFields...)
	value.Warnings = cloneNotices(value.Warnings)
	value.Limitations = cloneNotices(value.Limitations)
	return value
}

func cloneNotices(values []Notice) []Notice {
	return append([]Notice(nil), values...)
}

func validateNotices(kind string, values []Notice) error {
	for index, value := range values {
		if strings.TrimSpace(value.Code) == "" || strings.TrimSpace(value.Message) == "" {
			return fmt.Errorf("%s at index %d requires code and message", kind, index)
		}
	}
	return nil
}
