package dataqualitycontract

import (
	"fmt"
	"strings"
)

func (value Provenance) Validate() error {
	if strings.TrimSpace(value.SourceName) == "" {
		return ErrSourceNameRequired
	}
	if value.SourceRecordTime.IsZero() {
		return ErrSourceRecordTimeRequired
	}
	if value.ReceivedAt.IsZero() {
		return ErrReceivedAtRequired
	}
	if value.ReceivedAt.Before(value.SourceRecordTime) {
		return fmt.Errorf(
			"%w: source_record_time=%s received_at=%s",
			ErrReceivedBeforeSourceRecord,
			value.SourceRecordTime.UTC().Format(timeFormat),
			value.ReceivedAt.UTC().Format(timeFormat),
		)
	}
	if strings.TrimSpace(value.IngestionRunID) == "" {
		return ErrIngestionRunIDRequired
	}
	if strings.TrimSpace(value.Transformation) == "" {
		return ErrTransformationRequired
	}
	if strings.TrimSpace(value.AlgorithmVersion) == "" {
		return ErrAlgorithmVersionRequired
	}
	if strings.TrimSpace(value.InputFingerprint) == "" {
		return ErrInputFingerprintRequired
	}
	return nil
}
