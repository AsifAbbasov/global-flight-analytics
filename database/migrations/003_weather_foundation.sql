BEGIN;

CREATE TABLE weather_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    provider text NOT NULL,

    latitude numeric NOT NULL,
    longitude numeric NOT NULL,

    observed_at timestamptz NOT NULL,
    retrieved_at timestamptz NOT NULL,

    temperature_celsius numeric,
    relative_humidity_percent integer,

    precipitation_mm numeric,
    rain_mm numeric,

    weather_code integer,
    cloud_cover_percent integer,

    surface_pressure_hpa numeric,

    wind_speed_mps numeric,
    wind_direction_degrees integer,
    wind_gusts_mps numeric,

    metadata_json jsonb,

    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT weather_snapshots_provider_check
        CHECK (length(trim(provider)) > 0),

    CONSTRAINT weather_snapshots_coordinates_check
        CHECK (
            latitude >= -90
            AND latitude <= 90
            AND longitude >= -180
            AND longitude <= 180
        ),

    CONSTRAINT weather_snapshots_humidity_check
        CHECK (
            relative_humidity_percent IS NULL
            OR (
                relative_humidity_percent >= 0
                AND relative_humidity_percent <= 100
            )
        ),

    CONSTRAINT weather_snapshots_cloud_cover_check
        CHECK (
            cloud_cover_percent IS NULL
            OR (
                cloud_cover_percent >= 0
                AND cloud_cover_percent <= 100
            )
        ),

    CONSTRAINT weather_snapshots_precipitation_check
        CHECK (
            (precipitation_mm IS NULL OR precipitation_mm >= 0)
            AND (rain_mm IS NULL OR rain_mm >= 0)
        ),

    CONSTRAINT weather_snapshots_pressure_check
        CHECK (
            surface_pressure_hpa IS NULL
            OR surface_pressure_hpa > 0
        ),

    CONSTRAINT weather_snapshots_wind_check
        CHECK (
            (wind_speed_mps IS NULL OR wind_speed_mps >= 0)
            AND (
                wind_direction_degrees IS NULL
                OR (
                    wind_direction_degrees >= 0
                    AND wind_direction_degrees <= 360
                )
            )
            AND (wind_gusts_mps IS NULL OR wind_gusts_mps >= 0)
        ),

    CONSTRAINT weather_snapshots_unique_provider_location_time
        UNIQUE (provider, latitude, longitude, observed_at)
);

CREATE TABLE airport_weather_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    airport_id uuid NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    weather_snapshot_id uuid NOT NULL REFERENCES weather_snapshots(id) ON DELETE CASCADE,

    provider text NOT NULL,

    airport_icao_code varchar(4),
    airport_iata_code varchar(3),

    observed_at timestamptz NOT NULL,
    retrieved_at timestamptz NOT NULL,

    distance_km numeric NOT NULL DEFAULT 0,

    weather_impact_level text NOT NULL DEFAULT 'unknown',
    weather_impact_score numeric,

    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT airport_weather_provider_check
        CHECK (length(trim(provider)) > 0),

    CONSTRAINT airport_weather_distance_check
        CHECK (distance_km >= 0),

    CONSTRAINT airport_weather_impact_level_check
        CHECK (weather_impact_level IN ('low', 'medium', 'high', 'unknown')),

    CONSTRAINT airport_weather_impact_score_check
        CHECK (
            weather_impact_score IS NULL
            OR (
                weather_impact_score >= 0
                AND weather_impact_score <= 1
            )
        ),

    CONSTRAINT airport_weather_unique_snapshot
        UNIQUE (airport_id, weather_snapshot_id)
);

CREATE TABLE trajectory_weather_context (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    trajectory_id uuid NOT NULL REFERENCES flight_trajectories(id) ON DELETE CASCADE,
    trajectory_segment_id uuid REFERENCES trajectory_segments(id) ON DELETE SET NULL,
    weather_snapshot_id uuid NOT NULL REFERENCES weather_snapshots(id) ON DELETE CASCADE,

    context_type text NOT NULL,

    latitude numeric,
    longitude numeric,

    distance_km numeric,
    observed_at timestamptz NOT NULL,

    weather_impact_level text NOT NULL DEFAULT 'unknown',
    weather_impact_score numeric,

    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT trajectory_weather_context_type_check
        CHECK (context_type IN ('departure', 'arrival', 'enroute', 'nearest_point', 'unknown')),

    CONSTRAINT trajectory_weather_coordinates_check
        CHECK (
            (latitude IS NULL OR (latitude >= -90 AND latitude <= 90))
            AND (longitude IS NULL OR (longitude >= -180 AND longitude <= 180))
        ),

    CONSTRAINT trajectory_weather_distance_check
        CHECK (
            distance_km IS NULL
            OR distance_km >= 0
        ),

    CONSTRAINT trajectory_weather_impact_level_check
        CHECK (weather_impact_level IN ('low', 'medium', 'high', 'unknown')),

    CONSTRAINT trajectory_weather_impact_score_check
        CHECK (
            weather_impact_score IS NULL
            OR (
                weather_impact_score >= 0
                AND weather_impact_score <= 1
            )
        ),

    CONSTRAINT trajectory_weather_unique_context
        UNIQUE (trajectory_id, trajectory_segment_id, weather_snapshot_id, context_type)
);

CREATE INDEX weather_snapshots_provider_observed_idx
    ON weather_snapshots (provider, observed_at DESC);

CREATE INDEX weather_snapshots_location_observed_idx
    ON weather_snapshots (latitude, longitude, observed_at DESC);

CREATE INDEX airport_weather_airport_observed_idx
    ON airport_weather_snapshots (airport_id, observed_at DESC);

CREATE INDEX airport_weather_snapshot_idx
    ON airport_weather_snapshots (weather_snapshot_id);

CREATE INDEX trajectory_weather_trajectory_idx
    ON trajectory_weather_context (trajectory_id, observed_at DESC);

CREATE INDEX trajectory_weather_segment_idx
    ON trajectory_weather_context (trajectory_segment_id);

CREATE INDEX trajectory_weather_snapshot_idx
    ON trajectory_weather_context (weather_snapshot_id);

COMMIT;
