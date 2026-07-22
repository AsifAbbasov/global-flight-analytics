package dependency

import (
	"errors"
	"testing"
)

type testDependency struct{}

func TestMustRejectsNilAndTypedNil(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "typed nil", value: (*testDependency)(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				recoveredError, ok := recovered.(error)
				if !ok || !errors.Is(recoveredError, ErrRequired) {
					t.Fatalf("panic = %#v, want ErrRequired", recovered)
				}
			}()
			Must("repository", test.value)
		})
	}
}

func TestMustAcceptsPresentDependency(t *testing.T) {
	Must("repository", &testDependency{})
}
