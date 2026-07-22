package common

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadHTTPResponseBodyAcceptsExactLimit(
	t *testing.T,
) {
	response := &http.Response{
		Body:          io.NopCloser(strings.NewReader("1234")),
		ContentLength: 4,
	}
	defer response.Body.Close()

	content, err := ReadHTTPResponseBody(
		response,
		"test-provider",
		4,
	)
	if err != nil {
		t.Fatalf("read exact-limit response: %v", err)
	}
	if string(content) != "1234" {
		t.Fatalf("content = %q, want 1234", content)
	}
}

func TestReadHTTPResponseBodyRejectsDeclaredOversize(
	t *testing.T,
) {
	response := &http.Response{
		Body:          io.NopCloser(strings.NewReader("")),
		ContentLength: 5,
	}
	defer response.Body.Close()

	_, err := ReadHTTPResponseBody(
		response,
		"test-provider",
		4,
	)
	assertProviderResponseTooLarge(t, err, 5)
}

func TestReadHTTPResponseBodyRejectsStreamingOversize(
	t *testing.T,
) {
	response := &http.Response{
		Body:          io.NopCloser(strings.NewReader("12345")),
		ContentLength: -1,
	}
	defer response.Body.Close()

	_, err := ReadHTTPResponseBody(
		response,
		"test-provider",
		4,
	)
	assertProviderResponseTooLarge(t, err, 0)
}

func TestDecodeJSONHTTPResponseUsesBoundedBody(
	t *testing.T,
) {
	response := &http.Response{
		Body: io.NopCloser(
			strings.NewReader(`{"value":"ok"}`),
		),
		ContentLength: int64(len(`{"value":"ok"}`)),
	}
	defer response.Body.Close()

	var payload struct {
		Value string `json:"value"`
	}
	if err := DecodeJSONHTTPResponse(
		response,
		"test-provider",
		64,
		&payload,
	); err != nil {
		t.Fatalf("decode bounded JSON response: %v", err)
	}
	if payload.Value != "ok" {
		t.Fatalf("value = %q, want ok", payload.Value)
	}
}

func assertProviderResponseTooLarge(
	t *testing.T,
	err error,
	declaredBytes int64,
) {
	t.Helper()

	if !errors.Is(err, ErrProviderResponseTooLarge) {
		t.Fatalf(
			"expected ErrProviderResponseTooLarge, got %v",
			err,
		)
	}

	var sizeError *ProviderResponseTooLargeError
	if !errors.As(err, &sizeError) {
		t.Fatalf(
			"expected ProviderResponseTooLargeError, got %T",
			err,
		)
	}
	if sizeError.Provider != "test-provider" {
		t.Fatalf(
			"provider = %q, want test-provider",
			sizeError.Provider,
		)
	}
	if sizeError.LimitBytes != 4 {
		t.Fatalf(
			"limit = %d, want 4",
			sizeError.LimitBytes,
		)
	}
	if sizeError.DeclaredContentBytes != declaredBytes {
		t.Fatalf(
			"declared bytes = %d, want %d",
			sizeError.DeclaredContentBytes,
			declaredBytes,
		)
	}
}
