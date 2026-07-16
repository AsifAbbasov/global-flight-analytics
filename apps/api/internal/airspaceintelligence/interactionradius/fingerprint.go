package interactionradius

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const FingerprintVersionV1 = "interaction-radius-input-fingerprint-v1"

func inputFingerprint(request Request, policy Policy) string {
	hasher := sha256.New()
	parts := []string{
		FingerprintVersionV1,
		strings.ToUpper(strings.TrimSpace(request.RegionCode)),
		strings.TrimSpace(request.NodeID),
		strings.ToUpper(strings.TrimSpace(request.ICAO24)),
		strings.ToUpper(strings.TrimSpace(request.Callsign)),
		formatFloat(request.VelocityMetersPerSecond),
		formatFloat(request.VerticalRateMetersPerSecond),
		formatOptionalFloat(request.AltitudeMeters),
		string(request.AltitudeReference),
		strconv.FormatBool(request.OnGround),
		request.ObservedAt.UTC().Format(time.RFC3339Nano),
		request.AsOfTime.UTC().Format(time.RFC3339Nano),
		request.GeneratedAt.UTC().Format(time.RFC3339Nano),
		strings.TrimSpace(request.SourceName),
		formatFloat(request.QualityScore),
		policy.Version,
		formatFloat(policy.MinimumHorizontalRadiusKilometers),
		formatFloat(policy.BaseHorizontalRadiusKilometers),
		formatFloat(policy.MaximumHorizontalRadiusKilometers),
		policy.HorizontalLookaheadDuration.String(),
		formatFloat(policy.QualityUncertaintyFraction),
		formatFloat(policy.MinimumVerticalRadiusMeters),
		formatFloat(policy.BaseVerticalRadiusMeters),
		formatFloat(policy.MaximumVerticalRadiusMeters),
		policy.VerticalLookaheadDuration.String(),
		policy.MaximumObservationAge.String(),
		policy.MaximumPairTimeDifference.String(),
	}
	for _, part := range parts {
		_, _ = hasher.Write([]byte(part))
		_, _ = hasher.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil))
}

func formatOptionalFloat(value *float64) string {
	if value == nil {
		return "nil"
	}
	return formatFloat(*value)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func validFingerprint(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, prefix))
	return err == nil && len(decoded) == sha256.Size
}

func fingerprintIssue(path string) string {
	return fmt.Sprintf("%s must be a sha256 fingerprint", path)
}
