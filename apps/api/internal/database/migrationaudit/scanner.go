package migrationaudit

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/migrationfile"
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
		identity, parseErr :=
			migrationfile.Parse(fileName)
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
				Version:  identity.Version,
				Name:     identity.Name,
				FileName: identity.FileName,
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

func calculateChecksum(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(content)

	return hex.EncodeToString(sum[:]), nil
}
