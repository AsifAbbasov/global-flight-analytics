package providerbatch

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEvidenceInvalid = errors.New(
		"provider batch evidence is invalid",
	)
	ErrAllItemsRejected = errors.New(
		"provider batch contains no acceptable items",
	)
)

type Evidence struct {
	Received          int
	Accepted          int
	RejectedMalformed int
	RejectedUnusable  int
}

func AcceptedOnly(count int) Evidence {
	if count < 0 {
		count = 0
	}
	return Evidence{
		Received: count,
		Accepted: count,
	}
}

func Resolve(
	evidence Evidence,
	acceptedStateCount int,
) (Evidence, error) {
	if evidence == (Evidence{}) {
		evidence = AcceptedOnly(acceptedStateCount)
	}
	if evidence.Accepted != acceptedStateCount {
		return Evidence{}, fmt.Errorf(
			"%w: accepted=%d states=%d",
			ErrEvidenceInvalid,
			evidence.Accepted,
			acceptedStateCount,
		)
	}
	if err := evidence.Validate(); err != nil {
		return Evidence{}, err
	}
	return evidence, nil
}

func (evidence Evidence) RejectedCount() int {
	return evidence.RejectedMalformed +
		evidence.RejectedUnusable
}

func (evidence Evidence) Partial() bool {
	return evidence.RejectedCount() > 0 &&
		evidence.Accepted > 0
}

func (evidence Evidence) Validate() error {
	if evidence.Received < 0 ||
		evidence.Accepted < 0 ||
		evidence.RejectedMalformed < 0 ||
		evidence.RejectedUnusable < 0 {
		return ErrEvidenceInvalid
	}
	if evidence.Accepted+evidence.RejectedCount() !=
		evidence.Received {
		return fmt.Errorf(
			"%w: received=%d accepted=%d rejected=%d",
			ErrEvidenceInvalid,
			evidence.Received,
			evidence.Accepted,
			evidence.RejectedCount(),
		)
	}
	return nil
}

func (evidence Evidence) PartialMessage() string {
	return fmt.Sprintf(
		"provider batch partially accepted: received=%d accepted=%d rejected=%d malformed=%d unusable=%d",
		evidence.Received,
		evidence.Accepted,
		evidence.RejectedCount(),
		evidence.RejectedMalformed,
		evidence.RejectedUnusable,
	)
}

type AllItemsRejectedError struct {
	Provider string
	Evidence Evidence
}

func NewAllItemsRejectedError(
	provider string,
	evidence Evidence,
) error {
	return &AllItemsRejectedError{
		Provider: strings.TrimSpace(provider),
		Evidence: evidence,
	}
}

func (err *AllItemsRejectedError) Error() string {
	if err == nil {
		return ErrAllItemsRejected.Error()
	}
	return fmt.Sprintf(
		"%s: provider=%s received=%d malformed=%d unusable=%d",
		ErrAllItemsRejected,
		err.Provider,
		err.Evidence.Received,
		err.Evidence.RejectedMalformed,
		err.Evidence.RejectedUnusable,
	)
}

func (err *AllItemsRejectedError) Unwrap() error {
	return ErrAllItemsRejected
}

func (err *AllItemsRejectedError) ProviderBatchEvidence() Evidence {
	if err == nil {
		return Evidence{}
	}
	return err.Evidence
}

type EvidenceCarrier interface {
	ProviderBatchEvidence() Evidence
}

func FromError(err error) (Evidence, bool) {
	var carrier EvidenceCarrier
	if !errors.As(err, &carrier) {
		return Evidence{}, false
	}
	return carrier.ProviderBatchEvidence(), true
}
