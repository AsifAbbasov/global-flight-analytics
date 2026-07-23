BEGIN;

ALTER TABLE weather_snapshots
    ALTER COLUMN temperature_celsius DROP NOT NULL,
    ALTER COLUMN relative_humidity_percent DROP NOT NULL,
    ALTER COLUMN precipitation_mm DROP NOT NULL,
    ALTER COLUMN rain_mm DROP NOT NULL,
    ALTER COLUMN weather_code DROP NOT NULL,
    ALTER COLUMN cloud_cover_percent DROP NOT NULL,
    ALTER COLUMN surface_pressure_hpa DROP NOT NULL,
    ALTER COLUMN wind_speed_mps DROP NOT NULL,
    ALTER COLUMN wind_direction_degrees DROP NOT NULL,
    ALTER COLUMN wind_gusts_mps DROP NOT NULL;

COMMENT ON COLUMN weather_snapshots.temperature_celsius IS
    'NULL means the provider did not supply a temperature observation.';
COMMENT ON COLUMN weather_snapshots.relative_humidity_percent IS
    'NULL means the provider did not supply a relative humidity observation.';
COMMENT ON COLUMN weather_snapshots.precipitation_mm IS
    'NULL means the provider did not supply a precipitation observation.';
COMMENT ON COLUMN weather_snapshots.rain_mm IS
    'NULL means the provider did not supply a rain observation.';
COMMENT ON COLUMN weather_snapshots.weather_code IS
    'NULL means the provider did not supply a weather code.';
COMMENT ON COLUMN weather_snapshots.cloud_cover_percent IS
    'NULL means the provider did not supply a cloud cover observation.';
COMMENT ON COLUMN weather_snapshots.surface_pressure_hpa IS
    'NULL means the provider did not supply a surface pressure observation.';
COMMENT ON COLUMN weather_snapshots.wind_speed_mps IS
    'NULL means the provider did not supply a wind speed observation.';
COMMENT ON COLUMN weather_snapshots.wind_direction_degrees IS
    'NULL means the provider did not supply a wind direction observation.';
COMMENT ON COLUMN weather_snapshots.wind_gusts_mps IS
    'NULL means the provider did not supply a wind gust observation.';

COMMIT;
