package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProviderPublicationRepositoryPoolRequired = errors.New(
		"provider publication repository pool is required",
	)
	ErrProviderPublicationLeaseInvalid = errors.New(
		"provider publication lease must be greater than zero",
	)
	ErrProviderPublicationReservationNotFound = errors.New(
		"provider publication reservation not found",
	)
	ErrProviderPublicationReservationMismatch = errors.New(
		"provider publication reservation ownership mismatch",
	)
	ErrProviderPublicationAlreadyCommitted = errors.New(
		"provider publication is already committed",
	)
)

type ProviderPublicationRepository struct {
	pool          *pgxpool.Pool
	leaseDuration time.Duration
	now           func() time.Time
}

func NewProviderPublicationRepository(
	pool *pgxpool.Pool,
	leaseDuration time.Duration,
	now func() time.Time,
) *ProviderPublicationRepository {
	if now == nil {
		now = time.Now
	}
	return &ProviderPublicationRepository{
		pool:          pool,
		leaseDuration: leaseDuration,
		now:           now,
	}
}

func (repository *ProviderPublicationRepository) ReservePublication(
	ctx context.Context,
	provider providerpolicy.Provider,
	publicationID string,
) (providerbudget.PublicationReservation, error) {
	if err := repository.validate(ctx); err != nil {
		return providerbudget.PublicationReservation{}, err
	}

	normalizedPublicationID := strings.TrimSpace(publicationID)
	if normalizedPublicationID == "" {
		return providerbudget.PublicationReservation{},
			providerbudget.ErrPublicationIDRequired
	}
	if err := validatePublicationProvider(provider); err != nil {
		return providerbudget.PublicationReservation{}, err
	}

	transaction, err := repository.pool.BeginTx(
		ctx,
		pgx.TxOptions{},
	)
	if err != nil {
		return providerbudget.PublicationReservation{}, fmt.Errorf(
			"begin provider publication reservation: %w",
			err,
		)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	now := repository.now().UTC()
	reservationToken := uuid.NewString()
	leaseExpiresAt := now.Add(repository.leaseDuration)

	const insertQuery = `
		INSERT INTO provider_publications (
			provider_name,
			publication_id,
			status,
			reservation_token,
			reserved_at,
			lease_expires_at,
			updated_at
		)
		VALUES ($1, $2, 'reserved', $3, $4, $5, $4)
		ON CONFLICT (provider_name, publication_id) DO NOTHING;
	`
	commandTag, err := transaction.Exec(
		ctx,
		insertQuery,
		string(provider),
		normalizedPublicationID,
		reservationToken,
		now,
		leaseExpiresAt,
	)
	if err != nil {
		return providerbudget.PublicationReservation{}, fmt.Errorf(
			"insert provider publication reservation: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 1 {
		if err := transaction.Commit(ctx); err != nil {
			return providerbudget.PublicationReservation{}, fmt.Errorf(
				"commit provider publication reservation: %w",
				err,
			)
		}
		return allowedPublicationReservation(
			provider,
			normalizedPublicationID,
			reservationToken,
		), nil
	}

	const selectQuery = `
		SELECT
			status,
			lease_expires_at
		FROM provider_publications
		WHERE provider_name = $1
			AND publication_id = $2
		FOR UPDATE;
	`
	var status string
	var activeLeaseExpiresAt pgtype.Timestamptz
	if err := transaction.QueryRow(
		ctx,
		selectQuery,
		string(provider),
		normalizedPublicationID,
	).Scan(
		&status,
		&activeLeaseExpiresAt,
	); err != nil {
		return providerbudget.PublicationReservation{}, fmt.Errorf(
			"lock provider publication: %w",
			err,
		)
	}

	switch status {
	case "committed":
		if err := transaction.Commit(ctx); err != nil {
			return providerbudget.PublicationReservation{}, fmt.Errorf(
				"commit provider publication duplicate decision: %w",
				err,
			)
		}
		return deniedPublicationReservation(
			provider,
			normalizedPublicationID,
			providerbudget.DecisionReasonPublicationAlreadyProcessed,
			time.Time{},
		), nil

	case "reserved":
		if activeLeaseExpiresAt.Valid && now.Before(activeLeaseExpiresAt.Time.UTC()) {
			if err := transaction.Commit(ctx); err != nil {
				return providerbudget.PublicationReservation{}, fmt.Errorf(
					"commit provider publication in-progress decision: %w",
					err,
				)
			}
			return deniedPublicationReservation(
				provider,
				normalizedPublicationID,
				providerbudget.DecisionReasonPublicationInProgress,
				activeLeaseExpiresAt.Time.UTC(),
			), nil
		}

		const reclaimQuery = `
			UPDATE provider_publications
			SET
				reservation_token = $3,
				reserved_at = $4,
				lease_expires_at = $5,
				committed_at = NULL,
				updated_at = $4
			WHERE provider_name = $1
				AND publication_id = $2;
		`
		if _, err := transaction.Exec(
			ctx,
			reclaimQuery,
			string(provider),
			normalizedPublicationID,
			reservationToken,
			now,
			leaseExpiresAt,
		); err != nil {
			return providerbudget.PublicationReservation{}, fmt.Errorf(
				"reclaim expired provider publication reservation: %w",
				err,
			)
		}

	default:
		return providerbudget.PublicationReservation{}, fmt.Errorf(
			"provider publication has unsupported status %q",
			status,
		)
	}

	if err := transaction.Commit(ctx); err != nil {
		return providerbudget.PublicationReservation{}, fmt.Errorf(
			"commit reclaimed provider publication reservation: %w",
			err,
		)
	}
	return allowedPublicationReservation(
		provider,
		normalizedPublicationID,
		reservationToken,
	), nil
}

func (repository *ProviderPublicationRepository) CommitPublication(
	ctx context.Context,
	reservation providerbudget.PublicationReservation,
) error {
	if err := repository.validate(ctx); err != nil {
		return err
	}
	if err := validateDurablePublicationReservation(reservation); err != nil {
		return err
	}

	const query = `
		WITH updated AS (
			UPDATE provider_publications
			SET
				status = 'committed',
				committed_at = $4,
				lease_expires_at = NULL,
				updated_at = $4
			WHERE provider_name = $1
				AND publication_id = $2
				AND reservation_token = $3
				AND status = 'reserved'
			RETURNING 1
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM updated) THEN 'updated'
			WHEN EXISTS (
				SELECT 1
				FROM provider_publications
				WHERE provider_name = $1
					AND publication_id = $2
					AND reservation_token = $3
					AND status = 'committed'
			) THEN 'already_committed'
			WHEN EXISTS (
				SELECT 1
				FROM provider_publications
				WHERE provider_name = $1
					AND publication_id = $2
			) THEN 'mismatch'
			ELSE 'not_found'
		END;
	`

	var outcome string
	if err := repository.pool.QueryRow(
		ctx,
		query,
		string(reservation.Provider),
		reservation.PublicationID,
		reservation.Token,
		repository.now().UTC(),
	).Scan(&outcome); err != nil {
		return fmt.Errorf("commit provider publication: %w", err)
	}

	switch outcome {
	case "updated", "already_committed":
		return nil
	case "mismatch":
		return ErrProviderPublicationReservationMismatch
	case "not_found":
		return ErrProviderPublicationReservationNotFound
	default:
		return fmt.Errorf(
			"commit provider publication returned unknown outcome %q",
			outcome,
		)
	}
}

func (repository *ProviderPublicationRepository) ReleasePublication(
	ctx context.Context,
	reservation providerbudget.PublicationReservation,
) error {
	if err := repository.validate(ctx); err != nil {
		return err
	}
	if err := validateDurablePublicationReservation(reservation); err != nil {
		return err
	}

	const query = `
		WITH deleted AS (
			DELETE FROM provider_publications
			WHERE provider_name = $1
				AND publication_id = $2
				AND reservation_token = $3
				AND status = 'reserved'
			RETURNING 1
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM deleted) THEN 'deleted'
			WHEN EXISTS (
				SELECT 1
				FROM provider_publications
				WHERE provider_name = $1
					AND publication_id = $2
					AND reservation_token = $3
					AND status = 'committed'
			) THEN 'committed'
			WHEN EXISTS (
				SELECT 1
				FROM provider_publications
				WHERE provider_name = $1
					AND publication_id = $2
			) THEN 'mismatch'
			ELSE 'not_found'
		END;
	`

	var outcome string
	if err := repository.pool.QueryRow(
		ctx,
		query,
		string(reservation.Provider),
		reservation.PublicationID,
		reservation.Token,
	).Scan(&outcome); err != nil {
		return fmt.Errorf("release provider publication: %w", err)
	}

	switch outcome {
	case "deleted", "not_found":
		return nil
	case "committed":
		return ErrProviderPublicationAlreadyCommitted
	case "mismatch":
		return ErrProviderPublicationReservationMismatch
	default:
		return fmt.Errorf(
			"release provider publication returned unknown outcome %q",
			outcome,
		)
	}
}

func (repository *ProviderPublicationRepository) validate(
	ctx context.Context,
) error {
	if repository == nil || repository.pool == nil {
		return ErrProviderPublicationRepositoryPoolRequired
	}
	if repository.leaseDuration <= 0 {
		return ErrProviderPublicationLeaseInvalid
	}
	return requireRepositoryContext(ctx)
}

func validatePublicationProvider(
	provider providerpolicy.Provider,
) error {
	policy, err := providerpolicy.Get(provider)
	if err != nil {
		return err
	}
	if policy.BudgetMode != providerpolicy.BudgetModePublicationDriven {
		return fmt.Errorf(
			"provider %s is not publication-driven",
			provider,
		)
	}
	return nil
}

func validateDurablePublicationReservation(
	reservation providerbudget.PublicationReservation,
) error {
	if reservation.Provider == "" ||
		strings.TrimSpace(reservation.PublicationID) == "" ||
		strings.TrimSpace(reservation.Token) == "" {
		return providerbudget.ErrPublicationReservationRequired
	}
	if _, err := uuid.Parse(reservation.Token); err != nil {
		return fmt.Errorf(
			"parse provider publication reservation token: %w",
			err,
		)
	}
	return validatePublicationProvider(reservation.Provider)
}

func allowedPublicationReservation(
	provider providerpolicy.Provider,
	publicationID string,
	token string,
) providerbudget.PublicationReservation {
	return providerbudget.PublicationReservation{
		Provider:      provider,
		PublicationID: publicationID,
		Token:         token,
		Decision: providerbudget.Decision{
			Provider: provider,
			Allowed:  true,
			Reason:   providerbudget.DecisionReasonAllowed,
		},
	}
}

func deniedPublicationReservation(
	provider providerpolicy.Provider,
	publicationID string,
	reason providerbudget.DecisionReason,
	retryAt time.Time,
) providerbudget.PublicationReservation {
	return providerbudget.PublicationReservation{
		Provider:      provider,
		PublicationID: publicationID,
		Decision: providerbudget.Decision{
			Provider: provider,
			Allowed:  false,
			Reason:   reason,
			RetryAt:  retryAt.UTC(),
		},
	}
}
