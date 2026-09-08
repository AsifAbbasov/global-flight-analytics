package clientidentity

import (
	"net/netip"
	"testing"
)

func FuzzPolicyResolveReturnsBoundedIPAddress(
	f *testing.F,
) {
	for _, seed := range []string{
		"",
		"203.0.113.9",
		"203.0.113.9, 10.0.0.2",
		"not-an-ip",
		"2001:db8::1, 10.0.0.2",
	} {
		f.Add(seed)
	}

	policy, err := NewPolicy(
		Config{
			Header: HeaderXForwardedFor,
			TrustedProxyRanges: []string{
				"10.0.0.0/8",
			},
		},
	)
	if err != nil {
		f.Fatalf(
			"create client identity policy: %v",
			err,
		)
	}

	f.Fuzz(
		func(
			t *testing.T,
			forwardedValue string,
		) {
			if len(forwardedValue) > maximumForwardedHeaderSize*2 {
				t.Skip()
			}

			resolved := policy.Resolve(
				"10.0.0.1",
				forwardedValue,
			)
			if _, err := netip.ParseAddr(
				resolved,
			); err != nil {
				t.Fatalf(
					"resolver returned non-IP value %q for forwarded input %q: %v",
					resolved,
					forwardedValue,
					err,
				)
			}
		},
	)
}
