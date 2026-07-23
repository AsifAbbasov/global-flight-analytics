package ingestion

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
)

func providerBatchPartialFailure(
	evidence providerbatch.Evidence,
) (bool, string) {
	if !evidence.Partial() {
		return false, ""
	}
	return true, evidence.PartialMessage()
}
