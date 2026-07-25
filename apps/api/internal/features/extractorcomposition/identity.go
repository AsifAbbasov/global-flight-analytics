package extractorcomposition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
)

const fingerprintIdentityPrefix = "sha256:"

func resolveProcessingIdentity(
	config Config,
) (ProcessingIdentity, string, error) {
	policyVersion := strings.TrimSpace(
		config.AircraftNotFoundPolicyVersion,
	)
	if config.IsAircraftNotFound == nil {
		if policyVersion == "" {
			policyVersion =
				aircraftprovider.DefaultNotFoundPolicyVersion
		}
	} else if policyVersion == "" {
		return ProcessingIdentity{},
			"",
			ErrAircraftNotFoundPolicyVersionRequired
	}

	geographicCellPrecision := config.GeographicCellPrecision
	if geographicCellPrecision == 0 {
		geographicCellPrecision =
			geographicalbuilder.DefaultGeographicCellPrecision
	}

	positiveCacheTTL := config.AircraftPositiveCacheTTL
	if positiveCacheTTL == 0 {
		positiveCacheTTL =
			aircraftprovider.DefaultPositiveCacheTTL
	}

	negativeCacheTTL := config.AircraftNegativeCacheTTL
	if negativeCacheTTL == 0 {
		negativeCacheTTL =
			aircraftprovider.DefaultNegativeCacheTTL
	}

	identity := ProcessingIdentity{
		Versions:                      CurrentVersions(),
		GeographicCellPrecision:       geographicCellPrecision,
		AircraftPositiveCacheTTL:      positiveCacheTTL,
		AircraftNegativeCacheTTL:      negativeCacheTTL,
		AircraftNotFoundPolicyVersion: policyVersion,
	}

	payload, err := json.Marshal(identity)
	if err != nil {
		return ProcessingIdentity{}, "", err
	}
	sum := sha256.Sum256(payload)

	return identity,
		fingerprintIdentityPrefix + hex.EncodeToString(sum[:]),
		nil
}

func dependencyMissing(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
