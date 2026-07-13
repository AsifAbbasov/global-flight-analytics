package registry

import "testing"

type dummy struct{}

func TestRegistry(t *testing.T) {
	r := New()
	if err := r.Register("a", dummy{}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get("a"); err != nil {
		t.Fatal(err)
	}
}
