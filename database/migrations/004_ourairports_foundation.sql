BEGIN;

ALTER TABLE airports
    ADD COLUMN source_ident text,
    ADD COLUMN source_country_code varchar(2);

ALTER TABLE airports
    DROP CONSTRAINT airports_identifier_check;

ALTER TABLE airports
    ADD CONSTRAINT airports_identifier_check
        CHECK (
            source_ident IS NOT NULL
            OR icao_code IS NOT NULL
            OR iata_code IS NOT NULL
        );

ALTER TABLE airports
    ADD CONSTRAINT airports_source_ident_check
        CHECK (
            source_ident IS NULL
            OR btrim(source_ident) <> ''
        );

ALTER TABLE airports
    ADD CONSTRAINT airports_source_country_code_check
        CHECK (
            source_country_code IS NULL
            OR char_length(source_country_code) = 2
        );

CREATE UNIQUE INDEX airports_source_identity_unique
    ON airports (source_name, source_ident)
    WHERE source_ident IS NOT NULL;

CREATE INDEX airports_source_country_code_idx
    ON airports (source_country_code)
    WHERE source_country_code IS NOT NULL;

COMMIT;
