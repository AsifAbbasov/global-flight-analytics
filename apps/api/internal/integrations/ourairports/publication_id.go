package ourairports

import (
	"crypto/sha256"
	"fmt"
)

func publicationIDForResponseBody(responseBody []byte) string {
	digest := sha256.Sum256(responseBody)
	return fmt.Sprintf("sha256:%x", digest)
}
