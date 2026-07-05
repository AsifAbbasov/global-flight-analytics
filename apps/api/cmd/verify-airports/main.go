package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	conn, err := pgx.Connect(
		ctx,
		databaseURL,
	)
	if err != nil {
		log.Fatalf(
			"connect postgres: %v",
			err,
		)
	}
	defer conn.Close(ctx)

	fmt.Println("airport counts by country:")

	rows, err := conn.Query(
		ctx,
		`
			SELECT
				source_country_code,
				COUNT(*)
			FROM airports
			WHERE source_name = 'ourairports'
			GROUP BY source_country_code
			ORDER BY source_country_code;
		`,
	)
	if err != nil {
		log.Fatalf(
			"query airport counts by country: %v",
			err,
		)
	}

	for rows.Next() {
		var countryCode string
		var airportCount int64

		if err := rows.Scan(
			&countryCode,
			&airportCount,
		); err != nil {
			rows.Close()

			log.Fatalf(
				"scan airport count by country: %v",
				err,
			)
		}

		fmt.Printf(
			"country=%s airports=%d\n",
			countryCode,
			airportCount,
		)
	}

	if err := rows.Err(); err != nil {
		rows.Close()

		log.Fatalf(
			"iterate airport counts by country: %v",
			err,
		)
	}

	rows.Close()

	var totalAirports int64

	err = conn.QueryRow(
		ctx,
		`
			SELECT COUNT(*)
			FROM airports
			WHERE source_name = 'ourairports';
		`,
	).Scan(
		&totalAirports,
	)
	if err != nil {
		log.Fatalf(
			"query total imported airports: %v",
			err,
		)
	}

	var duplicateSourceIdentities int64

	err = conn.QueryRow(
		ctx,
		`
			SELECT COUNT(*)
			FROM (
				SELECT
					source_name,
					source_ident
				FROM airports
				WHERE source_name = 'ourairports'
				GROUP BY
					source_name,
					source_ident
				HAVING COUNT(*) > 1
			) AS duplicates;
		`,
	).Scan(
		&duplicateSourceIdentities,
	)
	if err != nil {
		log.Fatalf(
			"query duplicate source identities: %v",
			err,
		)
	}

	var duplicateICAOCodes int64

	err = conn.QueryRow(
		ctx,
		`
			SELECT COUNT(*)
			FROM (
				SELECT icao_code
				FROM airports
				WHERE icao_code IS NOT NULL
				GROUP BY icao_code
				HAVING COUNT(*) > 1
			) AS duplicates;
		`,
	).Scan(
		&duplicateICAOCodes,
	)
	if err != nil {
		log.Fatalf(
			"query duplicate ICAO codes: %v",
			err,
		)
	}

	fmt.Printf(
		"total_ourairports=%d duplicate_source_identities=%d duplicate_icao_codes=%d\n",
		totalAirports,
		duplicateSourceIdentities,
		duplicateICAOCodes,
	)
}
