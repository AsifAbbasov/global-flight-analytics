package featurestore

import "errors"

var (
	ErrContextRequired = errors.New(
		"feature snapshot request context is required",
	)
	ErrTrajectoryIDRequired = errors.New(
		"feature snapshot trajectory id is required",
	)
	ErrUnsupportedSchemaVersion = errors.New(
		"feature snapshot schema version is unsupported",
	)
	ErrProcessingVersionRequired = errors.New(
		"feature snapshot processing version is required",
	)
	ErrAsOfTimeRequired = errors.New(
		"feature snapshot as-of time is required",
	)
	ErrInputFingerprintRequired = errors.New(
		"feature snapshot input fingerprint is required",
	)
	ErrInvalidInputFingerprint = errors.New(
		"feature snapshot input fingerprint must use sha256 lowercase hexadecimal format",
	)
	ErrValidationProofRequired = errors.New(
		"feature snapshot requires a complete validator audit proof",
	)
	ErrNonFiniteFeatureValue = errors.New(
		"feature snapshot contains a non-finite numeric value",
	)
	ErrMemoryCapacityExceeded = errors.New(
		"feature snapshot memory store capacity is exhausted",
	)
	ErrFeaturesUnvalidated = errors.New(
		"unvalidated features cannot be stored",
	)
	ErrFeaturesInvalid = errors.New(
		"invalid features cannot be stored",
	)
	ErrSnapshotNotFound = errors.New(
		"feature snapshot was not found",
	)
	ErrSnapshotConflict = errors.New(
		"feature snapshot key already exists with different evidence",
	)
	ErrInvalidListLimit = errors.New(
		"feature snapshot list limit must be between one and one hundred",
	)
)
