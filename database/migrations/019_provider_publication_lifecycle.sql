BEGIN;

CREATE TABLE provider_publications (
    provider_name text NOT NULL,
    publication_id text NOT NULL,
    status text NOT NULL,
    reservation_token uuid NOT NULL,
    reserved_at timestamptz NOT NULL,
    lease_expires_at timestamptz,
    committed_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (provider_name, publication_id),
    CONSTRAINT provider_publications_provider_name_nonempty_check
        CHECK (btrim(provider_name) <> ''),
    CONSTRAINT provider_publications_publication_id_nonempty_check
        CHECK (btrim(publication_id) <> ''),
    CONSTRAINT provider_publications_status_check
        CHECK (status IN ('reserved', 'committed')),
    CONSTRAINT provider_publications_lifecycle_check
        CHECK (
            (
                status = 'reserved'
                AND committed_at IS NULL
                AND lease_expires_at IS NOT NULL
                AND reserved_at < lease_expires_at
            )
            OR
            (
                status = 'committed'
                AND committed_at IS NOT NULL
                AND lease_expires_at IS NULL
                AND reserved_at <= committed_at
            )
        ),
    CONSTRAINT provider_publications_updated_at_check
        CHECK (updated_at >= reserved_at)
);

CREATE INDEX provider_publications_reserved_lease_idx
    ON provider_publications (lease_expires_at)
    WHERE status = 'reserved';

COMMIT;
