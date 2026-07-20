package migrationfile

import (
	"fmt"
	"strings"
	"unicode"
)

// Identity is the canonical identity encoded by a migration file name.
type Identity struct {
	Version  string
	Name     string
	FileName string
}

// Parse validates and decodes a canonical migration file name.
//
// Canonical format:
//
//	NNN_name.sql
//
// Version is exactly three ASCII digits. Name is non-empty and may contain
// Unicode letters, Unicode digits, and underscores. Leading or trailing white
// space, path separators, and non-lowercase .sql extensions are rejected.
func Parse(fileName string) (Identity, error) {
	if fileName == "" {
		return Identity{}, fmt.Errorf("migration file name is empty")
	}
	if strings.TrimSpace(fileName) != fileName {
		return Identity{}, fmt.Errorf(
			"migration file name must not contain leading or trailing whitespace",
		)
	}
	if strings.ContainsAny(fileName, `/\\`) {
		return Identity{}, fmt.Errorf(
			"migration file name must not contain path separators",
		)
	}
	if !strings.HasSuffix(fileName, ".sql") {
		return Identity{}, fmt.Errorf(
			"migration file must have .sql extension",
		)
	}

	withoutExtension := strings.TrimSuffix(fileName, ".sql")
	parts := strings.SplitN(withoutExtension, "_", 2)
	if len(parts) != 2 {
		return Identity{}, fmt.Errorf(
			"migration file must use format 001_name.sql",
		)
	}

	version := parts[0]
	name := parts[1]
	if len(version) != 3 {
		return Identity{}, fmt.Errorf(
			"migration version must contain exactly three ASCII digits",
		)
	}
	for index := 0; index < len(version); index++ {
		if version[index] < '0' || version[index] > '9' {
			return Identity{}, fmt.Errorf(
				"migration version must contain only ASCII digits",
			)
		}
	}
	if name == "" {
		return Identity{}, fmt.Errorf("migration name is required")
	}
	for _, character := range name {
		if unicode.IsLetter(character) ||
			unicode.IsDigit(character) ||
			character == '_' {
			continue
		}

		return Identity{}, fmt.Errorf(
			"migration name may contain only letters, digits, and underscores",
		)
	}

	return Identity{
		Version:  version,
		Name:     name,
		FileName: fileName,
	}, nil
}

// MustParse validates a source-owned canonical migration file name and panics
// when the source constant is invalid. It is intended only for package-level
// migration identities controlled by the repository.
func MustParse(fileName string) Identity {
	identity, err := Parse(fileName)
	if err != nil {
		panic(fmt.Sprintf(
			"invalid source-owned migration file name %q: %v",
			fileName,
			err,
		))
	}

	return identity
}
