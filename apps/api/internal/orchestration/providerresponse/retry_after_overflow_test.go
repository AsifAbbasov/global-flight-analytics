package providerresponse

import (
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestProviderSpecificRetryAfterRejectsDurationOverflow(t *testing.T) {
	value := strconv.FormatInt(maxRetryAfterSeconds+1, 10)

	_, err := parseRetryAfterSeconds(value)
	if !errors.Is(err, ErrInvalidRetryAfterHeader) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidRetryAfterHeader)
	}
}

func TestStandardRetryAfterRejectsDurationOverflow(t *testing.T) {
	_, err := parseStandardRetryAfter(
		strconv.FormatInt(maxRetryAfterSeconds+1, 10),
		time.Now().UTC(),
	)
	if !errors.Is(err, ErrInvalidRetryAfterHeader) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidRetryAfterHeader)
	}
}

func TestRetryAfterOverflowIsRejectedAtControllerBoundary(t *testing.T) {
	policy, err := providerpolicy.Get(providerpolicy.ProviderOpenSky)
	if err != nil {
		t.Fatalf("get OpenSky provider policy: %v", err)
	}

	headers := make(http.Header)
	headers.Set(
		policy.ProviderReportedBudget.RetryAfterSecondsHeader,
		strconv.FormatInt(maxRetryAfterSeconds+1, 10),
	)

	_, _, err = readRetryAfter(policy, headers, time.Now().UTC())
	if !errors.Is(err, ErrInvalidRetryAfterHeader) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidRetryAfterHeader)
	}
}
