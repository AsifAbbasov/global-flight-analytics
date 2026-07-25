BEGIN;

ALTER TABLE flight_feature_snapshots
	ADD COLUMN processing_version text;

WITH resolved_versions AS (
	SELECT
		id,
		COALESCE(
			NULLIF(
				BTRIM(
					features_json #>> '{Provenance,ProcessingVersion}'
				),
				''
			),
			'flight-feature-processing-legacy-v1'
		) AS processing_version
	FROM flight_feature_snapshots
)
UPDATE flight_feature_snapshots AS snapshots
SET
	processing_version = resolved_versions.processing_version,
	features_json = jsonb_set(
		snapshots.features_json,
		'{Provenance,ProcessingVersion}',
		to_jsonb(resolved_versions.processing_version::text),
		true
	)
FROM resolved_versions
WHERE resolved_versions.id = snapshots.id;

ALTER TABLE flight_feature_snapshots
	ALTER COLUMN processing_version SET NOT NULL;

ALTER TABLE flight_feature_snapshots
	ADD CONSTRAINT flight_feature_snapshots_processing_version_not_blank
	CHECK (BTRIM(processing_version) <> '');

DO $$
DECLARE
	target_constraint text;
BEGIN
	SELECT conname
	INTO target_constraint
	FROM pg_constraint
	WHERE conrelid = 'flight_feature_snapshots'::regclass
	  AND contype = 'u'
	  AND pg_get_constraintdef(oid) =
		'UNIQUE (trajectory_id, schema_version, as_of_time_unix_nano)'
	LIMIT 1;

	IF target_constraint IS NOT NULL THEN
		EXECUTE format(
			'ALTER TABLE flight_feature_snapshots DROP CONSTRAINT %I',
			target_constraint
		);
	END IF;
END
$$;

CREATE UNIQUE INDEX
	flight_feature_snapshots_processing_identity_uq
ON flight_feature_snapshots (
	trajectory_id,
	schema_version,
	processing_version,
	as_of_time_unix_nano
);

CREATE INDEX
	flight_feature_snapshots_processing_latest_idx
ON flight_feature_snapshots (
	trajectory_id,
	schema_version,
	processing_version,
	as_of_time_unix_nano DESC,
	id ASC
);

COMMIT;
