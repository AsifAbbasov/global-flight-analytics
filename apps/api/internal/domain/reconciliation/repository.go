package reconciliation

import "context"

type Repository interface {
	MarkPendingDerivation(
		ctx context.Context,
		task PendingDerivation,
	) error
}
