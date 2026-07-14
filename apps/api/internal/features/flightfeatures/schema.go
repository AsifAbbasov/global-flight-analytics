package flightfeatures

type FeatureGroup string

const (
	FeatureGroupTemporal     FeatureGroup = "temporal"
	FeatureGroupGeographical FeatureGroup = "geographical"
	FeatureGroupOperational  FeatureGroup = "operational"
	FeatureGroupTrajectory   FeatureGroup = "trajectory"
	FeatureGroupAircraft     FeatureGroup = "aircraft"
)

type FeatureValueType string

const (
	FeatureValueTypeBoolean FeatureValueType = "boolean"
	FeatureValueTypeFloat64 FeatureValueType = "float64"
	FeatureValueTypeInteger FeatureValueType = "integer"
	FeatureValueTypeString  FeatureValueType = "string"
)

type FeatureDefinition struct {
	Name        string
	Group       FeatureGroup
	ValueType   FeatureValueType
	Unit        string
	Required    bool
	Description string
}

type Schema struct {
	Version     SchemaVersion
	Definitions []FeatureDefinition
}

var currentDefinitions = []FeatureDefinition{
	{
		Name:        "temporal.duration_seconds",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "seconds",
		Required:    true,
		Description: "Duration of the feature observation window.",
	},
	{
		Name:        "temporal.start_hour_utc",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "hour",
		Required:    true,
		Description: "UTC hour containing the first observation.",
	},
	{
		Name:        "temporal.end_hour_utc",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "hour",
		Required:    true,
		Description: "UTC hour containing the final observation.",
	},
	{
		Name:        "temporal.start_weekday",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "weekday",
		Required:    true,
		Description: "UTC weekday of the first observation.",
	},
	{
		Name:        "temporal.end_weekday",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "weekday",
		Required:    true,
		Description: "UTC weekday of the final observation.",
	},
	{
		Name:        "temporal.start_minute_of_day_utc",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "minute",
		Required:    true,
		Description: "Minute of the UTC day for the first observation.",
	},
	{
		Name:        "temporal.end_minute_of_day_utc",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "minute",
		Required:    true,
		Description: "Minute of the UTC day for the final observation.",
	},
	{
		Name:        "temporal.crosses_utc_midnight",
		Group:       FeatureGroupTemporal,
		ValueType:   FeatureValueTypeBoolean,
		Required:    true,
		Description: "Whether the observation window crosses a UTC calendar boundary.",
	},
	{
		Name:        "geographical.start_latitude",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "degrees",
		Required:    true,
		Description: "Latitude of the first usable trajectory point.",
	},
	{
		Name:        "geographical.start_longitude",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "degrees",
		Required:    true,
		Description: "Longitude of the first usable trajectory point.",
	},
	{
		Name:        "geographical.end_latitude",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "degrees",
		Required:    true,
		Description: "Latitude of the final usable trajectory point.",
	},
	{
		Name:        "geographical.end_longitude",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "degrees",
		Required:    true,
		Description: "Longitude of the final usable trajectory point.",
	},
	{
		Name:        "geographical.latitude_span_degrees",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "degrees",
		Required:    true,
		Description: "North-south span of usable trajectory points.",
	},
	{
		Name:        "geographical.longitude_span_degrees",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "degrees",
		Required:    true,
		Description: "East-west span of usable trajectory points.",
	},
	{
		Name:        "geographical.great_circle_distance_km",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "kilometres",
		Required:    true,
		Description: "Great-circle distance between the first and final usable points.",
	},
	{
		Name:        "geographical.observed_path_distance_km",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "kilometres",
		Required:    true,
		Description: "Cumulative distance along usable observed movement.",
	},
	{
		Name:        "geographical.maximum_displacement_km",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "kilometres",
		Required:    true,
		Description: "Maximum displacement from the first usable point.",
	},
	{
		Name:        "geographical.crosses_antimeridian",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeBoolean,
		Required:    true,
		Description: "Whether the usable path crosses the antimeridian.",
	},
	{
		Name:        "geographical.unique_geographic_cell_count",
		Group:       FeatureGroupGeographical,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "cells",
		Required:    true,
		Description: "Number of unique geographic cells occupied by usable points.",
	},
	{
		Name:        "operational.minimum_altitude_m",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres",
		Required:    true,
		Description: "Minimum usable aircraft altitude.",
	},
	{
		Name:        "operational.maximum_altitude_m",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres",
		Required:    true,
		Description: "Maximum usable aircraft altitude.",
	},
	{
		Name:        "operational.mean_altitude_m",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres",
		Required:    true,
		Description: "Mean usable aircraft altitude.",
	},
	{
		Name:        "operational.altitude_range_m",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres",
		Required:    true,
		Description: "Difference between maximum and minimum usable altitude.",
	},
	{
		Name:        "operational.mean_velocity_mps",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres_per_second",
		Required:    true,
		Description: "Mean usable ground velocity.",
	},
	{
		Name:        "operational.maximum_velocity_mps",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres_per_second",
		Required:    true,
		Description: "Maximum usable ground velocity.",
	},
	{
		Name:        "operational.mean_absolute_vertical_rate_mps",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres_per_second",
		Required:    true,
		Description: "Mean absolute vertical rate.",
	},
	{
		Name:        "operational.maximum_absolute_vertical_rate_mps",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "metres_per_second",
		Required:    true,
		Description: "Maximum absolute vertical rate.",
	},
	{
		Name:        "operational.heading_change_degrees",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "degrees",
		Required:    true,
		Description: "Cumulative normalized heading change.",
	},
	{
		Name:        "operational.ground_observation_share",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Share of usable observations marked on ground.",
	},
	{
		Name:        "operational.airborne_observation_share",
		Group:       FeatureGroupOperational,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Share of usable observations marked airborne.",
	},
	{
		Name:        "trajectory.point_count",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "points",
		Required:    true,
		Description: "Number of trajectory points used as evidence.",
	},
	{
		Name:        "trajectory.segment_count",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "segments",
		Required:    true,
		Description: "Number of persisted trajectory segments.",
	},
	{
		Name:        "trajectory.coverage_gap_count",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "gaps",
		Required:    true,
		Description: "Number of persisted coverage gaps.",
	},
	{
		Name:        "trajectory.quality_score",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Persisted trajectory quality score.",
	},
	{
		Name:        "trajectory.observed_segment_count",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "segments",
		Required:    true,
		Description: "Number of observed trajectory segments.",
	},
	{
		Name:        "trajectory.interpolated_segment_count",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "segments",
		Required:    true,
		Description: "Number of interpolated trajectory segments.",
	},
	{
		Name:        "trajectory.estimated_segment_count",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "segments",
		Required:    true,
		Description: "Number of estimated trajectory segments.",
	},
	{
		Name:        "trajectory.invalid_segment_count",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeInteger,
		Unit:        "segments",
		Required:    true,
		Description: "Number of invalid trajectory segments.",
	},
	{
		Name:        "trajectory.observed_segment_share",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Share of trajectory segments marked observed.",
	},
	{
		Name:        "trajectory.interpolated_segment_share",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Share of trajectory segments marked interpolated.",
	},
	{
		Name:        "trajectory.estimated_segment_share",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Share of trajectory segments marked estimated.",
	},
	{
		Name:        "trajectory.invalid_segment_share",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Share of trajectory segments marked invalid.",
	},
	{
		Name:        "trajectory.mean_sampling_interval_seconds",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "seconds",
		Required:    true,
		Description: "Mean time between consecutive usable points.",
	},
	{
		Name:        "trajectory.maximum_sampling_gap_seconds",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "seconds",
		Required:    true,
		Description: "Maximum time between consecutive usable points.",
	},
	{
		Name:        "trajectory.coverage_ratio",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Observed temporal coverage relative to the feature window.",
	},
	{
		Name:        "trajectory.path_efficiency_ratio",
		Group:       FeatureGroupTrajectory,
		ValueType:   FeatureValueTypeFloat64,
		Unit:        "ratio",
		Required:    true,
		Description: "Great-circle endpoint distance divided by observed path distance.",
	},
	{
		Name:        "aircraft.registration",
		Group:       FeatureGroupAircraft,
		ValueType:   FeatureValueTypeString,
		Required:    false,
		Description: "Registered aircraft identifier when available.",
	},
	{
		Name:        "aircraft.manufacturer",
		Group:       FeatureGroupAircraft,
		ValueType:   FeatureValueTypeString,
		Required:    false,
		Description: "Aircraft manufacturer when available.",
	},
	{
		Name:        "aircraft.model",
		Group:       FeatureGroupAircraft,
		ValueType:   FeatureValueTypeString,
		Required:    false,
		Description: "Aircraft model when available.",
	},
	{
		Name:        "aircraft.aircraft_type",
		Group:       FeatureGroupAircraft,
		ValueType:   FeatureValueTypeString,
		Required:    false,
		Description: "Aircraft type when available.",
	},
	{
		Name:        "aircraft.airline",
		Group:       FeatureGroupAircraft,
		ValueType:   FeatureValueTypeString,
		Required:    false,
		Description: "Airline or operator when available.",
	},
	{
		Name:        "aircraft.country",
		Group:       FeatureGroupAircraft,
		ValueType:   FeatureValueTypeString,
		Required:    false,
		Description: "Aircraft registration country when available.",
	},
}

func CurrentSchema() Schema {
	return Schema{
		Version: SchemaVersionV1,
		Definitions: append(
			[]FeatureDefinition(nil),
			currentDefinitions...,
		),
	}
}

func DefinitionByName(
	name string,
) (FeatureDefinition, bool) {
	for _, definition := range currentDefinitions {
		if definition.Name == name {
			return definition, true
		}
	}

	return FeatureDefinition{}, false
}
