package historicalcursor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregatecontract"
)

const (
	Version              = "historical-aggregate-cursor-v1"
	MaximumEncodedLength = 2048
)

var ErrInvalid = errors.New(
	"historical aggregate HTTP cursor is invalid",
)

type payload struct {
	Version     string `json:"v"`
	WindowEnd   string `json:"we"`
	WindowStart string `json:"ws"`
	AsOfTime    string `json:"ao"`
	ID          string `json:"id"`
}

func Encode(
	cursor historicalaggregatecontract.ListCursor,
) (string, error) {
	normalized, err :=
		historicalaggregatecontract.
			NormalizeListCursor(cursor)
	if err != nil {
		return "", fmt.Errorf(
			"%w: normalize cursor: %v",
			ErrInvalid,
			err,
		)
	}

	encodedPayload, err := json.Marshal(
		payload{
			Version: Version,
			WindowEnd: normalized.WindowEnd.
				Format(time.RFC3339Nano),
			WindowStart: normalized.WindowStart.
				Format(time.RFC3339Nano),
			AsOfTime: normalized.AsOfTime.
				Format(time.RFC3339Nano),
			ID: normalized.ID,
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"%w: encode cursor payload: %v",
			ErrInvalid,
			err,
		)
	}

	encoded := base64.RawURLEncoding.
		EncodeToString(encodedPayload)
	if len(encoded) > MaximumEncodedLength {
		return "", fmt.Errorf(
			"%w: encoded cursor exceeds maximum length",
			ErrInvalid,
		)
	}

	return encoded, nil
}

func Decode(
	value string,
) (
	*historicalaggregatecontract.ListCursor,
	error,
) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil, nil
	}
	if len(normalized) > MaximumEncodedLength {
		return nil, fmt.Errorf(
			"%w: encoded cursor exceeds maximum length",
			ErrInvalid,
		)
	}

	raw, err := base64.RawURLEncoding.
		DecodeString(normalized)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: decode base64: %v",
			ErrInvalid,
			err,
		)
	}

	var decoded payload
	decoder := json.NewDecoder(
		bytes.NewReader(raw),
	)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf(
			"%w: decode JSON: %v",
			ErrInvalid,
			err,
		)
	}
	if err := rejectTrailingJSON(decoder); err != nil {
		return nil, err
	}
	if decoded.Version != Version {
		return nil, fmt.Errorf(
			"%w: unsupported cursor version",
			ErrInvalid,
		)
	}

	windowEnd, err := parseTime(
		decoded.WindowEnd,
		"window end",
	)
	if err != nil {
		return nil, err
	}
	windowStart, err := parseTime(
		decoded.WindowStart,
		"window start",
	)
	if err != nil {
		return nil, err
	}
	asOfTime, err := parseTime(
		decoded.AsOfTime,
		"as-of time",
	)
	if err != nil {
		return nil, err
	}

	cursor, err :=
		historicalaggregatecontract.
			NormalizeListCursor(
				historicalaggregatecontract.ListCursor{
					WindowEnd:   windowEnd,
					WindowStart: windowStart,
					AsOfTime:    asOfTime,
					ID:          decoded.ID,
				},
			)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: normalize decoded cursor: %v",
			ErrInvalid,
			err,
		)
	}

	return &cursor, nil
}

func parseTime(
	value string,
	name string,
) (time.Time, error) {
	parsed, err := time.Parse(
		time.RFC3339Nano,
		strings.TrimSpace(value),
	)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"%w: parse %s: %v",
			ErrInvalid,
			name,
			err,
		)
	}
	return parsed.UTC(), nil
}

func rejectTrailingJSON(
	decoder *json.Decoder,
) error {
	var extra any
	err := decoder.Decode(&extra)
	switch {
	case errors.Is(err, io.EOF):
		return nil
	case err == nil:
		return fmt.Errorf(
			"%w: trailing JSON value",
			ErrInvalid,
		)
	default:
		return fmt.Errorf(
			"%w: trailing JSON data: %v",
			ErrInvalid,
			err,
		)
	}
}
