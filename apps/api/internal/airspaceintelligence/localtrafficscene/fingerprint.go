package localtrafficscene

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
)

type fingerprintEnvelope struct {
	RegionCode          string
	RegionBounds        Bounds
	AsOfTime            string
	ScenePolicyVersion  string
	RadiusPolicyVersion string
	Aircraft            []fingerprintAircraft
	Excluded            []fingerprintExclusion
}

type fingerprintAircraft struct {
	NodeID                      string
	ICAO24                      string
	Callsign                    string
	Latitude                    float64
	Longitude                   float64
	AltitudeMeters              *float64
	AltitudeReference           string
	VelocityMetersPerSecond     float64
	HeadingDegrees              float64
	VerticalRateMetersPerSecond float64
	ObservedAt                  string
	SourceName                  string
	QualityScore                float64
	RadiusStatus                string
	HorizontalRadiusKilometers  float64
	VerticalRadiusMeters        float64
	RadiusFingerprint           string
}

type fingerprintExclusion struct {
	NodeID     string
	ICAO24     string
	ObservedAt string
	Reason     string
	SourceName string
}

func inputFingerprint(
	request Request,
	policy Policy,
	radiusPolicy interactionradius.Policy,
	result Result,
) string {
	envelope := fingerprintEnvelope{
		RegionCode:          request.RegionCode,
		RegionBounds:        request.RegionBounds,
		AsOfTime:            request.AsOfTime.UTC().Format(timeLayout),
		ScenePolicyVersion:  policy.Version,
		RadiusPolicyVersion: radiusPolicy.Version,
		Aircraft:            make([]fingerprintAircraft, 0, len(result.Aircraft)),
		Excluded:            make([]fingerprintExclusion, 0, len(result.ExcludedObservations)),
	}
	for _, item := range result.Aircraft {
		envelope.Aircraft = append(envelope.Aircraft, fingerprintAircraft{
			NodeID:                      item.NodeID,
			ICAO24:                      item.ICAO24,
			Callsign:                    item.Callsign,
			Latitude:                    item.Latitude,
			Longitude:                   item.Longitude,
			AltitudeMeters:              cloneFloat64(item.AltitudeMeters),
			AltitudeReference:           string(item.AltitudeReference),
			VelocityMetersPerSecond:     item.VelocityMetersPerSecond,
			HeadingDegrees:              item.HeadingDegrees,
			VerticalRateMetersPerSecond: item.VerticalRateMetersPerSecond,
			ObservedAt:                  item.ObservedAt.UTC().Format(timeLayout),
			SourceName:                  item.SourceName,
			QualityScore:                item.QualityScore,
			RadiusStatus:                string(item.RadiusDecision.Status),
			HorizontalRadiusKilometers:  item.RadiusDecision.HorizontalRadiusKilometers,
			VerticalRadiusMeters:        item.RadiusDecision.VerticalRadiusMeters,
			RadiusFingerprint:           item.RadiusDecision.Provenance.InputFingerprint,
		})
	}
	for _, item := range result.ExcludedObservations {
		envelope.Excluded = append(envelope.Excluded, fingerprintExclusion{
			NodeID:     item.NodeID,
			ICAO24:     item.ICAO24,
			ObservedAt: item.ObservedAt.UTC().Format(timeLayout),
			Reason:     string(item.Reason),
			SourceName: item.SourceName,
		})
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}
