package server

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	internalmiddleware "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/middleware"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/clientidentity"
	"github.com/gofiber/fiber/v2"
)

const testTransportPeerHeader = "X-Test-Transport-Peer"

func TestClientIPResolverSeparatesTrustedClientIdentities(
	t *testing.T,
) {
	app := newClientIdentityRateLimitTestApp(
		t,
		clientidentity.Config{
			Header: clientidentity.HeaderXForwardedFor,
			TrustedProxyRanges: []string{
				"192.0.2.0/24",
			},
		},
	)

	assertRateLimitIdentityStatus(
		t,
		app,
		"192.0.2.10",
		"203.0.113.10",
		fiber.StatusNoContent,
	)
	assertRateLimitIdentityStatus(
		t,
		app,
		"192.0.2.10",
		"203.0.113.11",
		fiber.StatusNoContent,
	)
	assertRateLimitIdentityStatus(
		t,
		app,
		"192.0.2.10",
		"203.0.113.10",
		fiber.StatusTooManyRequests,
	)
}

func TestClientIPResolverIgnoresSpoofedIdentityFromUntrustedRemote(
	t *testing.T,
) {
	app := newClientIdentityRateLimitTestApp(
		t,
		clientidentity.Config{
			Header: clientidentity.HeaderXForwardedFor,
			TrustedProxyRanges: []string{
				"192.0.2.0/24",
			},
		},
	)

	assertRateLimitIdentityStatus(
		t,
		app,
		"198.51.100.10",
		"203.0.113.10",
		fiber.StatusNoContent,
	)
	assertRateLimitIdentityStatus(
		t,
		app,
		"198.51.100.10",
		"203.0.113.11",
		fiber.StatusTooManyRequests,
	)
}

func TestNewRejectsUnsafeTrustedProxyConfiguration(
	t *testing.T,
) {
	app, err := New(
		Config{
			Logger: newDiscardLogger(),
			Protection: ProtectionConfig{
				ClientIPHeader: clientidentity.HeaderXForwardedFor,
			},
		},
	)
	if err == nil {
		t.Fatal(
			"expected trusted proxy configuration error",
		)
	}
	if app != nil {
		t.Fatal(
			"expected nil app for unsafe trusted proxy configuration",
		)
	}
}

func newClientIdentityRateLimitTestApp(
	t *testing.T,
	policyConfig clientidentity.Config,
) *fiber.App {
	t.Helper()

	policy, err := clientidentity.NewPolicy(
		policyConfig,
	)
	if err != nil {
		t.Fatalf(
			"create client identity policy: %v",
			err,
		)
	}

	resolveClientIP := newClientIPResolver(
		policy,
		func(
			c *fiber.Ctx,
		) string {
			return c.Get(
				testTransportPeerHeader,
			)
		},
	)

	limiter, err := internalmiddleware.NewRateLimiter(
		internalmiddleware.RateLimiterConfig{
			MaxRequests: 1,
			Window:      time.Minute,
			KeyGenerator: func(
				c *fiber.Ctx,
			) string {
				return resolveClientIP(
					c,
				)
			},
			LimitReached: rateLimitReached,
		},
	)
	if err != nil {
		t.Fatalf(
			"create rate limiter: %v",
			err,
		)
	}

	app := fiber.New()
	app.Use(
		limiter,
	)
	app.Get(
		"/rate-limit-identity-test",
		func(
			c *fiber.Ctx,
		) error {
			return c.SendStatus(
				fiber.StatusNoContent,
			)
		},
	)

	return app
}

func assertRateLimitIdentityStatus(
	t *testing.T,
	app *fiber.App,
	transportPeer string,
	forwardedFor string,
	expectedStatus int,
) {
	t.Helper()

	request := httptest.NewRequest(
		fiber.MethodGet,
		"/rate-limit-identity-test",
		nil,
	)
	request.Header.Set(
		testTransportPeerHeader,
		transportPeer,
	)
	request.Header.Set(
		clientidentity.HeaderXForwardedFor,
		forwardedFor,
	)

	httpResponse, err := app.Test(
		request,
	)
	if err != nil {
		t.Fatalf(
			"execute rate limit request: %v",
			err,
		)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != expectedStatus {
		body, _ := io.ReadAll(
			httpResponse.Body,
		)
		t.Fatalf(
			"expected status %d, got %d body=%s",
			expectedStatus,
			httpResponse.StatusCode,
			body,
		)
	}
}
