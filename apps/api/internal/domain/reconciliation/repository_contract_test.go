package reconciliation

import "testing"

func TestRepositoryRemainsCompositionOfNarrowPorts(t *testing.T) {
	var _ DerivationWriter = (Repository)(nil)
	var _ TaskClaimer = (Repository)(nil)
	var _ TaskTransitionWriter = (Repository)(nil)
	var _ StaleTaskRequeuer = (Repository)(nil)
}
