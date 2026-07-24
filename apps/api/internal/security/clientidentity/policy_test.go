package clientidentity

import (
	"errors"
	"testing"
)

func TestPolicyIgnoresForwardedHeaderWithoutTrustedProxy(
	t *testing.T,
) {
	policy, err := NewPolicy(
		Config{},
	)
	if err != nil {
		t.Fatalf(
			"create direct-connection policy: %v",
			err,
		)
	}

	actual := policy.Resolve(
		"198.51.100.10:443",
		"203.0.113.99",
	)
	if actual != "198.51.100.10" {
		t.Fatalf(
			"expected direct remote address, got %q",
			actual,
		)
	}
}

func TestPolicyIgnoresSpoofedHeaderFromUntrustedRemote(
	t *testing.T,
) {
	policy := mustPolicy(
		t,
		Config{
			Header: HeaderXForwardedFor,
			TrustedProxyRanges: []string{
				"192.0.2.0/24",
			},
		},
	)

	actual := policy.Resolve(
		"198.51.100.10:443",
		"203.0.113.99",
	)
	if actual != "198.51.100.10" {
		t.Fatalf(
			"expected untrusted remote address, got %q",
			actual,
		)
	}
}

func TestPolicyWalksTrustedForwardedChainFromRight(
	t *testing.T,
) {
	policy := mustPolicy(
		t,
		Config{
			Header: HeaderXForwardedFor,
			TrustedProxyRanges: []string{
				"192.0.2.0/24",
				"10.0.0.0/8",
			},
		},
	)

	actual := policy.Resolve(
		"192.0.2.10:443",
		"203.0.113.50, 10.20.30.40",
	)
	if actual != "203.0.113.50" {
		t.Fatalf(
			"expected original client address, got %q",
			actual,
		)
	}
}

func TestPolicyFailsClosedForMalformedForwardedChain(
	t *testing.T,
) {
	policy := mustPolicy(
		t,
		Config{
			TrustedProxyRanges: []string{
				"192.0.2.0/24",
			},
		},
	)

	actual := policy.Resolve(
		"192.0.2.10:443",
		"203.0.113.50, not-an-ip",
	)
	if actual != "192.0.2.10" {
		t.Fatalf(
			"expected trusted proxy remote fallback, got %q",
			actual,
		)
	}
}

func TestNewPolicyNormalizesAndDeduplicatesRanges(
	t *testing.T,
) {
	policy := mustPolicy(
		t,
		Config{
			Header: "x-real-ip",
			TrustedProxyRanges: []string{
				"192.0.2.10",
				"192.0.2.10/32",
				"2001:db8::/64",
			},
		},
	)

	if policy.Header() != HeaderXRealIP {
		t.Fatalf(
			"unexpected normalized header: %q",
			policy.Header(),
		)
	}

	ranges := policy.TrustedProxyRanges()
	if len(ranges) != 2 {
		t.Fatalf(
			"expected two normalized ranges, got %#v",
			ranges,
		)
	}
	if ranges[0] != "192.0.2.10/32" ||
		ranges[1] != "2001:db8::/64" {
		t.Fatalf(
			"unexpected normalized ranges: %#v",
			ranges,
		)
	}
}

func TestNewPolicyRejectsUnsafeConfiguration(
	t *testing.T,
) {
	tests := []struct {
		name          string
		config        Config
		expectedError error
	}{
		{
			name: "header without trusted ranges",
			config: Config{
				Header: HeaderXForwardedFor,
			},
			expectedError: ErrTrustedProxyRangesRequired,
		},
		{
			name: "unsupported header",
			config: Config{
				Header: "Forwarded",
				TrustedProxyRanges: []string{
					"192.0.2.0/24",
				},
			},
			expectedError: ErrClientIPHeaderInvalid,
		},
		{
			name: "entire ipv4 internet",
			config: Config{
				TrustedProxyRanges: []string{
					"0.0.0.0/0",
				},
			},
			expectedError: ErrTrustedProxyRangeTooBroad,
		},
		{
			name: "invalid range",
			config: Config{
				TrustedProxyRanges: []string{
					"invalid",
				},
			},
			expectedError: ErrTrustedProxyRangeInvalid,
		},
		{
			name: "too many ranges",
			config: Config{
				TrustedProxyRanges: repeatedRanges(
					maximumTrustedProxyRanges + 1,
				),
			},
			expectedError: ErrTrustedProxyRangeLimitExceeded,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(
				t *testing.T,
			) {
				_, err := NewPolicy(
					test.config,
				)
				if !errors.Is(
					err,
					test.expectedError,
				) {
					t.Fatalf(
						"expected %v, got %v",
						test.expectedError,
						err,
					)
				}
			},
		)
	}
}

func mustPolicy(
	t *testing.T,
	config Config,
) Policy {
	t.Helper()

	policy, err := NewPolicy(
		config,
	)
	if err != nil {
		t.Fatalf(
			"create client identity policy: %v",
			err,
		)
	}
	return policy
}

func repeatedRanges(
	count int,
) []string {
	ranges := make(
		[]string,
		0,
		count,
	)
	for index := 0; index < count; index++ {
		ranges = append(
			ranges,
			"192.0.2.1/32",
		)
	}
	return ranges
}
