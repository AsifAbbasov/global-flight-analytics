CREATE TABLE IF NOT EXISTS provider_budget_fixed_windows (
    provider_name text NOT NULL,
    limit_index integer NOT NULL,
    window_start timestamptz NOT NULL,
    window_end timestamptz NOT NULL,
    request_count integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (
        provider_name,
        limit_index
    ),
    CONSTRAINT provider_budget_fixed_windows_limit_index_nonnegative
        CHECK (limit_index >= 0),
    CONSTRAINT provider_budget_fixed_windows_count_nonnegative
        CHECK (request_count >= 0),
    CONSTRAINT provider_budget_fixed_windows_bounds_valid
        CHECK (window_end > window_start)
);

CREATE INDEX IF NOT EXISTS
    provider_budget_fixed_windows_expiry_idx
ON provider_budget_fixed_windows (window_end);

CREATE TABLE IF NOT EXISTS provider_budget_reported_states (
    provider_name text PRIMARY KEY,
    remaining_known boolean NOT NULL,
    remaining integer NOT NULL DEFAULT 0,
    retry_at timestamptz,
    observed_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT provider_budget_reported_states_remaining_nonnegative
        CHECK (remaining >= 0),
    CONSTRAINT provider_budget_reported_states_unknown_has_no_remaining
        CHECK (remaining_known OR remaining = 0),
    CONSTRAINT provider_budget_reported_states_unknown_has_probe_lease
        CHECK (remaining_known OR retry_at IS NOT NULL)
);
