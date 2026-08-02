package observability

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
	"github.com/gofiber/fiber/v2"
)

const testMetricsKey = "0123456789abcdef0123456789abcdef"

func TestHTTPMiddlewareUsesRouteTemplateInsteadOfRawPath(
	t *testing.T,
) {
	registry := NewRegistry(BuildInfo{})
	app := fiber.New()
	app.Use(HTTPMiddleware(registry))
	app.Get(
		"/api/v1/aircraft/:icao24",
		func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(fiber.StatusNoContent)
		},
	)

	for _, value := range []string{"abc123", "def456"} {
		response, err := app.Test(
			httptest.NewRequest(
				fiber.MethodGet,
				"/api/v1/aircraft/"+value,
				nil,
			),
		)
		if err != nil {
			t.Fatalf("execute request: %v", err)
		}
		if response.StatusCode != fiber.StatusNoContent {
			t.Fatalf("unexpected status: %d", response.StatusCode)
		}
	}

	output := registry.Render(context.Background())
	if !strings.Contains(
		output,
		`route="/api/v1/aircraft/:icao24",status_class="2xx"} 2`,
	) {
		t.Fatalf("expected aggregated route template\n%s", output)
	}
	if strings.Contains(output, "abc123") || strings.Contains(output, "def456") {
		t.Fatalf("raw path parameter leaked into metrics\n%s", output)
	}
}

func TestMetricsAuthorizationAndFiberHandler(
	t *testing.T,
) {
	digest := internalapikey.DigestCandidate(testMetricsKey)
	authorization, err := NewAuthorization(
		AuthorizationConfig{
			ExpectedDigest: digest,
			Configured:     true,
		},
	)
	if err != nil {
		t.Fatalf("create metrics authorization: %v", err)
	}

	registry := NewRegistry(BuildInfo{})
	app := fiber.New()
	app.Get(MetricsPath, authorization, FiberHandler(registry))

	unauthorizedResponse, err := app.Test(
		httptest.NewRequest(fiber.MethodGet, MetricsPath, nil),
	)
	if err != nil {
		t.Fatalf("execute unauthorized request: %v", err)
	}
	if unauthorizedResponse.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthorizedResponse.StatusCode)
	}

	authorizedRequest := httptest.NewRequest(fiber.MethodGet, MetricsPath, nil)
	authorizedRequest.Header.Set(internalapikey.HeaderName, testMetricsKey)
	authorizedResponse, err := app.Test(authorizedRequest)
	if err != nil {
		t.Fatalf("execute authorized request: %v", err)
	}
	if authorizedResponse.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", authorizedResponse.StatusCode)
	}
	if contentType := authorizedResponse.Header.Get(fiber.HeaderContentType); !strings.Contains(
		contentType,
		"text/plain",
	) {
		t.Fatalf("unexpected content type: %q", contentType)
	}
}
