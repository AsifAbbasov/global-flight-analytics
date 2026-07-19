package server

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/middleware"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRouteIntelligenceMutationFailsClosedWhenAuthorizationIsUnconfigured(
	t *testing.T,
) {
	app, err := New(
		Config{
			DatabasePool:     &pgxpool.Pool{},
			Logger:           newDiscardLogger(),
			OpenMeteoTimeout: 5 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf(
			"create server: %v",
			err,
		)
	}

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodPost,
			"/api/v1/trajectories/example/route-intelligence",
			nil,
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode !=
		fiber.StatusServiceUnavailable {
		t.Fatalf(
			"status = %d, want %d",
			response.StatusCode,
			fiber.StatusServiceUnavailable,
		)
	}
	assertMutationRouteResponseContains(
		t,
		response.Body,
		middleware.
			MutationAuthenticationUnavailableCode,
	)
}

func TestRouteIntelligenceMutationRejectsInvalidKeyBeforeHandler(
	t *testing.T,
) {
	key := strings.Repeat(
		"route-intelligence-key-",
		2,
	)
	app, err := New(
		Config{
			DatabasePool:     &pgxpool.Pool{},
			Logger:           newDiscardLogger(),
			OpenMeteoTimeout: 5 * time.Second,
			Protection: ProtectionConfig{
				MutationKeyDigest: internalapikey.
					DigestCandidate(key),
				MutationKeyConfigured: true,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create server: %v",
			err,
		)
	}

	request := httptest.NewRequest(
		fiber.MethodPost,
		"/api/v1/trajectories/example/route-intelligence",
		nil,
	)
	request.Header.Set(
		internalapikey.HeaderName,
		strings.Repeat(
			"incorrect-key-",
			3,
		),
	)

	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode !=
		fiber.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			response.StatusCode,
			fiber.StatusUnauthorized,
		)
	}
	assertMutationRouteResponseContains(
		t,
		response.Body,
		middleware.
			MutationAuthenticationRequiredCode,
	)
}

func assertMutationRouteResponseContains(
	t *testing.T,
	body io.Reader,
	expected string,
) {
	t.Helper()

	content, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(
		string(content),
		expected,
	) {
		t.Fatalf(
			"response %q does not contain %q",
			content,
			expected,
		)
	}
}
