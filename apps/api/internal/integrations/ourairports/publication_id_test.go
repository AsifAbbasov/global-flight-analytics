package ourairports

import (
	"strings"
	"testing"
)

func TestPublicationIDForResponseBodyIsStable(t *testing.T) {
	body := []byte("ident,name\nUBBB,Airport A\n")
	first := publicationIDForResponseBody(body)
	second := publicationIDForResponseBody(append([]byte(nil), body...))

	if first == "" || !strings.HasPrefix(first, "sha256:") {
		t.Fatalf("unexpected publication identifier: %q", first)
	}
	if first != second {
		t.Fatalf(
			"same content produced different publication identifiers: %q and %q",
			first,
			second,
		)
	}
}

func TestPublicationIDForResponseBodyChangesWithContent(t *testing.T) {
	first := publicationIDForResponseBody([]byte("publication-a"))
	second := publicationIDForResponseBody([]byte("publication-b"))
	if first == second {
		t.Fatalf("changed content reused publication identifier %q", first)
	}
}
