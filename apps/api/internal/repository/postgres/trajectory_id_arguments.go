package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

var ErrTrajectoryIdentifierInvalid = errors.New(
	"trajectory identifier is invalid",
)

func trajectoryUUIDArguments(
	values []string,
) ([]pgtype.UUID, error) {
	result := make([]pgtype.UUID, 0, len(values))
	for index, value := range values {
		trimmed := strings.TrimSpace(value)
		var identifier pgtype.UUID
		if err := identifier.Scan(trimmed); err != nil || !identifier.Valid {
			return nil, fmt.Errorf(
				"%w: index=%d value=%q",
				ErrTrajectoryIdentifierInvalid,
				index,
				trimmed,
			)
		}
		result = append(result, identifier)
	}
	return result, nil
}
