package researchdataset

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func ValidateManifest(
	manifest Manifest,
) (Decision, error) {
	profile, err := ProfileByID(
		manifest.DatasetID,
	)
	if err != nil {
		return Decision{}, err
	}
	if profile.Selection != SelectionAdopted {
		return Decision{}, fmt.Errorf(
			"%w: dataset=%s selection=%s",
			ErrDatasetNotAdopted,
			manifest.DatasetID,
			profile.Selection,
		)
	}
	if !manifest.OfflineOnly ||
		manifest.ProductionDependency ||
		profile.ProductionDependencyAllowed {
		return Decision{}, fmt.Errorf(
			"%w: dataset must remain offline and non-production",
			ErrManifestInvalid,
		)
	}
	if !manifest.LicenseReviewed ||
		!manifest.AttributionProvided {
		return Decision{}, fmt.Errorf(
			"%w: licence review and attribution are required",
			ErrManifestInvalid,
		)
	}
	if strings.TrimSpace(manifest.Version) == "" ||
		manifest.PreparedAt.IsZero() {
		return Decision{}, fmt.Errorf(
			"%w: version and prepared time are required",
			ErrManifestInvalid,
		)
	}
	if len(manifest.Files) == 0 ||
		manifest.TotalBytes <= 0 ||
		manifest.TotalBytes > profile.MaximumDownloadBytes {
		return Decision{}, fmt.Errorf(
			"%w: total bytes=%d maximum=%d",
			ErrManifestInvalid,
			manifest.TotalBytes,
			profile.MaximumDownloadBytes,
		)
	}
	if manifest.MaximumRecords <= 0 ||
		manifest.MaximumRecords > profile.MaximumRecords {
		return Decision{}, fmt.Errorf(
			"%w: maximum records=%d profile maximum=%d",
			ErrManifestInvalid,
			manifest.MaximumRecords,
			profile.MaximumRecords,
		)
	}
	if profile.RequiresRegionFilter &&
		strings.TrimSpace(manifest.RegionFilter) == "" {
		return Decision{}, fmt.Errorf(
			"%w: bounded region filter is required",
			ErrManifestInvalid,
		)
	}

	var fileBytes int64
	for _, file := range manifest.Files {
		if strings.TrimSpace(file.Name) == "" ||
			strings.TrimSpace(file.Format) == "" ||
			file.SizeBytes <= 0 ||
			!validSHA256(file.SHA256) {
			return Decision{}, fmt.Errorf(
				"%w: invalid file manifest for %q",
				ErrManifestInvalid,
				file.Name,
			)
		}
		fileBytes += file.SizeBytes
	}
	if fileBytes != manifest.TotalBytes {
		return Decision{}, fmt.Errorf(
			"%w: file bytes=%d total bytes=%d",
			ErrManifestInvalid,
			fileBytes,
			manifest.TotalBytes,
		)
	}

	for _, table := range manifest.SelectedTables {
		if contains(profile.BlockedTables, table) {
			return Decision{}, fmt.Errorf(
				"%w: %s",
				ErrBlockedTable,
				table,
			)
		}
		if len(profile.AllowedTables) > 0 &&
			!contains(profile.AllowedTables, table) {
			return Decision{}, fmt.Errorf(
				"%w: table %s is not in the allowlist",
				ErrManifestInvalid,
				table,
			)
		}
	}

	sourceDecision, err := EvaluateSourceBoundary(
		manifest.DatasetID,
	)
	if err != nil {
		return Decision{}, err
	}
	if !sourceDecision.Usable() {
		return Decision{}, fmt.Errorf(
			"%w: source boundary level=%s",
			ErrManifestInvalid,
			sourceDecision.Level,
		)
	}

	return Decision{
		DatasetID: manifest.DatasetID,
		Allowed:   true,
		Reasons: []string{
			"Dataset passed the fixed source boundary.",
			"Manifest is bounded by files, bytes, records, and optional table and region allowlists.",
			"Dataset remains an offline benchmark and cannot become a production dependency.",
		},
		Labels: append(
			append(
				[]string(nil),
				profile.RequiredLabels...,
			),
			sourceDecision.RequiredLabels...,
		),
	}, nil
}

func validSHA256(
	value string,
) bool {
	normalized := strings.TrimSpace(value)
	normalized = strings.TrimPrefix(
		normalized,
		"sha256:",
	)
	if len(normalized) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(
		normalized,
	)
	return err == nil &&
		len(decoded) == sha256.Size
}
