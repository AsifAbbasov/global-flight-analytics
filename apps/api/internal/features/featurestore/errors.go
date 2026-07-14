package featurestore

import "errors"

var (
	ErrTrajectoryIDRequired = errors.New(
		"feature snapshot trajectory id is required",
	)
	ErrUnsupportedSchemaVersion = errors.New(
		"feature snapshot schema version is unsupported",
	)
	ErrAsOfTimeRequired = errors.New(
		"feature snapshot as-of time is required",
	)
	ErrInputFingerprintRequired = errors.New(
		"feature snapshot input fingerprint is required",
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
