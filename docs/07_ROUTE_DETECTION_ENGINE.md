# DOCUMENT 07

# ROUTE DETECTION ENGINE

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

The Route Detection Engine is an analytical module of the Global Flight Analytics platform.

The module determines a probable aircraft route from open aviation data.

The system does not use:

- official flight plans;
- air traffic control data;
- internal airline data.

The system uses:

- coordinates;
- velocity;
- heading;
- altitude;
- vertical rate;
- airport locations;
- observation history.

---

# 2. Goal

Determine:

- probable departure airport;
- probable destination airport;
- confidence level;
- numeric confidence score.

---

# 3. Architectural Position

The Route Detection Engine is located between:

```text
Flight States

↓

Route Detection Engine

↓

Route Predictions

↓

Analytics
```

---

Input data comes from:

```text
flight_states
airports
airport_profiles
```

---

The result is stored in:

```text
route_predictions
```

---

# 4. Input Data

Route determination uses:

## Aircraft Information

- ICAO24;
- Callsign.

---

## Position Information

- Latitude;
- Longitude.

---

## Flight Dynamics

- Velocity;
- Heading;
- Vertical Rate;
- Altitude.

---

## Airport Information

- airport coordinates;
- airport geofence;
- airport type.

---

# 5. Output Data

The algorithm produces a:

```text
Route Prediction
```

Contains:

- Origin Airport;
- Destination Airport;
- Confidence Level;
- Confidence Score;
- Detection Method;
- Calculated At.

---

# 6. Confidence Levels

## HIGH

The route is supported by several independent indicators.

Examples:

- the aircraft was observed in the departure area;
- the aircraft was observed in the arrival area;
- the route completed;
- the full flight lifecycle was observed.

---

## MEDIUM

The route is probable.

Examples:

- the aircraft is descending;
- the heading points toward the airport;
- distance to the airport decreases consistently.

---

## LOW

Insufficient data is available.

Examples:

- the aircraft first appeared in the air;
- observation history is unavailable;
- observation began after takeoff.

---

# 7. Confidence Score

A numeric score is also calculated.

Range:

```text
0.00 – 1.00
```

---

Recommended values:

```text
0.80 – 1.00 → HIGH

0.50 – 0.79 → MEDIUM

0.00 – 0.49 → LOW
```

---

# 8. Airport Geofencing

A virtual geofence is created for every airport.

---

MVP Parameters:

```text
Radius: 30 km
```

---

The algorithm also considers:

- aircraft altitude;
- aircraft velocity;
- vertical rate.

---

Goal:

Determine:

- takeoff;
- arrival;
- presence near an airport.

---

# 9. Origin Detection

The system determines the probable departure airport.

---

Indicators:

- the aircraft appeared inside the geofence;
- velocity increases;
- altitude increases;
- vertical rate is positive.

---

Additional indicators:

- extended presence near the airport before takeoff;
- movement direction matches the runway axis.

---

# 10. Destination Detection

The system determines the probable destination airport.

---

Indicators:

- the aircraft is descending;
- velocity decreases;
- distance to the airport decreases;
- heading points toward the airport.

---

Additional indicators:

- crossing the airport geofence;
- on-ground state after descent.

---

# 11. Route Lifecycle

Every route passes through a lifecycle.

```text
Unknown

↓

Detected

↓

Predicted

↓

Confirmed

↓

Archived
```

---

## Unknown

No route is available.

---

## Detected

Initial route indicators exist.

---

## Predicted

The route was calculated by the algorithm.

---

## Confirmed

Confidence is high.

---

## Archived

The flight has ended.

---

# 12. Failure Cases

The system must operate correctly when:

- the aircraft disappears;
- data is interrupted;
- no airport is found;
- no route can be determined;
- the route changes during observation.

---

In these cases, the confidence level must be reduced.

---

# 13. Route Prediction Rules

Route Prediction is:

```text
Inferred Data
```

---

Route Prediction is not:

- an official route;
- a flight plan;
- confirmed airline information.

---

The system must never display a predicted route as fact.

---

# 14. Storage Model

All results are stored in:

```text
route_predictions
```

---

Primary fields:

- flight_id;
- aircraft_id;
- origin_airport_id;
- destination_airport_id;
- confidence_level;
- confidence_score;
- method_name;
- calculated_at.

---

# 15. Future Improvements

After the MVP, the following may be added:

- machine learning;
- historical route analysis;
- analysis of typical airline routes;
- seasonal patterns;
- probabilistic models.

---

# 16. Summary

The Route Detection Engine is an analytical platform module.

It uses open aviation data to determine the probable route of an aircraft.

All results:

- include a confidence level;
- include a numeric confidence score;
- are stored as Inferred Data;
- are never presented to the user as confirmed fact.
