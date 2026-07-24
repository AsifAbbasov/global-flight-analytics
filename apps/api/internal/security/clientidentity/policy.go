package clientidentity

import (
	"errors"
	"net"
	"net/netip"
	"slices"
	"strings"
)

const (
	HeaderXForwardedFor  = "X-Forwarded-For"
	HeaderXRealIP        = "X-Real-IP"
	HeaderCFConnectingIP = "CF-Connecting-IP"

	maximumTrustedProxyRanges  = 64
	maximumForwardedChainItems = 32
	maximumForwardedHeaderSize = 4096
)

var (
	ErrTrustedProxyRangesRequired = errors.New(
		"trusted proxy ranges are required when a client ip header is configured",
	)
	ErrTrustedProxyRangeInvalid = errors.New(
		"trusted proxy range is invalid",
	)
	ErrTrustedProxyRangeTooBroad = errors.New(
		"trusted proxy range must not trust the entire internet",
	)
	ErrTrustedProxyRangeLimitExceeded = errors.New(
		"trusted proxy range limit exceeded",
	)
	ErrClientIPHeaderInvalid = errors.New(
		"client ip header is unsupported",
	)
)

type Config struct {
	Header             string
	TrustedProxyRanges []string
}

type Policy struct {
	header             string
	trustedProxyRanges []string
	trustedPrefixes    []netip.Prefix
}

func NewPolicy(
	config Config,
) (
	Policy,
	error,
) {
	hasTrustedProxyRanges := len(
		config.TrustedProxyRanges,
	) > 0

	header, err := normalizeHeader(
		config.Header,
		hasTrustedProxyRanges,
	)
	if err != nil {
		return Policy{}, err
	}

	if !hasTrustedProxyRanges {
		return Policy{}, nil
	}
	if len(config.TrustedProxyRanges) >
		maximumTrustedProxyRanges {
		return Policy{},
			ErrTrustedProxyRangeLimitExceeded
	}

	seen := make(
		map[string]struct{},
		len(config.TrustedProxyRanges),
	)
	normalizedRanges := make(
		[]string,
		0,
		len(config.TrustedProxyRanges),
	)
	prefixes := make(
		[]netip.Prefix,
		0,
		len(config.TrustedProxyRanges),
	)

	for _, rawRange := range config.TrustedProxyRanges {
		prefix, err := parseTrustedPrefix(
			rawRange,
		)
		if err != nil {
			return Policy{}, err
		}

		normalized := prefix.String()
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		normalizedRanges = append(
			normalizedRanges,
			normalized,
		)
		prefixes = append(
			prefixes,
			prefix,
		)
	}

	if len(prefixes) == 0 {
		return Policy{},
			ErrTrustedProxyRangesRequired
	}

	return Policy{
		header:             header,
		trustedProxyRanges: normalizedRanges,
		trustedPrefixes:    prefixes,
	}, nil
}

func (
	policy Policy,
) Configured() bool {
	return len(policy.trustedPrefixes) > 0
}

func (
	policy Policy,
) Header() string {
	return policy.header
}

func (
	policy Policy,
) TrustedProxyRanges() []string {
	return slices.Clone(
		policy.trustedProxyRanges,
	)
}

func (
	policy Policy,
) Resolve(
	remoteAddress string,
	forwardedValue string,
) string {
	remote, ok := parseAddress(
		remoteAddress,
	)
	if !ok {
		return strings.TrimSpace(
			remoteAddress,
		)
	}
	remote = remote.Unmap()

	if !policy.Configured() ||
		!policy.isTrusted(
			remote,
		) {
		return remote.String()
	}

	normalizedForwardedValue := strings.TrimSpace(
		forwardedValue,
	)
	if normalizedForwardedValue == "" ||
		len(normalizedForwardedValue) >
			maximumForwardedHeaderSize {
		return remote.String()
	}

	parts := strings.Split(
		normalizedForwardedValue,
		",",
	)
	if len(parts) == 0 ||
		len(parts) > maximumForwardedChainItems {
		return remote.String()
	}

	addresses := make(
		[]netip.Addr,
		0,
		len(parts),
	)
	for _, part := range parts {
		address, ok := parseAddress(
			part,
		)
		if !ok {
			return remote.String()
		}
		addresses = append(
			addresses,
			address.Unmap(),
		)
	}

	for index := len(addresses) - 1; index >= 0; index-- {
		if policy.isTrusted(
			addresses[index],
		) {
			continue
		}
		return addresses[index].String()
	}

	return addresses[0].String()
}

func normalizeHeader(
	value string,
	hasTrustedProxyRanges bool,
) (
	string,
	error,
) {
	normalized := strings.TrimSpace(
		value,
	)
	if normalized == "" {
		if !hasTrustedProxyRanges {
			return "", nil
		}
		return HeaderXForwardedFor, nil
	}
	if !hasTrustedProxyRanges {
		return "",
			ErrTrustedProxyRangesRequired
	}

	switch strings.ToLower(
		normalized,
	) {
	case strings.ToLower(
		HeaderXForwardedFor,
	):
		return HeaderXForwardedFor, nil

	case strings.ToLower(
		HeaderXRealIP,
	):
		return HeaderXRealIP, nil

	case strings.ToLower(
		HeaderCFConnectingIP,
	):
		return HeaderCFConnectingIP, nil

	default:
		return "",
			ErrClientIPHeaderInvalid
	}
}

func parseTrustedPrefix(
	value string,
) (
	netip.Prefix,
	error,
) {
	normalized := strings.TrimSpace(
		value,
	)
	if normalized == "" {
		return netip.Prefix{},
			ErrTrustedProxyRangeInvalid
	}

	if prefix, err := netip.ParsePrefix(
		normalized,
	); err == nil {
		prefix = prefix.Masked()
		if prefix.Bits() == 0 {
			return netip.Prefix{},
				ErrTrustedProxyRangeTooBroad
		}
		return prefix, nil
	}

	address, err := netip.ParseAddr(
		normalized,
	)
	if err != nil {
		return netip.Prefix{},
			ErrTrustedProxyRangeInvalid
	}
	address = address.Unmap()

	bits := 128
	if address.Is4() {
		bits = 32
	}

	return netip.PrefixFrom(
		address,
		bits,
	), nil
}

func parseAddress(
	value string,
) (
	netip.Addr,
	bool,
) {
	normalized := strings.TrimSpace(
		value,
	)
	if normalized == "" {
		return netip.Addr{}, false
	}

	if address, err := netip.ParseAddr(
		normalized,
	); err == nil {
		return address, true
	}

	host, _, err := net.SplitHostPort(
		normalized,
	)
	if err != nil {
		return netip.Addr{}, false
	}

	address, err := netip.ParseAddr(
		host,
	)
	if err != nil {
		return netip.Addr{}, false
	}

	return address, true
}

func (
	policy Policy,
) isTrusted(
	address netip.Addr,
) bool {
	for _, prefix := range policy.trustedPrefixes {
		if prefix.Contains(
			address,
		) {
			return true
		}
	}
	return false
}
