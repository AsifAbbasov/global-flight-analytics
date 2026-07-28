package historicalsimilarity

import (
	"fmt"
	"sort"
	"strings"
)

func roleNotices(
	role string,
	values []Notice,
) []Notice {
	role = strings.TrimSpace(role)
	result := make(
		[]Notice,
		0,
		len(values),
	)
	for _, value := range values {
		result = append(
			result,
			Notice{
				Code: role + "_" +
					strings.TrimSpace(
						value.Code,
					),
				Message: fmt.Sprintf(
					"%s trajectory: %s",
					role,
					strings.TrimSpace(
						value.Message,
					),
				),
			},
		)
	}
	return result
}

func normalizeNotices(
	values []Notice,
) []Notice {
	seen := make(map[string]struct{})
	result := make(
		[]Notice,
		0,
		len(values),
	)

	for _, value := range values {
		value.Code = strings.TrimSpace(
			value.Code,
		)
		value.Message = strings.TrimSpace(
			value.Message,
		)
		if value.Code == "" ||
			value.Message == "" {
			continue
		}
		key := value.Code + "\x00" +
			value.Message
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if result[left].Code !=
				result[right].Code {
				return result[left].Code <
					result[right].Code
			}
			return result[left].Message <
				result[right].Message
		},
	)
	return result
}
