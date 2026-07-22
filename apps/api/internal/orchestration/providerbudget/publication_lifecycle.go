package providerbudget

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var (
	ErrPublicationContextRequired = errors.New(
		"publication lifecycle context is required",
	)
	ErrPublicationReservationRequired = errors.New(
		"publication reservation is required",
	)
	ErrPublicationReservationMismatch = errors.New(
		"publication reservation does not own the active publication",
	)
	ErrPublicationAlreadyCommitted = errors.New(
		"publication is already committed",
	)
)

type PublicationReservation struct {
	Provider      providerpolicy.Provider
	PublicationID string
	Token         string
	Decision      Decision
}

type publicationKey struct {
	Provider      providerpolicy.Provider
	PublicationID string
}

type publicationStatus string

const (
	publicationStatusReserved  publicationStatus = "reserved"
	publicationStatusCommitted publicationStatus = "committed"
)

type publicationState struct {
	Status publicationStatus
	Token  string
}

func (manager *Manager) ReservePublication(
	ctx context.Context,
	provider providerpolicy.Provider,
	publicationID string,
) (PublicationReservation, error) {
	if ctx == nil {
		return PublicationReservation{}, ErrPublicationContextRequired
	}
	if err := ctx.Err(); err != nil {
		return PublicationReservation{}, err
	}

	normalizedPublicationID := strings.TrimSpace(publicationID)
	if normalizedPublicationID == "" {
		return PublicationReservation{}, ErrPublicationIDRequired
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	policy, err := manager.policy(provider)
	if err != nil {
		return PublicationReservation{}, err
	}
	if policy.BudgetMode != providerpolicy.BudgetModePublicationDriven {
		return PublicationReservation{}, fmt.Errorf(
			"provider %s is not publication-driven",
			provider,
		)
	}

	key := publicationKey{
		Provider:      provider,
		PublicationID: normalizedPublicationID,
	}
	state, exists := manager.publicationStates[key]
	if exists {
		switch state.Status {
		case publicationStatusCommitted:
			return PublicationReservation{
				Provider:      provider,
				PublicationID: normalizedPublicationID,
				Decision: Decision{
					Provider: provider,
					Allowed:  false,
					Reason:   DecisionReasonPublicationAlreadyProcessed,
				},
			}, nil

		case publicationStatusReserved:
			return PublicationReservation{
				Provider:      provider,
				PublicationID: normalizedPublicationID,
				Decision: Decision{
					Provider: provider,
					Allowed:  false,
					Reason:   DecisionReasonPublicationInProgress,
				},
			}, nil
		}
	}

	manager.nextPublicationToken++
	token := fmt.Sprintf(
		"%s:%d",
		provider,
		manager.nextPublicationToken,
	)
	manager.publicationStates[key] = publicationState{
		Status: publicationStatusReserved,
		Token:  token,
	}

	return PublicationReservation{
		Provider:      provider,
		PublicationID: normalizedPublicationID,
		Token:         token,
		Decision: Decision{
			Provider: provider,
			Allowed:  true,
			Reason:   DecisionReasonAllowed,
		},
	}, nil
}

func (manager *Manager) CommitPublication(
	ctx context.Context,
	reservation PublicationReservation,
) error {
	if ctx == nil {
		return ErrPublicationContextRequired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validatePublicationReservation(reservation); err != nil {
		return err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	key := publicationKey{
		Provider:      reservation.Provider,
		PublicationID: reservation.PublicationID,
	}
	state, exists := manager.publicationStates[key]
	if !exists {
		return ErrPublicationReservationRequired
	}
	if state.Token != reservation.Token {
		return ErrPublicationReservationMismatch
	}
	if state.Status == publicationStatusCommitted {
		return nil
	}
	if state.Status != publicationStatusReserved {
		return ErrPublicationReservationRequired
	}

	state.Status = publicationStatusCommitted
	manager.publicationStates[key] = state
	return nil
}

func (manager *Manager) ReleasePublication(
	ctx context.Context,
	reservation PublicationReservation,
) error {
	if ctx == nil {
		return ErrPublicationContextRequired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validatePublicationReservation(reservation); err != nil {
		return err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	key := publicationKey{
		Provider:      reservation.Provider,
		PublicationID: reservation.PublicationID,
	}
	state, exists := manager.publicationStates[key]
	if !exists {
		return nil
	}
	if state.Token != reservation.Token {
		return ErrPublicationReservationMismatch
	}
	if state.Status == publicationStatusCommitted {
		return ErrPublicationAlreadyCommitted
	}

	delete(manager.publicationStates, key)
	return nil
}

func (manager *Manager) AcquirePublication(
	provider providerpolicy.Provider,
	publicationID string,
) (Decision, error) {
	reservation, err := manager.ReservePublication(
		context.Background(),
		provider,
		publicationID,
	)
	if err != nil || !reservation.Decision.Allowed {
		return reservation.Decision, err
	}
	if err := manager.CommitPublication(
		context.Background(),
		reservation,
	); err != nil {
		return Decision{}, fmt.Errorf(
			"commit publication acquisition: %w",
			err,
		)
	}
	return reservation.Decision, nil
}

func validatePublicationReservation(
	reservation PublicationReservation,
) error {
	if reservation.Provider == "" ||
		strings.TrimSpace(reservation.PublicationID) == "" ||
		strings.TrimSpace(reservation.Token) == "" {
		return ErrPublicationReservationRequired
	}
	return nil
}
