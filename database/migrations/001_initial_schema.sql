BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE countries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    iso2 varchar(2) NOT NULL UNIQUE,
    iso3 varchar(3) NOT NULL UNIQUE,
    continent text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE regions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    slug text NOT NULL UNIQUE,
    description text,
    min_latitude numeric NOT NULL,
    max_latitude numeric NOT NULL,
    min_longitude numeric NOT NULL,
    max_longitude numeric NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT regions_latitude_range_check
        CHECK (
            min_latitude >= -90
            AND max_latitude <= 90
            AND min_latitude < max_latitude
        ),

    CONSTRAINT regions_longitude_range_check
        CHECK (
            min_longitude >= -180
            AND max_longitude <= 180
            AND min_longitude < max_longitude
        )
);

CREATE TABLE airlines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    icao_code varchar(3) UNIQUE,
    iata_code varchar(2),
    country_id uuid REFERENCES countries(id) ON DELETE SET NULL,
    website text,
    source_name text NOT NULL,
    last_synced_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE aircraft_models (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    manufacturer text NOT NULL,
    model text NOT NULL,
    aircraft_type text,
    max_speed_kmh integer,
    max_range_km integer,
    passenger_capacity integer,
    cargo_capacity_kg integer,
    source_name text NOT NULL,
    last_synced_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT aircraft_models_unique_manufacturer_model
        UNIQUE (manufacturer, model),

    CONSTRAINT aircraft_models_positive_values_check
        CHECK (
            (max_speed_kmh IS NULL OR max_speed_kmh > 0)
            AND (max_range_km IS NULL OR max_range_km > 0)
            AND (passenger_capacity IS NULL OR passenger_capacity >= 0)
            AND (cargo_capacity_kg IS NULL OR cargo_capacity_kg >= 0)
        )
);

CREATE TABLE aircraft (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    icao24 varchar(10) NOT NULL UNIQUE,
    registration text,
    model_id uuid REFERENCES aircraft_models(id) ON DELETE SET NULL,
    airline_id uuid REFERENCES airlines(id) ON DELETE SET NULL,
    country_id uuid REFERENCES countries(id) ON DELETE SET NULL,
    source_name text NOT NULL,
    first_seen_at timestamptz,
    last_seen_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE airports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    icao_code varchar(4) UNIQUE,
    iata_code varchar(3),
    name text NOT NULL,
    city text,
    country_id uuid REFERENCES countries(id) ON DELETE SET NULL,
    latitude numeric NOT NULL,
    longitude numeric NOT NULL,
    elevation_ft integer,
    timezone text,
    source_name text NOT NULL,
    last_synced_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT airports_coordinates_check
        CHECK (
            latitude >= -90
            AND latitude <= 90
            AND longitude >= -180
            AND longitude <= 180
        ),

    CONSTRAINT airports_identifier_check
        CHECK (
            icao_code IS NOT NULL
            OR iata_code IS NOT NULL
        )
);

CREATE TABLE runways (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    airport_id uuid NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    identifier text NOT NULL,
    length_m integer,
    width_m integer,
    surface text,
    source_name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT runways_positive_dimensions_check
        CHECK (
            (length_m IS NULL OR length_m > 0)
            AND (width_m IS NULL OR width_m > 0)
        )
);

CREATE TABLE airport_facilities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    airport_id uuid NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    facility_type text NOT NULL,
    name text,
    latitude numeric,
    longitude numeric,
    source_name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT airport_facilities_coordinates_check
        CHECK (
            (latitude IS NULL OR (latitude >= -90 AND latitude <= 90))
            AND (longitude IS NULL OR (longitude >= -180 AND longitude <= 180))
        )
);

CREATE TABLE airport_profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    airport_id uuid NOT NULL UNIQUE REFERENCES airports(id) ON DELETE CASCADE,
    description text,
    history text,
    passenger_traffic bigint,
    cargo_traffic_tons bigint,
    terminals_count integer,
    runways_count integer,
    metadata_json jsonb,
    source_name text,
    last_updated_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT airport_profiles_positive_values_check
        CHECK (
            (passenger_traffic IS NULL OR passenger_traffic >= 0)
            AND (cargo_traffic_tons IS NULL OR cargo_traffic_tons >= 0)
            AND (terminals_count IS NULL OR terminals_count >= 0)
            AND (runways_count IS NULL OR runways_count >= 0)
        )
);

CREATE TABLE flights (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    aircraft_id uuid REFERENCES aircraft(id) ON DELETE SET NULL,
    callsign text,
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    status text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT flights_status_check
        CHECK (status IN ('active', 'completed', 'lost', 'unknown')),

    CONSTRAINT flights_seen_at_check
        CHECK (first_seen_at <= last_seen_at)
);

CREATE TABLE ingestion_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_name text NOT NULL,
    region_id uuid REFERENCES regions(id) ON DELETE SET NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz,
    status text NOT NULL,
    records_received integer NOT NULL DEFAULT 0,
    records_inserted integer NOT NULL DEFAULT 0,
    records_updated integer NOT NULL DEFAULT 0,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT ingestion_runs_status_check
        CHECK (status IN ('running', 'success', 'failed', 'partial')),

    CONSTRAINT ingestion_runs_counts_check
        CHECK (
            records_received >= 0
            AND records_inserted >= 0
            AND records_updated >= 0
        ),

    CONSTRAINT ingestion_runs_time_check
        CHECK (
            finished_at IS NULL
            OR started_at <= finished_at
        )
);

CREATE TABLE flight_states (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id uuid REFERENCES flights(id) ON DELETE SET NULL,
    aircraft_id uuid REFERENCES aircraft(id) ON DELETE SET NULL,
    icao24 varchar(10) NOT NULL,
    callsign text,
    latitude numeric,
    longitude numeric,
    barometric_altitude_m integer,
    geometric_altitude_m integer,
    velocity_mps numeric,
    heading_degrees numeric,
    vertical_rate_mps numeric,
    on_ground boolean,
    origin_country text,
    observed_at timestamptz NOT NULL,
    source_name text NOT NULL,
    ingestion_run_id uuid REFERENCES ingestion_runs(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT flight_states_coordinates_check
        CHECK (
            (latitude IS NULL OR (latitude >= -90 AND latitude <= 90))
            AND (longitude IS NULL OR (longitude >= -180 AND longitude <= 180))
        ),

    CONSTRAINT flight_states_heading_check
        CHECK (
            heading_degrees IS NULL
            OR (heading_degrees >= 0 AND heading_degrees <= 360)
        ),

    CONSTRAINT flight_states_velocity_check
        CHECK (
            velocity_mps IS NULL
            OR velocity_mps >= 0
        )
);

CREATE TABLE route_predictions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id uuid REFERENCES flights(id) ON DELETE SET NULL,
    aircraft_id uuid REFERENCES aircraft(id) ON DELETE SET NULL,
    origin_airport_id uuid REFERENCES airports(id) ON DELETE SET NULL,
    destination_airport_id uuid REFERENCES airports(id) ON DELETE SET NULL,
    confidence_level varchar(20) NOT NULL,
    confidence_score numeric,
    method_name text NOT NULL,
    data_source text NOT NULL,
    calculated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT route_predictions_confidence_level_check
        CHECK (confidence_level IN ('High', 'Medium', 'Low')),

    CONSTRAINT route_predictions_confidence_score_check
        CHECK (
            confidence_score IS NULL
            OR (confidence_score >= 0 AND confidence_score <= 1)
        ),

    CONSTRAINT route_predictions_airports_check
        CHECK (
            origin_airport_id IS NULL
            OR destination_airport_id IS NULL
            OR origin_airport_id <> destination_airport_id
        )
);

CREATE TABLE traffic_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id uuid NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    snapshot_time timestamptz NOT NULL,
    flight_count integer NOT NULL DEFAULT 0,
    airport_count integer NOT NULL DEFAULT 0,
    route_count integer NOT NULL DEFAULT 0,
    payload_json jsonb,
    calculated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT traffic_snapshots_counts_check
        CHECK (
            flight_count >= 0
            AND airport_count >= 0
            AND route_count >= 0
        )
);

CREATE TABLE route_statistics (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    route_prediction_id uuid REFERENCES route_predictions(id) ON DELETE SET NULL,
    origin_airport_id uuid REFERENCES airports(id) ON DELETE SET NULL,
    destination_airport_id uuid REFERENCES airports(id) ON DELETE SET NULL,
    observation_date date NOT NULL,
    flight_count integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT route_statistics_flight_count_check
        CHECK (flight_count >= 0),

    CONSTRAINT route_statistics_airports_check
        CHECK (
            origin_airport_id IS NULL
            OR destination_airport_id IS NULL
            OR origin_airport_id <> destination_airport_id
        )
);

CREATE TABLE airport_statistics (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    airport_id uuid NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    observation_date date NOT NULL,
    arrivals integer NOT NULL DEFAULT 0,
    departures integer NOT NULL DEFAULT 0,
    total_flights integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT airport_statistics_counts_check
        CHECK (
            arrivals >= 0
            AND departures >= 0
            AND total_flights >= 0
        )
);

CREATE INDEX idx_airlines_country_id
    ON airlines(country_id);

CREATE INDEX idx_aircraft_registration
    ON aircraft(registration);

CREATE INDEX idx_aircraft_airline_id
    ON aircraft(airline_id);

CREATE INDEX idx_aircraft_country_id
    ON aircraft(country_id);

CREATE INDEX idx_airports_iata_code
    ON airports(iata_code);

CREATE INDEX idx_airports_country_id
    ON airports(country_id);

CREATE INDEX idx_runways_airport_id
    ON runways(airport_id);

CREATE INDEX idx_airport_facilities_airport_id
    ON airport_facilities(airport_id);

CREATE INDEX idx_flights_aircraft_id
    ON flights(aircraft_id);

CREATE INDEX idx_flights_callsign
    ON flights(callsign);

CREATE INDEX idx_flight_states_icao24
    ON flight_states(icao24);

CREATE INDEX idx_flight_states_flight_id
    ON flight_states(flight_id);

CREATE INDEX idx_flight_states_aircraft_id
    ON flight_states(aircraft_id);

CREATE INDEX idx_flight_states_observed_at
    ON flight_states(observed_at);

CREATE INDEX idx_flight_states_ingestion_run_id
    ON flight_states(ingestion_run_id);

CREATE INDEX idx_flight_states_icao24_observed_at
    ON flight_states(icao24, observed_at);

CREATE INDEX idx_route_predictions_flight_id
    ON route_predictions(flight_id);

CREATE INDEX idx_route_predictions_aircraft_id
    ON route_predictions(aircraft_id);

CREATE INDEX idx_route_predictions_origin_airport_id
    ON route_predictions(origin_airport_id);

CREATE INDEX idx_route_predictions_destination_airport_id
    ON route_predictions(destination_airport_id);

CREATE INDEX idx_traffic_snapshots_region_id
    ON traffic_snapshots(region_id);

CREATE INDEX idx_traffic_snapshots_snapshot_time
    ON traffic_snapshots(snapshot_time);

CREATE INDEX idx_traffic_snapshots_region_snapshot_time
    ON traffic_snapshots(region_id, snapshot_time);

CREATE INDEX idx_airport_statistics_airport_id
    ON airport_statistics(airport_id);

CREATE INDEX idx_airport_statistics_observation_date
    ON airport_statistics(observation_date);

CREATE INDEX idx_airport_statistics_airport_observation_date
    ON airport_statistics(airport_id, observation_date);

CREATE INDEX idx_route_statistics_observation_date
    ON route_statistics(observation_date);

CREATE INDEX idx_route_statistics_origin_destination_date
    ON route_statistics(origin_airport_id, destination_airport_id, observation_date);

CREATE INDEX idx_ingestion_runs_source_name
    ON ingestion_runs(source_name);

CREATE INDEX idx_ingestion_runs_started_at
    ON ingestion_runs(started_at);

COMMIT;
