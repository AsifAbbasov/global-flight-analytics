package interactiongraph

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"
)

const fingerprintVersion = "airborne-interaction-graph-fingerprint-v1"

func inputFingerprint(result Result) string {
	hasher := sha256.New()
	writeFingerprintPart(hasher, fingerprintVersion)
	writeFingerprintPart(hasher, string(result.SchemaVersion))
	writeFingerprintPart(hasher, strings.TrimSpace(result.RegionCode))
	writeFingerprintTime(hasher, result.AsOfTime)

	nodes := append([]Node(nil), result.Nodes...)
	sort.Slice(nodes, func(left int, right int) bool {
		return nodes[left].ID < nodes[right].ID
	})
	for _, node := range nodes {
		writeFingerprintPart(hasher, "node")
		writeFingerprintPart(hasher, node.ID)
		writeFingerprintPart(hasher, node.TrajectoryID)
		writeFingerprintPart(hasher, node.FlightID)
		writeFingerprintPart(hasher, node.AircraftID)
		writeFingerprintPart(hasher, node.ICAO24)
		writeFingerprintPart(hasher, node.Callsign)
		writeFingerprintFloat(hasher, node.Latitude)
		writeFingerprintFloat(hasher, node.Longitude)
		writeFingerprintOptionalFloat(hasher, node.AltitudeMeters)
		writeFingerprintPart(hasher, string(node.AltitudeReference))
		writeFingerprintFloat(hasher, node.VelocityMetersPerSecond)
		writeFingerprintFloat(hasher, node.HeadingDegrees)
		writeFingerprintFloat(hasher, node.VerticalRateMetersPerSecond)
		writeFingerprintTime(hasher, node.ObservedAt)
		writeFingerprintPart(hasher, node.SourceName)
		writeFingerprintFloat(hasher, node.QualityScore)
	}

	edges := append([]Edge(nil), result.Edges...)
	sort.Slice(edges, func(left int, right int) bool {
		return edges[left].ID < edges[right].ID
	})
	for _, edge := range edges {
		writeFingerprintPart(hasher, "edge")
		writeFingerprintPart(hasher, edge.ID)
		writeFingerprintPart(hasher, edge.SourceNodeID)
		writeFingerprintPart(hasher, edge.TargetNodeID)
		writeFingerprintPart(hasher, string(edge.Kind))
		writeFingerprintFloat(
			hasher,
			edge.HorizontalDistanceKilometers,
		)
		writeFingerprintOptionalFloat(
			hasher,
			edge.VerticalSeparationMeters,
		)
		writeFingerprintPart(
			hasher,
			strconv.FormatInt(
				edge.ObservationTimeDifference.Nanoseconds(),
				10,
			),
		)
		writeFingerprintTime(hasher, edge.EvaluatedAt)
		writeFingerprintPart(hasher, edge.SourceName)
		writeFingerprintFloat(hasher, edge.ConfidenceScore)
	}

	return "sha256:" + hex.EncodeToString(hasher.Sum(nil))
}

func writeFingerprintPart(
	hasher interface{ Write([]byte) (int, error) },
	value string,
) {
	_, _ = hasher.Write([]byte(value))
	_, _ = hasher.Write([]byte{0})
}

func writeFingerprintFloat(
	hasher interface{ Write([]byte) (int, error) },
	value float64,
) {
	writeFingerprintPart(
		hasher,
		strconv.FormatFloat(value, 'g', -1, 64),
	)
}

func writeFingerprintOptionalFloat(
	hasher interface{ Write([]byte) (int, error) },
	value *float64,
) {
	if value == nil {
		writeFingerprintPart(hasher, "nil")
		return
	}
	writeFingerprintFloat(hasher, *value)
}

func writeFingerprintTime(
	hasher interface{ Write([]byte) (int, error) },
	value time.Time,
) {
	writeFingerprintPart(
		hasher,
		value.UTC().Format(time.RFC3339Nano),
	)
}
