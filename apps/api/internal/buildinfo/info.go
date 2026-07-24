package buildinfo

import (
	"strings"
	"time"
)

const (
	DefaultVersion  = "1.0.0"
	UnknownRevision = "unknown"
	UnknownBuiltAt  = "unknown"
)

var (
	version  = DefaultVersion
	revision = UnknownRevision
	builtAt  = UnknownBuiltAt
)

type Info struct {
	Version  string
	Revision string
	BuiltAt  string
}

func Current() Info {
	return normalize(
		Info{
			Version:  version,
			Revision: revision,
			BuiltAt:  builtAt,
		},
	)
}

func normalize(
	info Info,
) Info {
	info.Version = strings.TrimSpace(
		info.Version,
	)
	if info.Version == "" {
		info.Version = DefaultVersion
	}

	info.Revision = strings.TrimSpace(
		info.Revision,
	)
	if info.Revision == "" {
		info.Revision = UnknownRevision
	}

	info.BuiltAt = normalizeBuiltAt(
		info.BuiltAt,
	)

	return info
}

func normalizeBuiltAt(
	value string,
) string {
	normalized := strings.TrimSpace(
		value,
	)
	if normalized == "" ||
		normalized == UnknownBuiltAt {
		return UnknownBuiltAt
	}

	parsed, err := time.Parse(
		time.RFC3339,
		normalized,
	)
	if err != nil {
		return UnknownBuiltAt
	}

	return parsed.UTC().Format(
		time.RFC3339,
	)
}
