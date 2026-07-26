package extractorcomposition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
)

const fingerprintIdentityPrefix = "sha256:"

func resolveProcessingIdentity(
	config Config,
) (ProcessingIdentity, string, error) {
	policyVersion := strings.TrimSpace(
		config.aircraftNotFoundPolicyVersion,
	)
	if policyVersion == "" {
		return ProcessingIdentity{},
			"",
			ErrAircraftNotFoundPolicyVersionRequired
	}

	identity := ProcessingIdentity{
		Versions:                      CurrentVersions(),
		GeographicCellPrecision:       config.geographicCellPrecision,
		AircraftPositiveCacheTTL:      config.aircraftPositiveCacheTTL,
		AircraftNegativeCacheTTL:      config.aircraftNegativeCacheTTL,
		AircraftNotFoundPolicyVersion: policyVersion,
		AircraftMetadataSourceName:    aircraftprovider.MetadataSourceName,
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
