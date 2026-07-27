package featurestore

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/google/uuid"
)

var snapshotFingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func requireStoreContext(ctx context.Context) error {
	if ctx == nil {
		return ErrContextRequired
	}
	return ctx.Err()
}

func canonicalTrajectoryID(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrTrajectoryIDRequired
	}
	parsed, err := uuid.Parse(normalized)
	if err != nil {
		return "", ErrInvalidTrajectoryID
	}
	return strings.ToLower(parsed.String()), nil
}

func validateInputFingerprint(value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrInputFingerprintRequired
	}
	if !snapshotFingerprintPattern.MatchString(value) {
		return ErrInvalidInputFingerprint
	}
	return nil
}

func validateWritableFeatures(
	features flightfeatures.FlightFeatures,
) error {
	if err := validateStorableFeatures(features); err != nil {
		return err
	}
	return validateWritableValidationProof(features)
}

func validateWritableValidationProof(
	features flightfeatures.FlightFeatures,
) error {
	report := features.ValidationReport
	if report.AuditState != flightfeatures.ValidationAuditStateComplete ||
		strings.TrimSpace(report.ValidatorVersion) == "" ||
		report.ValidatedAt.IsZero() ||
		report.Status != features.Quality.Status {
		return ErrValidationProofRequired
	}
	return nil
}

func validateFiniteFeatureValues(
	features flightfeatures.FlightFeatures,
) error {
	return validateFiniteValue(
		reflect.ValueOf(features),
		"features",
	)
}

func validateFiniteValue(value reflect.Value, path string) error {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return validateFiniteValue(value.Elem(), path)
	}

	switch value.Kind() {
	case reflect.Float32, reflect.Float64:
		candidate := value.Float()
		if math.IsNaN(candidate) || math.IsInf(candidate, 0) {
			return fmt.Errorf(
				"%w: %s",
				ErrNonFiniteFeatureValue,
				path,
			)
		}
	case reflect.Struct:
		valueType := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := valueType.Field(index)
			if field.PkgPath != "" {
				continue
			}
			if err := validateFiniteValue(
				value.Field(index),
				path+"."+field.Name,
			); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if err := validateFiniteValue(
				value.Index(index),
				fmt.Sprintf("%s[%d]", path, index),
			); err != nil {
				return err
			}
		}
	}
	return nil
}
