package routepipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func unavailableEndpointEvidence(
	trajectoryID string,
	role routecontract.EndpointRole,
	asOfTime time.Time,
	reasonCode string,
	reasonMessage string,
) endpointevidence.Result {
	sum := sha256.Sum256(
		[]byte(
			fmt.Sprintf(
				"%s\x00%s\x00%s\x00%s\x00%s",
				Version,
				trajectoryID,
				role,
				asOfTime.UTC().Format(
					time.RFC3339Nano,
				),
				reasonCode,
			),
		),
	)

	return endpointevidence.Result{
		Version: endpointevidence.Version,
		Status: endpointevidence.
			SelectionStatusUnavailable,
		Role: role,
		InputFingerprint: "sha256:" +
			hex.EncodeToString(sum[:]),
		Limitations: []routecontract.Limitation{
			{
				Code:    reasonCode,
				Message: reasonMessage,
				Scope:   string(role),
			},
		},
	}
}
