package migrator

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

var ErrDuplicateMigrationVersion = errors.New(
	"duplicate migration version",
)

type DuplicateMigrationVersionError struct {
	Version   string
	FileNames []string
}

func (err *DuplicateMigrationVersionError) Error() string {
	if err == nil {
		return ErrDuplicateMigrationVersion.Error()
	}

	return fmt.Sprintf(
		"%s %s: %s",
		ErrDuplicateMigrationVersion,
		err.Version,
		strings.Join(err.FileNames, ", "),
	)
}

func (err *DuplicateMigrationVersionError) Unwrap() error {
	return ErrDuplicateMigrationVersion
}

func validateUniqueMigrationVersions(
	migrations []Migration,
) error {
	filesByVersion := make(
		map[string][]string,
		len(migrations),
	)
	for _, migration := range migrations {
		filesByVersion[migration.Version] = append(
			filesByVersion[migration.Version],
			filepath.Base(migration.Path),
		)
	}

	versions := make(
		[]string,
		0,
		len(filesByVersion),
	)
	for version, fileNames := range filesByVersion {
		if len(fileNames) > 1 {
			versions = append(versions, version)
		}
	}
	if len(versions) == 0 {
		return nil
	}

	sort.Strings(versions)
	version := versions[0]
	fileNames := append(
		[]string(nil),
		filesByVersion[version]...,
	)
	sort.Strings(fileNames)

	return &DuplicateMigrationVersionError{
		Version:   version,
		FileNames: fileNames,
	}
}
