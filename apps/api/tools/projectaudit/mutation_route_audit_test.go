package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditMutationRouteProtectionAcceptsProtectedRoutes(
	t *testing.T,
) {
	root := mutationAuditFixture(
		t,
		`package server

func register(v1 router, mutationAuthorization handler, final handler) {
	v1.Post(
		"/items",
		mutationAuthorization,
		final,
	)
	v1.Get(
		"/items",
		final,
	)
}
`,
	)

	var output bytes.Buffer
	if err := auditMutationRouteProtection(
		root,
		&output,
	); err != nil {
		t.Fatalf(
			"audit protected mutation route: %v",
			err,
		)
	}
	if !strings.Contains(
		output.String(),
		"protected_routes=1",
	) {
		t.Fatalf(
			"output = %q",
			output.String(),
		)
	}
}

func TestAuditMutationRouteProtectionRejectsUnprotectedRoutes(
	t *testing.T,
) {
	root := mutationAuditFixture(
		t,
		`package server

func register(v1 router, final handler) {
	v1.Delete(
		"/items/:id",
		final,
	)
}
`,
	)

	err := auditMutationRouteProtection(
		root,
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal(
			"expected unprotected mutation route failure",
		)
	}
	if !strings.Contains(
		err.Error(),
		"must use mutationAuthorization",
	) {
		t.Fatalf(
			"error = %q",
			err,
		)
	}
}

func TestAuditMutationRouteProtectionRejectsFrontendCredentialCoupling(
	t *testing.T,
) {
	root := mutationAuditFixture(
		t,
		`package server

func register(v1 router, mutationAuthorization handler, final handler) {
	v1.Patch(
		"/items/:id",
		mutationAuthorization,
		final,
	)
}
`,
	)
	webFile := filepath.Join(
		root,
		"apps",
		"web",
		"lib",
		"api.ts",
	)
	if err := os.MkdirAll(
		filepath.Dir(webFile),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		webFile,
		[]byte(
			`export const header = "X-Internal-API-Key"`,
		),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	err := auditMutationRouteProtection(
		root,
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal(
			"expected frontend credential coupling failure",
		)
	}
	if !strings.Contains(
		err.Error(),
		"frontend mutation secret separation",
	) {
		t.Fatalf(
			"error = %q",
			err,
		)
	}
}

func mutationAuditFixture(
	t *testing.T,
	serverSource string,
) string {
	t.Helper()

	root := t.TempDir()
	serverFile := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"server",
		"routes.go",
	)
	if err := os.MkdirAll(
		filepath.Dir(serverFile),
		0o755,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		serverFile,
		[]byte(serverSource),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	webDirectory := filepath.Join(
		root,
		"apps",
		"web",
	)
	if err := os.MkdirAll(
		webDirectory,
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	return root
}
