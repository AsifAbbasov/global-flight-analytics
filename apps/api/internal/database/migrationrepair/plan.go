package migrationrepair

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/migrationfile"
)

type Plan struct {
	Anchor         migrationfile.Identity
	AnchorChecksum string
}

func LoadPlan(
	migrationsDir string,
	anchorFileName string,
) (Plan, error) {
	directory := strings.TrimSpace(migrationsDir)
	if directory == "" {
		return Plan{}, ErrMigrationsDirectoryRequired
	}

	fileName := strings.TrimSpace(anchorFileName)
	if fileName == "" {
		fileName = DefaultRepairAnchorFileName
	}
	identity, err := migrationfile.Parse(fileName)
	if err != nil {
		return Plan{}, fmt.Errorf("%w: parse anchor file name: %v", ErrRepairPlanInvalid, err)
	}

	content, err := os.ReadFile(filepath.Join(directory, identity.FileName))
	if err != nil {
		return Plan{}, fmt.Errorf(
			"%w: read anchor migration %s: %v",
			ErrRepairPlanInvalid,
			identity.FileName,
			err,
		)
	}
	digest := sha256.Sum256(content)
	plan := Plan{
		Anchor:         identity,
		AnchorChecksum: hex.EncodeToString(digest[:]),
	}
	if err := plan.Validate(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (plan Plan) Validate() error {
	if plan.Anchor.Version == "" ||
		plan.Anchor.Name == "" ||
		plan.Anchor.FileName == "" ||
		strings.TrimSpace(plan.AnchorChecksum) == "" {
		return ErrRepairPlanInvalid
	}
	return nil
}

func (plan Plan) IsLaterVersion(version string) bool {
	return strings.TrimSpace(version) > plan.Anchor.Version
}
