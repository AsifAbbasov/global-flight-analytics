BEGIN;

CREATE TABLE source_http_validators (
    source_name text NOT NULL,
    resource_url text NOT NULL,
    etag text,
    last_modified text,
    observed_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (
        source_name,
        resource_url
    ),

    CONSTRAINT source_http_validators_source_name_check
        CHECK (
            btrim(source_name) <> ''
        ),

    CONSTRAINT source_http_validators_resource_url_check
        CHECK (
            btrim(resource_url) <> ''
        ),

    CONSTRAINT source_http_validators_etag_check
        CHECK (
            etag IS NULL
            OR btrim(etag) <> ''
        ),

    CONSTRAINT source_http_validators_last_modified_check
        CHECK (
            last_modified IS NULL
            OR btrim(last_modified) <> ''
        ),

    CONSTRAINT source_http_validators_validator_check
        CHECK (
            etag IS NOT NULL
            OR last_modified IS NOT NULL
        )
);

COMMIT;
