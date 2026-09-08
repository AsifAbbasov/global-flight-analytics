package server

import "testing"

func TestFiberTransportLimitsAreExplicitlyBounded(
	t *testing.T,
) {
	normalized, err := normalizeConfig(
		Config{},
	)
	if err != nil {
		t.Fatalf(
			"normalize server config: %v",
			err,
		)
	}

	fiberConfig := newFiberConfig(
		normalized,
	)

	if fiberConfig.ReadBufferSize != defaultReadBufferSize {
		t.Fatalf(
			"expected read buffer/header limit %d, got %d",
			defaultReadBufferSize,
			fiberConfig.ReadBufferSize,
		)
	}
	if fiberConfig.Concurrency != defaultConnectionConcurrency {
		t.Fatalf(
			"expected connection concurrency %d, got %d",
			defaultConnectionConcurrency,
			fiberConfig.Concurrency,
		)
	}
}
