package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrProviderResponseTooLarge = errors.New(
	"provider response exceeds configured size limit",
)

type ProviderResponseTooLargeError struct {
	Provider             string
	LimitBytes           int64
	DeclaredContentBytes int64
}

func (err *ProviderResponseTooLargeError) Error() string {
	if err == nil {
		return ErrProviderResponseTooLarge.Error()
	}

	provider := strings.TrimSpace(err.Provider)
	if provider == "" {
		provider = "unknown"
	}

	if err.DeclaredContentBytes > 0 {
		return fmt.Sprintf(
			"provider response exceeds configured size limit: provider=%s limit_bytes=%d declared_content_bytes=%d",
			provider,
			err.LimitBytes,
			err.DeclaredContentBytes,
		)
	}

	return fmt.Sprintf(
		"provider response exceeds configured size limit: provider=%s limit_bytes=%d",
		provider,
		err.LimitBytes,
	)
}

func (*ProviderResponseTooLargeError) Unwrap() error {
	return ErrProviderResponseTooLarge
}

func ReadHTTPResponseBody(
	response *http.Response,
	provider string,
	maxBytes int64,
) ([]byte, error) {
	if response == nil {
		return nil, errors.New("provider HTTP response is required")
	}
	if response.Body == nil {
		return nil, errors.New("provider HTTP response body is required")
	}
	if maxBytes <= 0 {
		return nil, errors.New("provider HTTP response size limit must be greater than zero")
	}

	if response.ContentLength > maxBytes {
		return nil, &ProviderResponseTooLargeError{
			Provider:             provider,
			LimitBytes:           maxBytes,
			DeclaredContentBytes: response.ContentLength,
		}
	}

	content, err := io.ReadAll(
		io.LimitReader(
			response.Body,
			maxBytes+1,
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"read provider HTTP response body: %w",
			err,
		)
	}
	if int64(len(content)) > maxBytes {
		return nil, &ProviderResponseTooLargeError{
			Provider:   provider,
			LimitBytes: maxBytes,
		}
	}

	return content, nil
}

func DecodeJSONHTTPResponse(
	response *http.Response,
	provider string,
	maxBytes int64,
	target any,
) error {
	if target == nil {
		return errors.New("provider JSON response target is required")
	}

	content, err := ReadHTTPResponseBody(
		response,
		provider,
		maxBytes,
	)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(
		content,
		target,
	); err != nil {
		return fmt.Errorf(
			"decode provider JSON response: %w",
			err,
		)
	}

	return nil
}
