package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
	"github.com/gofiber/fiber/v2"
)

func TestMutationAuthorizationRejectsUnconfiguredService(
	t *testing.T,
) {
	handler, err :=
		NewMutationAuthorization(
			MutationAuthorizationConfig{},
		)
	if err != nil {
		t.Fatal(err)
	}

	app := newMutationAuthorizationTestApp(
		handler,
	)
	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodPost,
			"/mutation",
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
	assertResponseContains(
		t,
		response.Body,
		MutationAuthenticationUnavailableCode,
	)
}

func TestMutationAuthorizationRejectsMissingAndInvalidKeys(
	t *testing.T,
) {
	key := strings.Repeat(
		"stage14-secret-",
		3,
	)
	handler, err :=
		NewMutationAuthorization(
			MutationAuthorizationConfig{
				ExpectedDigest: internalapikey.
					DigestCandidate(key),
				Configured: true,
			},
		)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name string
		key  string
	}{
		{
			name: "missing",
		},
		{
			name: "invalid",
			key: strings.Repeat(
				"invalid-secret-",
				3,
			),
		},
	} {
		t.Run(
			test.name,
			func(t *testing.T) {
				app :=
					newMutationAuthorizationTestApp(
						handler,
					)
				request := httptest.NewRequest(
					fiber.MethodPost,
					"/mutation",
					nil,
				)
				if test.key != "" {
					request.Header.Set(
						internalapikey.
							HeaderName,
						test.key,
					)
				}

				response, requestErr :=
					app.Test(request)
				if requestErr != nil {
					t.Fatal(requestErr)
				}
				if response.StatusCode !=
					fiber.StatusUnauthorized {
					t.Fatalf(
						"status = %d, want %d",
						response.StatusCode,
						fiber.StatusUnauthorized,
					)
				}
				if actual :=
					response.Header.Get(
						fiber.HeaderCacheControl,
					); actual != "no-store" {
					t.Fatalf(
						"cache control = %q",
						actual,
					)
				}
				assertResponseContains(
					t,
					response.Body,
					MutationAuthenticationRequiredCode,
				)
			},
		)
	}
}

func TestMutationAuthorizationAllowsCorrectKey(
	t *testing.T,
) {
	key := strings.Repeat(
		"stage14-secret-",
		3,
	)
	handler, err :=
		NewMutationAuthorization(
			MutationAuthorizationConfig{
				ExpectedDigest: internalapikey.
					DigestCandidate(key),
				Configured: true,
			},
		)
	if err != nil {
		t.Fatal(err)
	}

	app := newMutationAuthorizationTestApp(
		handler,
	)
	request := httptest.NewRequest(
		fiber.MethodPost,
		"/mutation",
		nil,
	)
	request.Header.Set(
		internalapikey.HeaderName,
		key,
	)

	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode !=
		fiber.StatusNoContent {
		t.Fatalf(
			"status = %d, want %d",
			response.StatusCode,
			fiber.StatusNoContent,
		)
	}
}

func TestMutationAuthorizationRejectsConfiguredZeroDigest(
	t *testing.T,
) {
	handler, err :=
		NewMutationAuthorization(
			MutationAuthorizationConfig{
				Configured: true,
			},
		)
	if err == nil {
		t.Fatal(
			"expected zero digest error",
		)
	}
	if handler != nil {
		t.Fatal(
			"expected nil handler",
		)
	}
}

func newMutationAuthorizationTestApp(
	authorization fiber.Handler,
) *fiber.App {
	app := fiber.New()
	app.Post(
		"/mutation",
		authorization,
		func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(
				fiber.StatusNoContent,
			)
		},
	)
	return app
}

func assertResponseContains(
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
