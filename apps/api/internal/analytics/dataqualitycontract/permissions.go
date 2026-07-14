package dataqualitycontract

import (
	"fmt"
	"strings"
)

func AllowedPermission() Permission {
	return Permission{Allowed: true, Reasons: []string{}}
}

func DeniedPermission(reasons ...string) (Permission, error) {
	result := Permission{
		Allowed: false,
		Reasons: append([]string(nil), reasons...),
	}
	if err := result.Validate(); err != nil {
		return Permission{}, err
	}
	return result, nil
}

func (value Permission) Validate() error {
	if !value.Allowed && len(value.Reasons) == 0 {
		return ErrPermissionReasonRequired
	}
	for index, reason := range value.Reasons {
		if strings.TrimSpace(reason) == "" {
			return fmt.Errorf("%w: index=%d", ErrPermissionReasonInvalid, index)
		}
	}
	return nil
}

func (value Permission) Clone() Permission {
	value.Reasons = append([]string(nil), value.Reasons...)
	return value
}

func (value AnalyticsPermissions) Validate() error {
	items := []struct {
		name       string
		permission Permission
	}{
		{"route_inference", value.RouteInference},
		{"phase_detection", value.PhaseDetection},
		{"historical_analytics", value.HistoricalAnalytics},
		{"historical_similarity", value.HistoricalSimilarity},
		{"projection", value.Projection},
	}
	for _, item := range items {
		if err := item.permission.Validate(); err != nil {
			return fmt.Errorf("validate %s permission: %w", item.name, err)
		}
	}
	return nil
}

func (value AnalyticsPermissions) Clone() AnalyticsPermissions {
	value.RouteInference = value.RouteInference.Clone()
	value.PhaseDetection = value.PhaseDetection.Clone()
	value.HistoricalAnalytics = value.HistoricalAnalytics.Clone()
	value.HistoricalSimilarity = value.HistoricalSimilarity.Clone()
	value.Projection = value.Projection.Clone()
	return value
}
