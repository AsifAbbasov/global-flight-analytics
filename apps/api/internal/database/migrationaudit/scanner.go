package migrationaudit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type localScan struct {
	migrations []LocalMigration
	invalid    []InvalidLocalFile
}

func scanLocalMigrations(
	migrationsDir string,
) (localScan, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return localScan{}, &LocalScanError{
			Path: migrationsDir,
			Err:  err,
		}
	}

	result := localScan{
		migrations: make(
			[]LocalMigration,
			0,
			len(entries),
		),
		invalid: make(
			[]InvalidLocalFile,
			0,
		),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		path := filepath.Join(
			migrationsDir,
			fileName,
		)
		version, name, parseErr :=
			parseLocalMigrationFileName(fileName)
		if parseErr != nil {
			result.invalid = append(
				result.invalid,
				InvalidLocalFile{
					FileName: fileName,
					Path:     path,
					Reason:   parseErr.Error(),
				},
			)
			continue
		}

		checksum, checksumErr :=
			calculateChecksum(path)
		if checksumErr != nil {
			return localScan{}, &LocalScanError{
				Path: path,
				Err:  checksumErr,
			}
		}

		result.migrations = append(
			result.migrations,
			LocalMigration{
				Version:  version,
				Name:     name,
				FileName: fileName,
				Path:     path,
				Checksum: checksum,
			},
		)
	}

	sort.SliceStable(
		result.migrations,
		func(left int, right int) bool {
			if result.migrations[left].Version !=
				result.migrations[right].Version {
				return result.migrations[left].Version <
					result.migrations[right].Version
			}
			if result.migrations[left].Name !=
				result.migrations[right].Name {
				return result.migrations[left].Name <
					result.migrations[right].Name
			}

			return result.migrations[left].FileName <
				result.migrations[right].FileName
		},
	)
	sort.SliceStable(
		result.invalid,
		func(left int, right int) bool {
			return result.invalid[left].FileName <
				result.invalid[right].FileName
		},
	)

	return result, nil
}

func parseLocalMigrationFileName(
	fileName string,
) (string, string, error) {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "", "", fmt.Errorf(
			"migration file name is empty",
		)
	}
	if !strings.HasSuffix(trimmed, ".sql") {
		return "", "", fmt.Errorf(
			"migration file must have .sql extension",
		)
	}

	withoutExtension := strings.TrimSuffix(
		trimmed,
		".sql",
	)
	parts := strings.SplitN(
		withoutExtension,
		"_",
		2,
	)
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"migration file must use format 001_name.sql",
		)
	}

	version := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])
	if len(version) != 3 {
		return "", "", fmt.Errorf(
			"migration version must contain exactly three digits",
		)
	}
	for _, character := range version {
		if !unicode.IsDigit(character) {
			return "", "", fmt.Errorf(
				"migration version must contain only digits",
			)
		}
	}
	if name == "" {
		return "", "", fmt.Errorf(
			"migration name is required",
		)
	}
	for _, character := range name {
		if unicode.IsLetter(character) ||
			unicode.IsDigit(character) ||
			character == '_' {
			continue
		}

		return "", "", fmt.Errorf(
			"migration name may contain only letters, digits, and underscores",
		)
	}

	return version, name, nil
}

func calculateChecksum(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(content)

	return hex.EncodeToString(sum[:]), nil
}
