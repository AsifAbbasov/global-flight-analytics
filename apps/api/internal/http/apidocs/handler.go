package apidocs

import (
	"crypto/sha256"
	_ "embed"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	RootPath       = "/api/docs"
	OpenAPIPath    = "/api/docs/openapi.json"
	JavaScriptPath = "/api/docs/assets/app.js"
	StylesheetPath = "/api/docs/assets/app.css"
)

const documentationContentSecurityPolicy = "default-src 'none'; base-uri 'none'; frame-ancestors 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; form-action 'none'"

//go:embed openapi.json
var openAPISpec []byte

//go:embed assets/index.html
var indexHTML []byte

//go:embed assets/app.js
var applicationJavaScript []byte

//go:embed assets/app.css
var applicationStylesheet []byte

func Register(router fiber.Router) {
	router.Get(RootPath, serveIndex)
	router.Get(RootPath+"/", serveIndex)
	router.Get(OpenAPIPath, serveOpenAPI)
	router.Get(JavaScriptPath, serveJavaScript)
	router.Get(StylesheetPath, serveStylesheet)
}

func serveIndex(c *fiber.Ctx) error {
	setDocumentationHeaders(c)
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Type("html", "utf-8")
	return c.Send(indexHTML)
}

func serveOpenAPI(c *fiber.Ctx) error {
	return serveCacheableAsset(
		c,
		openAPISpec,
		"application/json; charset=utf-8",
	)
}

func serveJavaScript(c *fiber.Ctx) error {
	return serveCacheableAsset(
		c,
		applicationJavaScript,
		"text/javascript; charset=utf-8",
	)
}

func serveStylesheet(c *fiber.Ctx) error {
	return serveCacheableAsset(
		c,
		applicationStylesheet,
		"text/css; charset=utf-8",
	)
}

func serveCacheableAsset(
	c *fiber.Ctx,
	content []byte,
	contentType string,
) error {
	setDocumentationHeaders(c)
	etag := contentETag(content)
	c.Set(fiber.HeaderETag, etag)
	c.Set(fiber.HeaderCacheControl, "public, max-age=300")
	c.Set(fiber.HeaderContentType, contentType)
	if c.Get(fiber.HeaderIfNoneMatch) == etag {
		return c.SendStatus(fiber.StatusNotModified)
	}
	return c.Send(content)
}

func setDocumentationHeaders(c *fiber.Ctx) {
	c.Set("Content-Security-Policy", documentationContentSecurityPolicy)
	c.Set("X-Robots-Tag", "noindex, nofollow")
}

func contentETag(content []byte) string {
	digest := sha256.Sum256(content)
	return fmt.Sprintf("\"%x\"", digest)
}
