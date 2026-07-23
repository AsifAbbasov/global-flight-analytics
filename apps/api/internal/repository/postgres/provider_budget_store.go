package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProviderBudgetStorePoolRequired = errors.New(
		"provider budget store pool is required",
	)
	ErrProviderBudgetStoreTimeoutInvalid = errors.New(
		"provider budget store timeout must be greater than zero",
	)
	ErrProviderBudgetFallbackRetryAfterInvalid = errors.New(
		"provider budget fallback retry-after must be greater than zero",
	)
	ErrProviderBudgetReservationInvalid = errors.New(
		"provider budget fixed-window reservation is invalid",
	)
	ErrProviderBudgetRetryAtRequired = errors.New(
		"exhausted provider budget requires retry time",
	)
)

type ProviderBudgetStore struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

var _ providerbudget.StateStore = (*ProviderBudgetStore)(nil)

func NewProviderBudgetStore(
	db *pgxpool.Pool,
	timeout time.Duration,
) (*ProviderBudgetStore, error) {
	if db == nil {
		return nil, ErrProviderBudgetStorePoolRequired
	}
	if timeout <= 0 {
		return nil, ErrProviderBudgetStoreTimeoutInvalid
	}

	return &ProviderBudgetStore{
		db:      db,
		timeout: timeout,
	}, nil
}

func (store *ProviderBudgetStore) AcquireFixedWindow(
	provider providerpolicy.Provider,
	reservations []providerbudget.FixedWindowReservation,
	now time.Time,
) (providerbudget.Decision, error) {
	if err := store.validate(); err != nil {
		return providerbudget.Decision{}, err
	}

	normalizedReservations, err := normalizeFixedWindowReservations(
		reservations,
	)
	if err != nil {
		return providerbudget.Decision{}, err
	}

	now = now.UTC()
	ctx, cancel := context.WithTimeout(
		context.Background(),
		store.timeout,
	)
	defer cancel()

	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return providerbudget.Decision{}, fmt.Errorf(
			"begin fixed-window provider budget transaction: %w",
			err,
		)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	retryAt := time.Time{}

	for _, reservation := range normalizedReservations {
		_, err = tx.Exec(
			ctx,
			`
                INSERT INTO provider_budget_fixed_windows (
                    provider_name,
                    limit_index,
                    window_start,
                    window_end,
                    request_count,
                    updated_at
                )
                VALUES ($1, $2, $3, $4, 0, $5)
                ON CONFLICT (
                    provider_name,
                    limit_index
                )
                DO NOTHING
            `,
			string(provider),
			reservation.LimitIndex,
			reservation.WindowStart,
			reservation.WindowEnd,
			now,
		)
		if err != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"initialize fixed-window provider budget: %w",
				err,
			)
		}

		var storedWindowStart time.Time
		var requestCount int
		err = tx.QueryRow(
			ctx,
			`
                SELECT
                    window_start,
                    request_count
                FROM provider_budget_fixed_windows
                WHERE provider_name = $1
                    AND limit_index = $2
                FOR UPDATE
            `,
			string(provider),
			reservation.LimitIndex,
		).Scan(
			&storedWindowStart,
			&requestCount,
		)
		if err != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"lock fixed-window provider budget: %w",
				err,
			)
		}

		if !storedWindowStart.UTC().Equal(
			reservation.WindowStart,
		) {
			commandTag, resetErr := tx.Exec(
				ctx,
				`
                    UPDATE provider_budget_fixed_windows
                    SET
                        window_start = $3,
                        window_end = $4,
                        request_count = 0,
                        updated_at = $5
                    WHERE provider_name = $1
                        AND limit_index = $2
                `,
				string(provider),
				reservation.LimitIndex,
				reservation.WindowStart,
				reservation.WindowEnd,
				now,
			)
			if resetErr != nil {
				return providerbudget.Decision{}, fmt.Errorf(
					"reset fixed-window provider budget: %w",
					resetErr,
				)
			}
			if commandTag.RowsAffected() != 1 {
				return providerbudget.Decision{}, fmt.Errorf(
					"reset fixed-window provider budget affected %d rows",
					commandTag.RowsAffected(),
				)
			}
			requestCount = 0
		} else {
			_, updateErr := tx.Exec(
				ctx,
				`
                    UPDATE provider_budget_fixed_windows
                    SET
                        window_end = $3,
                        updated_at = $4
                    WHERE provider_name = $1
                        AND limit_index = $2
                `,
				string(provider),
				reservation.LimitIndex,
				reservation.WindowEnd,
				now,
			)
			if updateErr != nil {
				return providerbudget.Decision{}, fmt.Errorf(
					"refresh fixed-window provider budget bounds: %w",
					updateErr,
				)
			}
		}

		if requestCount >= reservation.MaxRequests &&
			(retryAt.IsZero() || reservation.WindowEnd.After(retryAt)) {
			retryAt = reservation.WindowEnd
		}
	}

	if !retryAt.IsZero() {
		if err := tx.Commit(ctx); err != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"commit fixed-window provider budget denial: %w",
				err,
			)
		}

		return providerbudget.Decision{
			Provider: provider,
			Allowed:  false,
			Reason: providerbudget.
				DecisionReasonFixedWindowExhausted,
			RetryAt: retryAt,
		}, nil
	}

	for _, reservation := range normalizedReservations {
		commandTag, updateErr := tx.Exec(
			ctx,
			`
                UPDATE provider_budget_fixed_windows
                SET
                    request_count = request_count + 1,
                    updated_at = $4
                WHERE provider_name = $1
                    AND limit_index = $2
                    AND window_start = $3
            `,
			string(provider),
			reservation.LimitIndex,
			reservation.WindowStart,
			now,
		)
		if updateErr != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"consume fixed-window provider budget: %w",
				updateErr,
			)
		}
		if commandTag.RowsAffected() != 1 {
			return providerbudget.Decision{}, fmt.Errorf(
				"consume fixed-window provider budget affected %d rows",
				commandTag.RowsAffected(),
			)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return providerbudget.Decision{}, fmt.Errorf(
			"commit fixed-window provider budget acquisition: %w",
			err,
		)
	}

	return providerbudget.Decision{
		Provider: provider,
		Allowed:  true,
		Reason:   providerbudget.DecisionReasonAllowed,
	}, nil
}

func (store *ProviderBudgetStore) AcquireProviderReported(
	provider providerpolicy.Provider,
	now time.Time,
	fallbackRetryAfter time.Duration,
) (providerbudget.Decision, error) {
	if err := store.validate(); err != nil {
		return providerbudget.Decision{}, err
	}
	if fallbackRetryAfter <= 0 {
		return providerbudget.Decision{},
			ErrProviderBudgetFallbackRetryAfterInvalid
	}

	now = now.UTC()
	probeRetryAt := now.Add(fallbackRetryAfter).UTC()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		store.timeout,
	)
	defer cancel()

	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return providerbudget.Decision{}, fmt.Errorf(
			"begin provider-reported budget transaction: %w",
			err,
		)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	commandTag, err := tx.Exec(
		ctx,
		`
            INSERT INTO provider_budget_reported_states (
                provider_name,
                remaining_known,
                remaining,
                retry_at,
                observed_at,
                updated_at
            )
            VALUES ($1, false, 0, $2, $3, $3)
            ON CONFLICT (provider_name)
            DO NOTHING
        `,
		string(provider),
		probeRetryAt,
		now,
	)
	if err != nil {
		return providerbudget.Decision{}, fmt.Errorf(
			"initialize provider-reported probe lease: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 1 {
		if err := tx.Commit(ctx); err != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"commit initial provider-reported probe lease: %w",
				err,
			)
		}

		return providerbudget.Decision{
			Provider: provider,
			Allowed:  true,
			Reason:   providerbudget.DecisionReasonAllowed,
		}, nil
	}

	var remainingKnown bool
	var remaining int
	var retryAt sql.NullTime
	err = tx.QueryRow(
		ctx,
		`
            SELECT
                remaining_known,
                remaining,
                retry_at
            FROM provider_budget_reported_states
            WHERE provider_name = $1
            FOR UPDATE
        `,
		string(provider),
	).Scan(
		&remainingKnown,
		&remaining,
		&retryAt,
	)
	if err != nil {
		return providerbudget.Decision{}, fmt.Errorf(
			"lock provider-reported budget: %w",
			err,
		)
	}

	if retryAt.Valid && now.Before(retryAt.Time.UTC()) {
		if err := tx.Commit(ctx); err != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"commit provider-reported cooldown denial: %w",
				err,
			)
		}

		return providerbudget.Decision{
			Provider: provider,
			Allowed:  false,
			Reason:   providerbudget.DecisionReasonProviderCooldown,
			RetryAt:  retryAt.Time.UTC(),
		}, nil
	}

	if !remainingKnown {
		commandTag, updateErr := tx.Exec(
			ctx,
			`
                UPDATE provider_budget_reported_states
                SET
                    remaining = 0,
                    retry_at = $2,
                    updated_at = $3
                WHERE provider_name = $1
            `,
			string(provider),
			probeRetryAt,
			now,
		)
		if updateErr != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"renew provider-reported probe lease: %w",
				updateErr,
			)
		}
		if commandTag.RowsAffected() != 1 {
			return providerbudget.Decision{}, fmt.Errorf(
				"renew provider-reported probe lease affected %d rows",
				commandTag.RowsAffected(),
			)
		}
		if err := tx.Commit(ctx); err != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"commit provider-reported probe allowance: %w",
				err,
			)
		}

		return providerbudget.Decision{
			Provider: provider,
			Allowed:  true,
			Reason:   providerbudget.DecisionReasonAllowed,
		}, nil
	}

	if remaining <= 0 {
		commandTag, updateErr := tx.Exec(
			ctx,
			`
                UPDATE provider_budget_reported_states
                SET
                    retry_at = $2,
                    updated_at = $3
                WHERE provider_name = $1
            `,
			string(provider),
			probeRetryAt,
			now,
		)
		if updateErr != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"schedule exhausted provider budget retry: %w",
				updateErr,
			)
		}
		if commandTag.RowsAffected() != 1 {
			return providerbudget.Decision{}, fmt.Errorf(
				"schedule exhausted provider budget retry affected %d rows",
				commandTag.RowsAffected(),
			)
		}
		if err := tx.Commit(ctx); err != nil {
			return providerbudget.Decision{}, fmt.Errorf(
				"commit exhausted provider budget retry: %w",
				err,
			)
		}

		return providerbudget.Decision{
			Provider: provider,
			Allowed:  false,
			Reason: providerbudget.
				DecisionReasonProviderBudgetExhausted,
			RetryAt: probeRetryAt,
		}, nil
	}

	commandTag, err = tx.Exec(
		ctx,
		`
            UPDATE provider_budget_reported_states
            SET
                remaining = remaining - 1,
                retry_at = NULL,
                updated_at = $2
            WHERE provider_name = $1
        `,
		string(provider),
		now,
	)
	if err != nil {
		return providerbudget.Decision{}, fmt.Errorf(
			"consume provider-reported budget: %w",
			err,
		)
	}
	if commandTag.RowsAffected() != 1 {
		return providerbudget.Decision{}, fmt.Errorf(
			"consume provider-reported budget affected %d rows",
			commandTag.RowsAffected(),
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return providerbudget.Decision{}, fmt.Errorf(
			"commit provider-reported budget acquisition: %w",
			err,
		)
	}

	return providerbudget.Decision{
		Provider: provider,
		Allowed:  true,
		Reason:   providerbudget.DecisionReasonAllowed,
	}, nil
}

func (store *ProviderBudgetStore) ObserveProviderReportedBudget(
	provider providerpolicy.Provider,
	remaining int,
	retryAt time.Time,
	observedAt time.Time,
) error {
	if err := store.validate(); err != nil {
		return err
	}
	if remaining < 0 {
		return providerbudget.ErrInvalidRemainingBudget
	}
	if remaining == 0 && retryAt.IsZero() {
		return ErrProviderBudgetRetryAtRequired
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		store.timeout,
	)
	defer cancel()

	var retryAtValue any
	if !retryAt.IsZero() {
		retryAtValue = retryAt.UTC()
	}

	_, err := store.db.Exec(
		ctx,
		`
            INSERT INTO provider_budget_reported_states (
                provider_name,
                remaining_known,
                remaining,
                retry_at,
                observed_at,
                updated_at
            )
            VALUES ($1, true, $2, $3, $4, $4)
            ON CONFLICT (provider_name)
            DO UPDATE SET
                remaining_known = EXCLUDED.remaining_known,
                remaining = EXCLUDED.remaining,
                retry_at = EXCLUDED.retry_at,
                observed_at = EXCLUDED.observed_at,
                updated_at = EXCLUDED.updated_at
            WHERE EXCLUDED.observed_at >=
                provider_budget_reported_states.observed_at
        `,
		string(provider),
		remaining,
		retryAtValue,
		observedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf(
			"persist provider-reported budget observation: %w",
			err,
		)
	}

	return nil
}

func normalizeFixedWindowReservations(
	reservations []providerbudget.FixedWindowReservation,
) ([]providerbudget.FixedWindowReservation, error) {
	if len(reservations) == 0 {
		return nil, ErrProviderBudgetReservationInvalid
	}

	normalized := append(
		[]providerbudget.FixedWindowReservation(nil),
		reservations...,
	)
	sort.Slice(
		normalized,
		func(left int, right int) bool {
			return normalized[left].LimitIndex <
				normalized[right].LimitIndex
		},
	)

	previousLimitIndex := -1
	for index := range normalized {
		reservation := &normalized[index]
		reservation.WindowStart = reservation.WindowStart.UTC()
		reservation.WindowEnd = reservation.WindowEnd.UTC()

		if reservation.LimitIndex < 0 ||
			reservation.LimitIndex == previousLimitIndex ||
			reservation.MaxRequests <= 0 ||
			reservation.WindowStart.IsZero() ||
			!reservation.WindowEnd.After(reservation.WindowStart) {
			return nil, ErrProviderBudgetReservationInvalid
		}

		previousLimitIndex = reservation.LimitIndex
	}

	return normalized, nil
}

func (store *ProviderBudgetStore) validate() error {
	if store == nil || store.db == nil {
		return ErrProviderBudgetStorePoolRequired
	}
	if store.timeout <= 0 {
		return ErrProviderBudgetStoreTimeoutInvalid
	}

	return nil
}
