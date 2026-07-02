package postgres

import "strings"

func nullableUUID(value string) any {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return nil
	}

	return trimmed
}

func nullableText(value string) any {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return nil
	}

	return trimmed
}

func sourceNameOrUnknown(value string) string {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return "unknown"
	}

	return trimmed
}
