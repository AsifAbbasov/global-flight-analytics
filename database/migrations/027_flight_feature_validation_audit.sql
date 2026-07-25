BEGIN;

UPDATE flight_feature_snapshots
SET features_json = jsonb_set(
	features_json,
	'{ValidationReport}',
	jsonb_build_object(
		'AuditState',
		'legacy_unavailable',
		'ValidatorVersion',
		'',
		'Status',
		validation_status,
		'ErrorCount',
		0,
		'WarningCount',
		0,
		'Issues',
		'[]'::jsonb,
		'ValidatedAt',
		'0001-01-01T00:00:00Z'
	),
	true
)
WHERE NOT (features_json ? 'ValidationReport');

ALTER TABLE flight_feature_snapshots
	ADD CONSTRAINT
		flight_feature_snapshots_validation_report_present
	CHECK (
		features_json ? 'ValidationReport'
	);

ALTER TABLE flight_feature_snapshots
	ADD CONSTRAINT
		flight_feature_snapshots_validation_report_object
	CHECK (
		jsonb_typeof(
			features_json -> 'ValidationReport'
		) = 'object'
	);

ALTER TABLE flight_feature_snapshots
	ADD CONSTRAINT
		flight_feature_snapshots_validation_report_state
	CHECK (
		features_json #>>
			'{ValidationReport,AuditState}'
		IN (
			'complete',
			'legacy_unavailable'
		)
	);

ALTER TABLE flight_feature_snapshots
	ADD CONSTRAINT
		flight_feature_snapshots_validation_report_status_match
	CHECK (
		features_json #>>
			'{ValidationReport,Status}'
		= validation_status
	);

COMMIT;
