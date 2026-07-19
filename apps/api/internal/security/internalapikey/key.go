// Package internalapikey defines the server-side internal mutation key
// contract. The server stores only a SHA-256 digest and never needs the raw
// administrative key in configuration.
package internalapikey

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	HeaderName = "X-Internal-API-Key"

	MinimumCandidateLength = 32
	MaximumCandidateLength = 256

	DigestHexLength = sha256.Size * 2
)

var (
	ErrDigestRequired = errors.New(
		"internal mutation key digest is required",
	)
	ErrDigestLength = errors.New(
		"internal mutation key digest must contain exactly 64 hexadecimal characters",
	)
	ErrDigestFormat = errors.New(
		"internal mutation key digest must contain only hexadecimal characters without surrounding whitespace",
	)
	ErrDigestZero = errors.New(
		"internal mutation key digest must not be the all-zero digest",
	)
)

type Digest [sha256.Size]byte

func ParseDigestHex(
	value string,
) (Digest, error) {
	if value == "" {
		return Digest{}, ErrDigestRequired
	}
	if strings.TrimSpace(value) != value {
		return Digest{}, ErrDigestFormat
	}
	if len(value) != DigestHexLength {
		return Digest{}, ErrDigestLength
	}

	decoded, err := hex.DecodeString(value)
	if err != nil {
		return Digest{},
			fmt.Errorf(
				"%w: %v",
				ErrDigestFormat,
				err,
			)
	}

	var digest Digest
	copy(digest[:], decoded)

	if digest.IsZero() {
		return Digest{}, ErrDigestZero
	}

	return digest, nil
}

func DigestCandidate(
	value string,
) Digest {
	return sha256.Sum256(
		[]byte(value),
	)
}

func (
	digest Digest,
) MatchesCandidate(
	candidate string,
) bool {
	candidateDigest := DigestCandidate(
		candidate,
	)
	compareResult := subtle.ConstantTimeCompare(
		digest[:],
		candidateDigest[:],
	)

	lengthValid :=
		len(candidate) >= MinimumCandidateLength &&
			len(candidate) <= MaximumCandidateLength

	return lengthValid && compareResult == 1
}

func (
	digest Digest,
) Hex() string {
	return hex.EncodeToString(
		digest[:],
	)
}

func (
	digest Digest,
) IsZero() bool {
	var zero Digest
	return subtle.ConstantTimeCompare(
		digest[:],
		zero[:],
	) == 1
}
