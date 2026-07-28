package trajectoryeligibility

import "testing"

func TestProjectionPolicyIdentityIsDeterministicAndVersioned(t *testing.T) {
	evaluator := NewDefault()
	first := evaluator.ProjectionPolicyIdentity()
	second := evaluator.ProjectionPolicyIdentity()

	if first != second {
		t.Fatalf("identity is not deterministic: %#v %#v", first, second)
	}
	if first.Name != ProjectionPolicyName ||
		first.Version != ProjectionPolicyVersion ||
		len(first.Fingerprint) != len("sha256:")+64 {
		t.Fatalf("unexpected identity: %#v", first)
	}
}
