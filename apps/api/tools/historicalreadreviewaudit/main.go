package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type requirement struct {
	path      string
	fragments []string
}

func main() {
	strict := flag.Bool("strict", false, "fail when a Historical Read hardening contract is absent")
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalread/postgres_contracts.go",
			fragments: []string{
				"pgx.RepeatableRead",
				"pgx.ReadOnly",
				"BeginSnapshot",
			},
		},
		{
			path: "internal/historicalintelligence/historicalread/postgres.go",
			fragments: []string{
				"last_seen_at > $1",
				"end_time > $1",
				"DISTINCT ON (result.trajectory_id)",
				"trajectory.start_time < $2",
				"result.as_of_time <= $3",
				"cumulative_payload_bytes <= $5",
				"COUNT(*) OVER ()",
				"ErrContextRequired",
			},
		},
		{
			path: "internal/historicalintelligence/historicalread/contracts.go",
			fragments: []string{
				"historical-read-repository-v2",
				"RoutePayloadByteLimit",
				"FlightMatchedCount",
				"RouteMatchedCount",
				"RepresentedCoverage",
			},
		},
		{
			path: "../../database/migrations/028_harden_historical_read_snapshot.sql",
			fragments: []string{
				"historical_read_flight_versions",
				"historical_read_trajectory_versions",
				"capture_historical_read_flight_version",
				"flight_states_historical_read_idx",
				"flight_route_results_historical_read_idx",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range requirements {
		content, err := os.ReadFile(filepath.Clean(item.path))
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", item.path, err))
			continue
		}
		text := string(content)
		for _, fragment := range item.fragments {
			if !strings.Contains(text, fragment) {
				failures = append(
					failures,
					fmt.Sprintf("%s misses %q", item.path, fragment),
				)
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Historical read review audit: PASS")
		return
	}

	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Historical read review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}
