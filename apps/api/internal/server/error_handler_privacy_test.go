package server

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestServerErrorLogDoesNotRecordRawPath(
	t *testing.T,
) {
	var output bytes.Buffer
	log := slog.New(
		slog.NewJSONHandler(
			&output,
			nil,
		),
	)

	app, err := New(
		Config{
			Logger: log,
		},
	)
	if err != nil {
		t.Fatalf(
			"create server: %v",
			err,
		)
	}

	app.Get(
		"/api/v1/trajectories/:id/failure",
		func(
			*fiber.Ctx,
		) error {
			return errors.New(
				"sensitive implementation failure",
			)
		},
	)

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/api/v1/trajectories/sensitive-aircraft-id/failure?token=secret-value",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute failing request: %v",
			err,
		)
	}
	defer response.Body.Close()

	logged := output.String()
	if !strings.Contains(
		logged,
		`"route":"/api/v1/trajectories/:id/failure"`,
	) {
		t.Fatalf(
			"expected route template in error log, got %s",
			logged,
		)
	}

	for _, forbidden := range []string{
		`"path":`,
		"sensitive-aircraft-id",
		"secret-value",
		"token=",
		"sensitive implementation failure",
	} {
		if strings.Contains(
			logged,
			forbidden,
		) {
			t.Fatalf(
				"error log leaked forbidden value %q: %s",
				forbidden,
				logged,
			)
		}
	}
}
