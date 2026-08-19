package apidocs

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestDocumentationSurfaceServesEmbeddedContractAndAssets(t *testing.T) {
	app := fiber.New()
	Register(app)

	for _, route := range []struct {
		path        string
		contentType string
	}{
		{path: RootPath, contentType: "text/html"},
		{path: OpenAPIPath, contentType: "application/json"},
		{path: JavaScriptPath, contentType: "text/javascript"},
		{path: StylesheetPath, contentType: "text/css"},
	} {
		response, err := app.Test(httptest.NewRequest("GET", route.path, nil))
		if err != nil {
			t.Fatalf("request %s: %v", route.path, err)
		}
		if response.StatusCode != fiber.StatusOK {
			t.Fatalf("expected %s status 200, got %d", route.path, response.StatusCode)
		}
		if !strings.HasPrefix(response.Header.Get("Content-Type"), route.contentType) {
			t.Fatalf("expected %s content type %s, got %s", route.path, route.contentType, response.Header.Get("Content-Type"))
		}
		if response.Header.Get("X-Robots-Tag") != "noindex, nofollow" {
			t.Fatalf("expected noindex header for %s", route.path)
		}
		_ = response.Body.Close()
	}
}

func TestEmbeddedOpenAPIContainsCompletePublicOperationSurface(t *testing.T) {
	var spec struct {
		OpenAPI string                            `json:"openapi"`
		Paths   map[string]map[string]interface{} `json:"paths"`
	}
	if err := json.Unmarshal(openAPISpec, &spec); err != nil {
		t.Fatalf("parse embedded OpenAPI: %v", err)
	}
	if spec.OpenAPI != "3.1.0" {
		t.Fatalf("expected OpenAPI 3.1.0, got %q", spec.OpenAPI)
	}

	operations := 0
	for _, pathItem := range spec.Paths {
		for method := range pathItem {
			switch strings.ToUpper(method) {
			case "GET", "POST", "PUT", "PATCH", "DELETE":
				operations++
			}
		}
	}
	if operations != 39 {
		t.Fatalf("expected 39 embedded OpenAPI operations, got %d", operations)
	}
}

func TestOpenAPIAssetUsesStrongETag(t *testing.T) {
	app := fiber.New()
	Register(app)

	first, err := app.Test(httptest.NewRequest("GET", OpenAPIPath, nil))
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	etag := first.Header.Get("ETag")
	body, err := io.ReadAll(first.Body)
	if err != nil {
		t.Fatalf("read first response: %v", err)
	}
	_ = first.Body.Close()
	if etag == "" {
		t.Fatal("expected ETag")
	}
	if string(body) != string(openAPISpec) {
		t.Fatal("served OpenAPI differs from embedded contract")
	}

	request := httptest.NewRequest("GET", OpenAPIPath, nil)
	request.Header.Set("If-None-Match", etag)
	second, err := app.Test(request)
	if err != nil {
		t.Fatalf("conditional request: %v", err)
	}
	defer second.Body.Close()
	if second.StatusCode != fiber.StatusNotModified {
		t.Fatalf("expected 304, got %d", second.StatusCode)
	}
}

func TestOpenAPIAssetAcceptsWeakAndListedIfNoneMatchValidators(t *testing.T) {
	app := fiber.New()
	Register(app)

	first, err := app.Test(
		httptest.NewRequest(
			"GET",
			OpenAPIPath,
			nil,
		),
	)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	etag := first.Header.Get("ETag")
	_ = first.Body.Close()
	if etag == "" {
		t.Fatal("expected ETag")
	}

	for _, headerValue := range []string{
		"W/" + etag,
		`"different", W/` + etag,
		"*",
	} {
		request := httptest.NewRequest(
			"GET",
			OpenAPIPath,
			nil,
		)
		request.Header.Set(
			"If-None-Match",
			headerValue,
		)

		response, requestErr := app.Test(request)
		if requestErr != nil {
			t.Fatalf(
				"conditional request %q: %v",
				headerValue,
				requestErr,
			)
		}
		_ = response.Body.Close()
		if response.StatusCode !=
			fiber.StatusNotModified {
			t.Fatalf(
				"expected %q to return 304, got %d",
				headerValue,
				response.StatusCode,
			)
		}
	}

	mismatch := httptest.NewRequest(
		"GET",
		OpenAPIPath,
		nil,
	)
	mismatch.Header.Set(
		"If-None-Match",
		`W/"different"`,
	)
	mismatchResponse, err := app.Test(mismatch)
	if err != nil {
		t.Fatalf("mismatch request: %v", err)
	}
	defer mismatchResponse.Body.Close()
	if mismatchResponse.StatusCode !=
		fiber.StatusOK {
		t.Fatalf(
			"expected mismatched validator to return 200, got %d",
			mismatchResponse.StatusCode,
		)
	}
}

func TestBrowserExplorerDoesNotCollectProtectedMutationCredentials(t *testing.T) {
	source := string(applicationJavaScript)
	for _, forbidden := range []string{"localStorage", "sessionStorage", "document.cookie", "internalApiKeyInput"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("browser explorer contains forbidden credential surface %q", forbidden)
		}
	}
	if !strings.Contains(source, "Protected mutation") {
		t.Fatal("browser explorer must explain protected mutation boundary")
	}
}
