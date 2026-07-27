package featurestore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const snapshotPayloadVersionV1 = "flight-feature-snapshot-payload-v1"
const snapshotFingerprintPrefix = "sha256:"

type snapshotPayloadProbe struct {
	PayloadVersion string `json:"PayloadVersion"`
}

type snapshotPayloadV1 struct {
	PayloadVersion    string                    `json:"PayloadVersion"`
	OutputFingerprint string                    `json:"OutputFingerprint"`
	SchemaVersion     string                    `json:"SchemaVersion"`
	TrajectoryID      string                    `json:"TrajectoryID"`
	IdentityKey       string                    `json:"IdentityKey"`
	FlightID          string                    `json:"FlightID"`
	AircraftID        string                    `json:"AircraftID"`
	ICAO24            string                    `json:"ICAO24"`
	Callsign          string                    `json:"Callsign"`
	Window            featureWindowPayloadV1    `json:"Window"`
	ExtractedAt       time.Time                 `json:"ExtractedAt"`
	ValidationReport  validationReportPayloadV1 `json:"ValidationReport"`
	Temporal          temporalPayloadV1         `json:"Temporal"`
	Geographical      geographicalPayloadV1     `json:"Geographical"`
	Operational       operationalPayloadV1      `json:"Operational"`
	Trajectory        trajectoryPayloadV1       `json:"Trajectory"`
	Aircraft          aircraftPayloadV1         `json:"Aircraft"`
	Quality           qualityPayloadV1          `json:"Quality"`
	Provenance        provenancePayloadV1       `json:"Provenance"`
}

type featureWindowPayloadV1 struct {
	StartTime time.Time `json:"StartTime"`
	EndTime   time.Time `json:"EndTime"`
	AsOfTime  time.Time `json:"AsOfTime"`
}

type validationIssuePayloadV1 struct {
	Code     string `json:"Code"`
	Message  string `json:"Message"`
	Path     string `json:"Path"`
	Group    string `json:"Group"`
	Severity string `json:"Severity"`
}

type validationReportPayloadV1 struct {
	AuditState       string                     `json:"AuditState"`
	ValidatorVersion string                     `json:"ValidatorVersion"`
	Status           string                     `json:"Status"`
	ErrorCount       int                        `json:"ErrorCount"`
	WarningCount     int                        `json:"WarningCount"`
	Issues           []validationIssuePayloadV1 `json:"Issues"`
	ValidatedAt      time.Time                  `json:"ValidatedAt"`
}

type limitationPayloadV1 struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type evidencePayloadV1 struct {
	Status               string                `json:"Status"`
	AvailableFieldCount  int                   `json:"AvailableFieldCount"`
	TotalFieldCount      int                   `json:"TotalFieldCount"`
	SupportingPointCount int                   `json:"SupportingPointCount"`
	Limitations          []limitationPayloadV1 `json:"Limitations"`
}

type temporalPayloadV1 struct {
	Evidence            evidencePayloadV1 `json:"Evidence"`
	DurationSeconds     int64             `json:"DurationSeconds"`
	StartHourUTC        int               `json:"StartHourUTC"`
	EndHourUTC          int               `json:"EndHourUTC"`
	StartWeekday        int               `json:"StartWeekday"`
	EndWeekday          int               `json:"EndWeekday"`
	StartMinuteOfDayUTC int               `json:"StartMinuteOfDayUTC"`
	EndMinuteOfDayUTC   int               `json:"EndMinuteOfDayUTC"`
	CrossesUTCMidnight  bool              `json:"CrossesUTCMidnight"`
}

type geographicalPayloadV1 struct {
	Evidence                  evidencePayloadV1 `json:"Evidence"`
	StartLatitude             float64           `json:"StartLatitude"`
	StartLongitude            float64           `json:"StartLongitude"`
	EndLatitude               float64           `json:"EndLatitude"`
	EndLongitude              float64           `json:"EndLongitude"`
	MinimumLatitude           float64           `json:"MinimumLatitude"`
	MaximumLatitude           float64           `json:"MaximumLatitude"`
	MinimumLongitude          float64           `json:"MinimumLongitude"`
	MaximumLongitude          float64           `json:"MaximumLongitude"`
	LatitudeSpanDegrees       float64           `json:"LatitudeSpanDegrees"`
	LongitudeSpanDegrees      float64           `json:"LongitudeSpanDegrees"`
	GreatCircleDistanceKM     float64           `json:"GreatCircleDistanceKM"`
	ObservedPathDistanceKM    float64           `json:"ObservedPathDistanceKM"`
	MaximumDisplacementKM     float64           `json:"MaximumDisplacementKM"`
	CrossesAntimeridian       bool              `json:"CrossesAntimeridian"`
	UniqueGeographicCellCount int               `json:"UniqueGeographicCellCount"`
	GeographicCellPrecision   int               `json:"GeographicCellPrecision"`
}

type operationalPayloadV1 struct {
	Evidence                       evidencePayloadV1 `json:"Evidence"`
	MinimumAltitudeM               float64           `json:"MinimumAltitudeM"`
	MaximumAltitudeM               float64           `json:"MaximumAltitudeM"`
	MeanAltitudeM                  float64           `json:"MeanAltitudeM"`
	AltitudeRangeM                 float64           `json:"AltitudeRangeM"`
	MeanVelocityMPS                float64           `json:"MeanVelocityMPS"`
	MaximumVelocityMPS             float64           `json:"MaximumVelocityMPS"`
	MeanAbsoluteVerticalRateMPS    float64           `json:"MeanAbsoluteVerticalRateMPS"`
	MaximumAbsoluteVerticalRateMPS float64           `json:"MaximumAbsoluteVerticalRateMPS"`
	HeadingChangeDegrees           float64           `json:"HeadingChangeDegrees"`
	GroundObservationShare         float64           `json:"GroundObservationShare"`
	AirborneObservationShare       float64           `json:"AirborneObservationShare"`
}

type trajectoryPayloadV1 struct {
	Evidence                    evidencePayloadV1 `json:"Evidence"`
	PointCount                  int               `json:"PointCount"`
	SegmentCount                int               `json:"SegmentCount"`
	CoverageGapCount            int               `json:"CoverageGapCount"`
	TrajectoryQualityScore      float64           `json:"TrajectoryQualityScore"`
	ObservedSegmentCount        int               `json:"ObservedSegmentCount"`
	InterpolatedSegmentCount    int               `json:"InterpolatedSegmentCount"`
	EstimatedSegmentCount       int               `json:"EstimatedSegmentCount"`
	InvalidSegmentCount         int               `json:"InvalidSegmentCount"`
	ObservedSegmentShare        float64           `json:"ObservedSegmentShare"`
	InterpolatedSegmentShare    float64           `json:"InterpolatedSegmentShare"`
	EstimatedSegmentShare       float64           `json:"EstimatedSegmentShare"`
	InvalidSegmentShare         float64           `json:"InvalidSegmentShare"`
	MeanSamplingIntervalSeconds float64           `json:"MeanSamplingIntervalSeconds"`
	MaximumSamplingGapSeconds   float64           `json:"MaximumSamplingGapSeconds"`
	CoverageRatio               float64           `json:"CoverageRatio"`
	PathEfficiencyRatio         float64           `json:"PathEfficiencyRatio"`
}

type aircraftPayloadV1 struct {
	Evidence          evidencePayloadV1 `json:"Evidence"`
	Registration      string            `json:"Registration"`
	Manufacturer      string            `json:"Manufacturer"`
	Model             string            `json:"Model"`
	AircraftType      string            `json:"AircraftType"`
	Airline           string            `json:"Airline"`
	Country           string            `json:"Country"`
	MetadataUpdatedAt time.Time         `json:"MetadataUpdatedAt"`
}

type qualityPayloadV1 struct {
	Status                string                `json:"Status"`
	CompletenessScore     float64               `json:"CompletenessScore"`
	OptionalCoverageScore float64               `json:"OptionalCoverageScore"`
	InputQualityScore     float64               `json:"InputQualityScore"`
	SupportingPointCount  int                   `json:"SupportingPointCount"`
	Limitations           []limitationPayloadV1 `json:"Limitations"`
}

type processingVersionsPayloadV1 struct {
	Composition         string `json:"Composition"`
	Extractor           string `json:"Extractor"`
	AircraftProvider    string `json:"AircraftProvider"`
	TemporalBuilder     string `json:"TemporalBuilder"`
	GeographicalBuilder string `json:"GeographicalBuilder"`
	OperationalBuilder  string `json:"OperationalBuilder"`
	TrajectoryBuilder   string `json:"TrajectoryBuilder"`
}

type processingIdentityPayloadV1 struct {
	Versions                      processingVersionsPayloadV1 `json:"Versions"`
	GeographicCellPrecision       int                         `json:"GeographicCellPrecision"`
	AircraftEnrichmentMode        string                      `json:"AircraftEnrichmentMode"`
	AircraftCacheMode             string                      `json:"AircraftCacheMode"`
	AircraftPositiveCacheTTL      int64                       `json:"AircraftPositiveCacheTTL"`
	AircraftNegativeCacheTTL      int64                       `json:"AircraftNegativeCacheTTL"`
	AircraftNotFoundPolicyVersion string                      `json:"AircraftNotFoundPolicyVersion"`
	AircraftMetadataSourceName    string                      `json:"AircraftMetadataSourceName"`
}

type provenancePayloadV1 struct {
	ProcessingVersion               string                      `json:"ProcessingVersion"`
	ExtractorVersion                string                      `json:"ExtractorVersion"`
	InputFingerprint                string                      `json:"InputFingerprint"`
	ProcessingIdentityFingerprint   string                      `json:"ProcessingIdentityFingerprint"`
	ProcessingIdentity              processingIdentityPayloadV1 `json:"ProcessingIdentity"`
	TrajectoryCreatedAt             time.Time                   `json:"TrajectoryCreatedAt"`
	TrajectoryUpdatedAt             time.Time                   `json:"TrajectoryUpdatedAt"`
	AircraftMetadataSourceName      string                      `json:"AircraftMetadataSourceName"`
	AircraftMetadataProviderVersion string                      `json:"AircraftMetadataProviderVersion"`
	AircraftMetadataRetrievedAt     time.Time                   `json:"AircraftMetadataRetrievedAt"`
	SourceNames                     []string                    `json:"SourceNames"`
}

func encodeSnapshotPayload(
	features flightfeatures.FlightFeatures,
) ([]byte, string, error) {
	payload := toSnapshotPayloadV1(features)
	fingerprint, err := fingerprintSnapshotPayload(payload)
	if err != nil {
		return nil, "", err
	}
	payload.OutputFingerprint = fingerprint

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("marshal versioned feature snapshot payload: %w", err)
	}
	return encoded, fingerprint, nil
}

func decodeSnapshotPayload(
	encoded []byte,
) (flightfeatures.FlightFeatures, string, error) {
	var probe snapshotPayloadProbe
	if err := json.Unmarshal(encoded, &probe); err != nil {
		return flightfeatures.FlightFeatures{}, "", &DatabaseError{
			Operation: "decode snapshot payload version",
			Err:       err,
		}
	}
	if strings.TrimSpace(probe.PayloadVersion) == "" {
		var legacy flightfeatures.FlightFeatures
		if err := json.Unmarshal(encoded, &legacy); err != nil {
			return flightfeatures.FlightFeatures{}, "", &DatabaseError{
				Operation: "decode legacy snapshot payload",
				Err:       err,
			}
		}
		normalizeValidationReport(&legacy)
		fingerprint, err := fingerprintSnapshotOutput(legacy)
		return legacy, fingerprint, err
	}
	if probe.PayloadVersion != snapshotPayloadVersionV1 {
		return flightfeatures.FlightFeatures{}, "", &CorruptSnapshotError{Field: "payload_version"}
	}

	if err := validateSnapshotPayloadShape(encoded); err != nil {
		return flightfeatures.FlightFeatures{}, "", err
	}

	var payload snapshotPayloadV1
	if err := strictDecodeJSON(encoded, &payload); err != nil {
		return flightfeatures.FlightFeatures{}, "", err
	}
	features := fromSnapshotPayloadV1(payload)
	expected, err := fingerprintSnapshotOutput(features)
	if err != nil {
		return flightfeatures.FlightFeatures{}, "", err
	}
	if payload.OutputFingerprint != expected {
		return flightfeatures.FlightFeatures{}, "", &CorruptSnapshotError{Field: "output_fingerprint"}
	}
	return features, expected, nil
}

var timeType = reflect.TypeOf(time.Time{})

func validateSnapshotPayloadShape(encoded []byte) error {
	return validateJSONShape(
		json.RawMessage(encoded),
		reflect.TypeOf(snapshotPayloadV1{}),
		"features_json",
	)
}

func validateJSONShape(
	raw json.RawMessage,
	expected reflect.Type,
	path string,
) error {
	if expected == timeType {
		return nil
	}

	switch expected.Kind() {
	case reflect.Struct:
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			return &DatabaseError{
				Operation: "decode snapshot payload shape",
				Err:       err,
			}
		}
		for index := 0; index < expected.NumField(); index++ {
			field := expected.Field(index)
			if field.PkgPath != "" {
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" {
				name = field.Name
			}
			value, exists := object[name]
			if !exists {
				return &CorruptSnapshotError{Field: path + "." + name}
			}
			if err := validateJSONShape(
				value,
				field.Type,
				path+"."+name,
			); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil
		}
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return &DatabaseError{
				Operation: "decode snapshot payload shape",
				Err:       err,
			}
		}
		for index, item := range items {
			if err := validateJSONShape(
				item,
				expected.Elem(),
				fmt.Sprintf("%s[%d]", path, index),
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func strictDecodeJSON(encoded []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return &DatabaseError{Operation: "decode snapshot payload", Err: err}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return &DatabaseError{Operation: "decode snapshot payload", Err: err}
	}
	return nil
}

func fingerprintSnapshotOutput(
	features flightfeatures.FlightFeatures,
) (string, error) {
	return fingerprintSnapshotPayload(toSnapshotPayloadV1(features))
}

func fingerprintSnapshotPayload(
	payload snapshotPayloadV1,
) (string, error) {
	payload.OutputFingerprint = ""
	payload.ExtractedAt = time.Time{}
	payload.ValidationReport.ValidatedAt = time.Time{}
	payload.Provenance.AircraftMetadataRetrievedAt = time.Time{}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal feature snapshot output fingerprint: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return snapshotFingerprintPrefix + hex.EncodeToString(sum[:]), nil
}

func toSnapshotPayloadV1(
	features flightfeatures.FlightFeatures,
) snapshotPayloadV1 {
	return snapshotPayloadV1{
		PayloadVersion:   snapshotPayloadVersionV1,
		SchemaVersion:    string(features.SchemaVersion),
		TrajectoryID:     features.TrajectoryID,
		IdentityKey:      features.IdentityKey,
		FlightID:         features.FlightID,
		AircraftID:       features.AircraftID,
		ICAO24:           features.ICAO24,
		Callsign:         features.Callsign,
		Window:           toFeatureWindowPayloadV1(features.Window),
		ExtractedAt:      features.ExtractedAt,
		ValidationReport: toValidationReportPayloadV1(features.ValidationReport),
		Temporal:         toTemporalPayloadV1(features.Temporal),
		Geographical:     toGeographicalPayloadV1(features.Geographical),
		Operational:      toOperationalPayloadV1(features.Operational),
		Trajectory:       toTrajectoryPayloadV1(features.Trajectory),
		Aircraft:         toAircraftPayloadV1(features.Aircraft),
		Quality:          toQualityPayloadV1(features.Quality),
		Provenance:       toProvenancePayloadV1(features.Provenance),
	}
}

func fromSnapshotPayloadV1(
	payload snapshotPayloadV1,
) flightfeatures.FlightFeatures {
	return flightfeatures.FlightFeatures{
		SchemaVersion:    flightfeatures.SchemaVersion(payload.SchemaVersion),
		TrajectoryID:     payload.TrajectoryID,
		IdentityKey:      payload.IdentityKey,
		FlightID:         payload.FlightID,
		AircraftID:       payload.AircraftID,
		ICAO24:           payload.ICAO24,
		Callsign:         payload.Callsign,
		Window:           fromFeatureWindowPayloadV1(payload.Window),
		ExtractedAt:      payload.ExtractedAt,
		ValidationReport: fromValidationReportPayloadV1(payload.ValidationReport),
		Temporal:         fromTemporalPayloadV1(payload.Temporal),
		Geographical:     fromGeographicalPayloadV1(payload.Geographical),
		Operational:      fromOperationalPayloadV1(payload.Operational),
		Trajectory:       fromTrajectoryPayloadV1(payload.Trajectory),
		Aircraft:         fromAircraftPayloadV1(payload.Aircraft),
		Quality:          fromQualityPayloadV1(payload.Quality),
		Provenance:       fromProvenancePayloadV1(payload.Provenance),
	}
}

func toFeatureWindowPayloadV1(value flightfeatures.FeatureWindow) featureWindowPayloadV1 {
	return featureWindowPayloadV1{StartTime: value.StartTime, EndTime: value.EndTime, AsOfTime: value.AsOfTime}
}

func fromFeatureWindowPayloadV1(value featureWindowPayloadV1) flightfeatures.FeatureWindow {
	return flightfeatures.FeatureWindow{StartTime: value.StartTime, EndTime: value.EndTime, AsOfTime: value.AsOfTime}
}

func toValidationReportPayloadV1(value flightfeatures.ValidationReport) validationReportPayloadV1 {
	var issues []validationIssuePayloadV1
	if value.Issues != nil {
		issues = make([]validationIssuePayloadV1, 0, len(value.Issues))
	}
	for _, issue := range value.Issues {
		issues = append(issues, validationIssuePayloadV1{
			Code: issue.Code, Message: issue.Message, Path: issue.Path,
			Group: string(issue.Group), Severity: string(issue.Severity),
		})
	}
	return validationReportPayloadV1{
		AuditState: string(value.AuditState), ValidatorVersion: value.ValidatorVersion,
		Status: string(value.Status), ErrorCount: value.ErrorCount,
		WarningCount: value.WarningCount, Issues: issues, ValidatedAt: value.ValidatedAt,
	}
}

func fromValidationReportPayloadV1(value validationReportPayloadV1) flightfeatures.ValidationReport {
	var issues []flightfeatures.ValidationIssue
	if value.Issues != nil {
		issues = make([]flightfeatures.ValidationIssue, 0, len(value.Issues))
	}
	for _, issue := range value.Issues {
		issues = append(issues, flightfeatures.ValidationIssue{
			Code: issue.Code, Message: issue.Message, Path: issue.Path,
			Group:    flightfeatures.FeatureGroup(issue.Group),
			Severity: flightfeatures.ValidationIssueSeverity(issue.Severity),
		})
	}
	return flightfeatures.ValidationReport{
		AuditState:       flightfeatures.ValidationAuditState(value.AuditState),
		ValidatorVersion: value.ValidatorVersion,
		Status:           flightfeatures.ValidationStatus(value.Status),
		ErrorCount:       value.ErrorCount, WarningCount: value.WarningCount,
		Issues: issues, ValidatedAt: value.ValidatedAt,
	}
}

func toLimitationsPayloadV1(values []flightfeatures.FeatureLimitation) []limitationPayloadV1 {
	if values == nil {
		return nil
	}
	result := make([]limitationPayloadV1, 0, len(values))
	for _, value := range values {
		result = append(result, limitationPayloadV1{Code: value.Code, Message: value.Message})
	}
	return result
}

func fromLimitationsPayloadV1(values []limitationPayloadV1) []flightfeatures.FeatureLimitation {
	if values == nil {
		return nil
	}
	result := make([]flightfeatures.FeatureLimitation, 0, len(values))
	for _, value := range values {
		result = append(result, flightfeatures.FeatureLimitation{Code: value.Code, Message: value.Message})
	}
	return result
}

func toEvidencePayloadV1(value flightfeatures.GroupEvidence) evidencePayloadV1 {
	return evidencePayloadV1{
		Status: string(value.Status), AvailableFieldCount: value.AvailableFieldCount,
		TotalFieldCount: value.TotalFieldCount, SupportingPointCount: value.SupportingPointCount,
		Limitations: toLimitationsPayloadV1(value.Limitations),
	}
}

func fromEvidencePayloadV1(value evidencePayloadV1) flightfeatures.GroupEvidence {
	return flightfeatures.GroupEvidence{
		Status:              flightfeatures.AvailabilityStatus(value.Status),
		AvailableFieldCount: value.AvailableFieldCount, TotalFieldCount: value.TotalFieldCount,
		SupportingPointCount: value.SupportingPointCount,
		Limitations:          fromLimitationsPayloadV1(value.Limitations),
	}
}

func toTemporalPayloadV1(value flightfeatures.TemporalFeatures) temporalPayloadV1 {
	return temporalPayloadV1{
		Evidence: toEvidencePayloadV1(value.Evidence), DurationSeconds: value.DurationSeconds,
		StartHourUTC: value.StartHourUTC, EndHourUTC: value.EndHourUTC,
		StartWeekday: int(value.StartWeekday), EndWeekday: int(value.EndWeekday),
		StartMinuteOfDayUTC: value.StartMinuteOfDayUTC, EndMinuteOfDayUTC: value.EndMinuteOfDayUTC,
		CrossesUTCMidnight: value.CrossesUTCMidnight,
	}
}

func fromTemporalPayloadV1(value temporalPayloadV1) flightfeatures.TemporalFeatures {
	return flightfeatures.TemporalFeatures{
		Evidence: fromEvidencePayloadV1(value.Evidence), DurationSeconds: value.DurationSeconds,
		StartHourUTC: value.StartHourUTC, EndHourUTC: value.EndHourUTC,
		StartWeekday: time.Weekday(value.StartWeekday), EndWeekday: time.Weekday(value.EndWeekday),
		StartMinuteOfDayUTC: value.StartMinuteOfDayUTC, EndMinuteOfDayUTC: value.EndMinuteOfDayUTC,
		CrossesUTCMidnight: value.CrossesUTCMidnight,
	}
}

func toGeographicalPayloadV1(value flightfeatures.GeographicalFeatures) geographicalPayloadV1 {
	return geographicalPayloadV1{
		Evidence:      toEvidencePayloadV1(value.Evidence),
		StartLatitude: value.StartLatitude, StartLongitude: value.StartLongitude,
		EndLatitude: value.EndLatitude, EndLongitude: value.EndLongitude,
		MinimumLatitude: value.MinimumLatitude, MaximumLatitude: value.MaximumLatitude,
		MinimumLongitude: value.MinimumLongitude, MaximumLongitude: value.MaximumLongitude,
		LatitudeSpanDegrees: value.LatitudeSpanDegrees, LongitudeSpanDegrees: value.LongitudeSpanDegrees,
		GreatCircleDistanceKM: value.GreatCircleDistanceKM, ObservedPathDistanceKM: value.ObservedPathDistanceKM,
		MaximumDisplacementKM: value.MaximumDisplacementKM, CrossesAntimeridian: value.CrossesAntimeridian,
		UniqueGeographicCellCount: value.UniqueGeographicCellCount,
		GeographicCellPrecision:   value.GeographicCellPrecision,
	}
}

func fromGeographicalPayloadV1(value geographicalPayloadV1) flightfeatures.GeographicalFeatures {
	return flightfeatures.GeographicalFeatures{
		Evidence:      fromEvidencePayloadV1(value.Evidence),
		StartLatitude: value.StartLatitude, StartLongitude: value.StartLongitude,
		EndLatitude: value.EndLatitude, EndLongitude: value.EndLongitude,
		MinimumLatitude: value.MinimumLatitude, MaximumLatitude: value.MaximumLatitude,
		MinimumLongitude: value.MinimumLongitude, MaximumLongitude: value.MaximumLongitude,
		LatitudeSpanDegrees: value.LatitudeSpanDegrees, LongitudeSpanDegrees: value.LongitudeSpanDegrees,
		GreatCircleDistanceKM: value.GreatCircleDistanceKM, ObservedPathDistanceKM: value.ObservedPathDistanceKM,
		MaximumDisplacementKM: value.MaximumDisplacementKM, CrossesAntimeridian: value.CrossesAntimeridian,
		UniqueGeographicCellCount: value.UniqueGeographicCellCount,
		GeographicCellPrecision:   value.GeographicCellPrecision,
	}
}

func toOperationalPayloadV1(value flightfeatures.OperationalFeatures) operationalPayloadV1 {
	return operationalPayloadV1{
		Evidence: toEvidencePayloadV1(value.Evidence), MinimumAltitudeM: value.MinimumAltitudeM,
		MaximumAltitudeM: value.MaximumAltitudeM, MeanAltitudeM: value.MeanAltitudeM,
		AltitudeRangeM: value.AltitudeRangeM, MeanVelocityMPS: value.MeanVelocityMPS,
		MaximumVelocityMPS:             value.MaximumVelocityMPS,
		MeanAbsoluteVerticalRateMPS:    value.MeanAbsoluteVerticalRateMPS,
		MaximumAbsoluteVerticalRateMPS: value.MaximumAbsoluteVerticalRateMPS,
		HeadingChangeDegrees:           value.HeadingChangeDegrees,
		GroundObservationShare:         value.GroundObservationShare,
		AirborneObservationShare:       value.AirborneObservationShare,
	}
}

func fromOperationalPayloadV1(value operationalPayloadV1) flightfeatures.OperationalFeatures {
	return flightfeatures.OperationalFeatures{
		Evidence: fromEvidencePayloadV1(value.Evidence), MinimumAltitudeM: value.MinimumAltitudeM,
		MaximumAltitudeM: value.MaximumAltitudeM, MeanAltitudeM: value.MeanAltitudeM,
		AltitudeRangeM: value.AltitudeRangeM, MeanVelocityMPS: value.MeanVelocityMPS,
		MaximumVelocityMPS:             value.MaximumVelocityMPS,
		MeanAbsoluteVerticalRateMPS:    value.MeanAbsoluteVerticalRateMPS,
		MaximumAbsoluteVerticalRateMPS: value.MaximumAbsoluteVerticalRateMPS,
		HeadingChangeDegrees:           value.HeadingChangeDegrees,
		GroundObservationShare:         value.GroundObservationShare,
		AirborneObservationShare:       value.AirborneObservationShare,
	}
}

func toTrajectoryPayloadV1(value flightfeatures.TrajectoryFeatures) trajectoryPayloadV1 {
	return trajectoryPayloadV1{
		Evidence: toEvidencePayloadV1(value.Evidence), PointCount: value.PointCount,
		SegmentCount: value.SegmentCount, CoverageGapCount: value.CoverageGapCount,
		TrajectoryQualityScore:      value.TrajectoryQualityScore,
		ObservedSegmentCount:        value.ObservedSegmentCount,
		InterpolatedSegmentCount:    value.InterpolatedSegmentCount,
		EstimatedSegmentCount:       value.EstimatedSegmentCount,
		InvalidSegmentCount:         value.InvalidSegmentCount,
		ObservedSegmentShare:        value.ObservedSegmentShare,
		InterpolatedSegmentShare:    value.InterpolatedSegmentShare,
		EstimatedSegmentShare:       value.EstimatedSegmentShare,
		InvalidSegmentShare:         value.InvalidSegmentShare,
		MeanSamplingIntervalSeconds: value.MeanSamplingIntervalSeconds,
		MaximumSamplingGapSeconds:   value.MaximumSamplingGapSeconds,
		CoverageRatio:               value.CoverageRatio, PathEfficiencyRatio: value.PathEfficiencyRatio,
	}
}

func fromTrajectoryPayloadV1(value trajectoryPayloadV1) flightfeatures.TrajectoryFeatures {
	return flightfeatures.TrajectoryFeatures{
		Evidence: fromEvidencePayloadV1(value.Evidence), PointCount: value.PointCount,
		SegmentCount: value.SegmentCount, CoverageGapCount: value.CoverageGapCount,
		TrajectoryQualityScore:      value.TrajectoryQualityScore,
		ObservedSegmentCount:        value.ObservedSegmentCount,
		InterpolatedSegmentCount:    value.InterpolatedSegmentCount,
		EstimatedSegmentCount:       value.EstimatedSegmentCount,
		InvalidSegmentCount:         value.InvalidSegmentCount,
		ObservedSegmentShare:        value.ObservedSegmentShare,
		InterpolatedSegmentShare:    value.InterpolatedSegmentShare,
		EstimatedSegmentShare:       value.EstimatedSegmentShare,
		InvalidSegmentShare:         value.InvalidSegmentShare,
		MeanSamplingIntervalSeconds: value.MeanSamplingIntervalSeconds,
		MaximumSamplingGapSeconds:   value.MaximumSamplingGapSeconds,
		CoverageRatio:               value.CoverageRatio, PathEfficiencyRatio: value.PathEfficiencyRatio,
	}
}

func toAircraftPayloadV1(value flightfeatures.AircraftFeatures) aircraftPayloadV1 {
	return aircraftPayloadV1{
		Evidence: toEvidencePayloadV1(value.Evidence), Registration: value.Registration,
		Manufacturer: value.Manufacturer, Model: value.Model, AircraftType: value.AircraftType,
		Airline: value.Airline, Country: value.Country, MetadataUpdatedAt: value.MetadataUpdatedAt,
	}
}

func fromAircraftPayloadV1(value aircraftPayloadV1) flightfeatures.AircraftFeatures {
	return flightfeatures.AircraftFeatures{
		Evidence: fromEvidencePayloadV1(value.Evidence), Registration: value.Registration,
		Manufacturer: value.Manufacturer, Model: value.Model, AircraftType: value.AircraftType,
		Airline: value.Airline, Country: value.Country, MetadataUpdatedAt: value.MetadataUpdatedAt,
	}
}

func toQualityPayloadV1(value flightfeatures.FeatureQuality) qualityPayloadV1 {
	return qualityPayloadV1{
		Status: string(value.Status), CompletenessScore: value.CompletenessScore,
		OptionalCoverageScore: value.OptionalCoverageScore,
		InputQualityScore:     value.InputQualityScore,
		SupportingPointCount:  value.SupportingPointCount,
		Limitations:           toLimitationsPayloadV1(value.Limitations),
	}
}

func fromQualityPayloadV1(value qualityPayloadV1) flightfeatures.FeatureQuality {
	return flightfeatures.FeatureQuality{
		Status:                flightfeatures.ValidationStatus(value.Status),
		CompletenessScore:     value.CompletenessScore,
		OptionalCoverageScore: value.OptionalCoverageScore,
		InputQualityScore:     value.InputQualityScore,
		SupportingPointCount:  value.SupportingPointCount,
		Limitations:           fromLimitationsPayloadV1(value.Limitations),
	}
}

func toProcessingVersionsPayloadV1(value flightfeatures.ProcessingComponentVersions) processingVersionsPayloadV1 {
	return processingVersionsPayloadV1{
		Composition: value.Composition, Extractor: value.Extractor,
		AircraftProvider: value.AircraftProvider, TemporalBuilder: value.TemporalBuilder,
		GeographicalBuilder: value.GeographicalBuilder,
		OperationalBuilder:  value.OperationalBuilder, TrajectoryBuilder: value.TrajectoryBuilder,
	}
}

func fromProcessingVersionsPayloadV1(value processingVersionsPayloadV1) flightfeatures.ProcessingComponentVersions {
	return flightfeatures.ProcessingComponentVersions{
		Composition: value.Composition, Extractor: value.Extractor,
		AircraftProvider: value.AircraftProvider, TemporalBuilder: value.TemporalBuilder,
		GeographicalBuilder: value.GeographicalBuilder,
		OperationalBuilder:  value.OperationalBuilder, TrajectoryBuilder: value.TrajectoryBuilder,
	}
}

func toProcessingIdentityPayloadV1(value flightfeatures.ProcessingIdentity) processingIdentityPayloadV1 {
	return processingIdentityPayloadV1{
		Versions:                      toProcessingVersionsPayloadV1(value.Versions),
		GeographicCellPrecision:       value.GeographicCellPrecision,
		AircraftEnrichmentMode:        string(value.AircraftEnrichmentMode),
		AircraftCacheMode:             value.AircraftCacheMode,
		AircraftPositiveCacheTTL:      int64(value.AircraftPositiveCacheTTL),
		AircraftNegativeCacheTTL:      int64(value.AircraftNegativeCacheTTL),
		AircraftNotFoundPolicyVersion: value.AircraftNotFoundPolicyVersion,
		AircraftMetadataSourceName:    value.AircraftMetadataSourceName,
	}
}

func fromProcessingIdentityPayloadV1(value processingIdentityPayloadV1) flightfeatures.ProcessingIdentity {
	return flightfeatures.ProcessingIdentity{
		Versions:                      fromProcessingVersionsPayloadV1(value.Versions),
		GeographicCellPrecision:       value.GeographicCellPrecision,
		AircraftEnrichmentMode:        flightfeatures.AircraftEnrichmentMode(value.AircraftEnrichmentMode),
		AircraftCacheMode:             value.AircraftCacheMode,
		AircraftPositiveCacheTTL:      time.Duration(value.AircraftPositiveCacheTTL),
		AircraftNegativeCacheTTL:      time.Duration(value.AircraftNegativeCacheTTL),
		AircraftNotFoundPolicyVersion: value.AircraftNotFoundPolicyVersion,
		AircraftMetadataSourceName:    value.AircraftMetadataSourceName,
	}
}

func toProvenancePayloadV1(value flightfeatures.FeatureProvenance) provenancePayloadV1 {
	return provenancePayloadV1{
		ProcessingVersion: string(value.ProcessingVersion), ExtractorVersion: value.ExtractorVersion,
		InputFingerprint:                value.InputFingerprint,
		ProcessingIdentityFingerprint:   value.ProcessingIdentityFingerprint,
		ProcessingIdentity:              toProcessingIdentityPayloadV1(value.ProcessingIdentity),
		TrajectoryCreatedAt:             value.TrajectoryCreatedAt,
		TrajectoryUpdatedAt:             value.TrajectoryUpdatedAt,
		AircraftMetadataSourceName:      value.AircraftMetadataSourceName,
		AircraftMetadataProviderVersion: value.AircraftMetadataProviderVersion,
		AircraftMetadataRetrievedAt:     value.AircraftMetadataRetrievedAt,
		SourceNames:                     cloneStringSlice(value.SourceNames),
	}
}

func fromProvenancePayloadV1(value provenancePayloadV1) flightfeatures.FeatureProvenance {
	return flightfeatures.FeatureProvenance{
		ProcessingVersion: flightfeatures.ProcessingVersion(value.ProcessingVersion),
		ExtractorVersion:  value.ExtractorVersion, InputFingerprint: value.InputFingerprint,
		ProcessingIdentityFingerprint:   value.ProcessingIdentityFingerprint,
		ProcessingIdentity:              fromProcessingIdentityPayloadV1(value.ProcessingIdentity),
		TrajectoryCreatedAt:             value.TrajectoryCreatedAt,
		TrajectoryUpdatedAt:             value.TrajectoryUpdatedAt,
		AircraftMetadataSourceName:      value.AircraftMetadataSourceName,
		AircraftMetadataProviderVersion: value.AircraftMetadataProviderVersion,
		AircraftMetadataRetrievedAt:     value.AircraftMetadataRetrievedAt,
		SourceNames:                     cloneStringSlice(value.SourceNames),
	}
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string{}, values...)
}
