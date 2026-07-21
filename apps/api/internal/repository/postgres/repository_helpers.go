package postgres

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrRepositoryUUIDArgumentInvalid = errors.New(
		"repository UUID argument is invalid",
	)
	ErrRepositorySourceNameRequired = errors.New(
		"repository source name is required",
	)
)

type nullableUUIDArgument struct {
	value string
}

func nullableUUID(value string) nullableUUIDArgument {
	return nullableUUIDArgument{value: value}
}

func (argument nullableUUIDArgument) Value() (driver.Value, error) {
	trimmed := strings.TrimSpace(argument.value)
	if trimmed == "" {
		return nil, nil
	}

	var identifier pgtype.UUID
	if err := identifier.Scan(trimmed); err != nil || !identifier.Valid {
		return nil, fmt.Errorf(
			"%w: %q",
			ErrRepositoryUUIDArgumentInvalid,
			trimmed,
		)
	}

	value, err := identifier.Value()
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %q: %v",
			ErrRepositoryUUIDArgumentInvalid,
			trimmed,
			err,
		)
	}
	return value, nil
}

type nullableTextArgument struct {
	value string
}

func nullableText(value string) nullableTextArgument {
	return nullableTextArgument{value: value}
}

func (argument nullableTextArgument) Value() (driver.Value, error) {
	trimmed := strings.TrimSpace(argument.value)
	if trimmed == "" {
		return nil, nil
	}
	return trimmed, nil
}

type requiredSourceNameArgument struct {
	value string
}

func requiredSourceNameValue(value string) requiredSourceNameArgument {
	return requiredSourceNameArgument{value: value}
}

func (argument requiredSourceNameArgument) Value() (driver.Value, error) {
	trimmed := strings.TrimSpace(argument.value)
	if trimmed == "" {
		return nil, ErrRepositorySourceNameRequired
	}
	return trimmed, nil
}
