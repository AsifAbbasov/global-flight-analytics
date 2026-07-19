package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/transponderalert"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/gofiber/fiber/v2"
)

type transponderEvidenceReaderStub struct {
	result transponderalert.LatestEvidence
	err    error
}

func (stub transponderEvidenceReaderStub) GetLatest(
	_ context.Context,
	_ string,
) (transponderalert.LatestEvidence, error) {
	return stub.result, stub.err
}

func TestTransponderEvidenceHandlerReturnsEvidenceOnlyResponse(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		19,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	handler := NewTransponderEvidenceHandler(
		transponderEvidenceReaderStub{
			result: transponderalert.LatestEvidence{
				Evidence: transponderalert.Evidence{
					SchemaVersion: transponderalert.SchemaVersion,
					Fingerprint:   "sha256:test",
					ICAO24:        "4A001A",
					SquawkCode:    "7700",
					Kind: transponderalert.
						KindGeneralEmergencyCode,
					Label: "Observed general emergency transponder code",
					Strength: transponderalert.
						StrengthSingleObservation,
					FirstObservedAt:  now,
					LastObservedAt:   now,
					AsOfTime:         now,
					ObservationCount: 1,
					SourceNames: []string{
						"opensky",
					},
					MaximumClaimStrength: "observed_transponder_code_only",
					Limitations: []string{
						"research only",
					},
				},
				FreshnessStatus: transponderalert.FreshnessRecent,
				MaximumFreshAge: 5 * time.Minute,
				Confidence: transponderalert.Confidence{
					Level: transponderalert.ConfidenceLimited,
					Reasons: []string{
						"single observation",
					},
				},
				EvidenceOnly:       true,
				ConfirmedEmergency: false,
				OperationalAlert:   false,
			},
		},
	)
	app := fiber.New()
	app.Get(
		"/api/v1/aircraft/:icao24/transponder-evidence/latest",
		handler.GetLatest,
	)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/aircraft/4A001A/transponder-evidence/latest",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"status = %d, want 200",
			result.StatusCode,
		)
	}
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	text := string(body)
	for _, fragment := range []string{
		`"success":true`,
		`"evidence_only":true`,
		`"confirmed_emergency":false`,
		`"operational_alert":false`,
		`"observed_transponder_code":"7700"`,
		`"maximum_claim_strength":"observed_transponder_code_only"`,
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf(
				"response does not contain %q: %s",
				fragment,
				text,
			)
		}
	}
}

func TestTransponderEvidenceHandlerMapsSemanticErrors(
	t *testing.T,
) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		code       string
	}{
		{
			name:       "invalid ICAO24",
			err:        transponderalert.ErrICAO24Invalid,
			statusCode: fiber.StatusBadRequest,
			code:       "INVALID_TRANSPONDER_EVIDENCE_ICAO24",
		},
		{
			name:       "flight state missing",
			err:        flightstate.ErrNotFound,
			statusCode: fiber.StatusNotFound,
			code:       "TRANSPONDER_EVIDENCE_SOURCE_NOT_FOUND",
		},
		{
			name:       "special code evidence missing",
			err:        transponderalert.ErrEvidenceNotFound,
			statusCode: fiber.StatusNotFound,
			code:       "TRANSPONDER_EVIDENCE_NOT_FOUND",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewTransponderEvidenceHandler(
				transponderEvidenceReaderStub{
					err: test.err,
				},
			)
			app := fiber.New()
			app.Get("/evidence", handler.GetLatest)

			result, err := app.Test(
				httptest.NewRequest(
					http.MethodGet,
					"/evidence",
					nil,
				),
			)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer result.Body.Close()

			if result.StatusCode != test.statusCode {
				t.Fatalf(
					"status = %d, want %d",
					result.StatusCode,
					test.statusCode,
				)
			}
			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if !strings.Contains(
				string(body),
				test.code,
			) {
				t.Fatalf(
					"response does not contain %q: %s",
					test.code,
					string(body),
				)
			}
		})
	}
}

func TestTransponderEvidenceHandlerRejectsMissingReader(
	t *testing.T,
) {
	handler := NewTransponderEvidenceHandler(nil)
	app := fiber.New()
	app.Get("/evidence", handler.GetLatest)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/evidence",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode !=
		fiber.StatusServiceUnavailable {
		t.Fatalf(
			"status = %d, want 503",
			result.StatusCode,
		)
	}
}

func TestTransponderEvidenceHandlerMapsUnexpectedError(
	t *testing.T,
) {
	handler := NewTransponderEvidenceHandler(
		transponderEvidenceReaderStub{
			err: errors.New("unexpected"),
		},
	)
	app := fiber.New()
	app.Get("/evidence", handler.GetLatest)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/evidence",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer result.Body.Close()

	if result.StatusCode !=
		fiber.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want 500",
			result.StatusCode,
		)
	}
}
