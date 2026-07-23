package main

import (
	"fmt"
	"time"

	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

const (
	trafficProviderRecoveryProbeAfter = 2 * time.Minute

	trafficHealthReasonConfiguredOrder     = "configured_provider_order_preserved"
	trafficHealthReasonSnapshotUnavailable = "configured_provider_order_preserved_health_snapshot_unavailable"
	trafficHealthReasonStrongerProvider    = "provider_with_stronger_health_evidence_preferred"
	trafficHealthReasonRecoveryProbe       = "configured_primary_recovery_probe_due"
)

type trafficProviderHealthSource interface {
	Snapshot(
		provider providerpolicy.Provider,
	) (providerhealthdomain.Snapshot, error)
}

type trafficProviderHealthOrder struct {
	ConfiguredPrimary providerpolicy.Provider
	FirstProvider     providerpolicy.Provider
	Statuses          map[providerpolicy.Provider]providerhealthdomain.Status
	Reordered         bool
	Reason            string
}

func orderTrafficProviderSelections(
	configured []trafficProviderSelection,
	source trafficProviderHealthSource,
) (
	[]trafficProviderSelection,
	trafficProviderHealthOrder,
) {
	ordered := append(
		[]trafficProviderSelection(nil),
		configured...,
	)
	snapshots := make(
		map[providerpolicy.Provider]providerhealthdomain.Snapshot,
		len(configured),
	)
	evidence := trafficProviderHealthOrder{
		Statuses: make(
			map[providerpolicy.Provider]providerhealthdomain.Status,
			len(configured),
		),
		Reason: trafficHealthReasonConfiguredOrder,
	}
	if len(configured) == 0 {
		return ordered, evidence
	}

	evidence.ConfiguredPrimary = configured[0].ProviderID
	snapshotUnavailable := false

	for _, selection := range configured {
		status := providerhealthdomain.StatusUnknown
		snapshot := providerhealthdomain.Snapshot{
			ProviderName: string(selection.ProviderID),
			Status:       providerhealthdomain.StatusUnknown,
		}

		if source != nil {
			observedSnapshot, err := source.Snapshot(
				selection.ProviderID,
			)
			if err != nil {
				snapshotUnavailable = true
			} else if validTrafficProviderHealthStatus(
				observedSnapshot.Status,
			) {
				snapshot = observedSnapshot
				status = observedSnapshot.Status
			} else {
				snapshotUnavailable = true
			}
		} else {
			snapshotUnavailable = true
		}

		evidence.Statuses[selection.ProviderID] = status
		snapshots[selection.ProviderID] = snapshot
	}

	recoveryProbeDue := trafficProviderRecoveryProbeDue(
		snapshots[evidence.ConfiguredPrimary],
	)

	for index := 1; index < len(ordered); index++ {
		current := ordered[index]
		currentRank := trafficProviderEffectiveHealthRank(
			current.ProviderID,
			evidence.ConfiguredPrimary,
			evidence.Statuses[current.ProviderID],
			recoveryProbeDue,
		)
		position := index

		for position > 0 {
			previous := ordered[position-1]
			previousRank := trafficProviderEffectiveHealthRank(
				previous.ProviderID,
				evidence.ConfiguredPrimary,
				evidence.Statuses[previous.ProviderID],
				recoveryProbeDue,
			)
			if currentRank >= previousRank {
				break
			}
			ordered[position] = previous
			position--
		}

		ordered[position] = current
	}

	evidence.FirstProvider = ordered[0].ProviderID
	evidence.Reordered =
		evidence.FirstProvider != evidence.ConfiguredPrimary

	switch {
	case recoveryProbeDue && !evidence.Reordered:
		evidence.Reason = trafficHealthReasonRecoveryProbe
	case evidence.Reordered:
		evidence.Reason = trafficHealthReasonStrongerProvider
	case snapshotUnavailable:
		evidence.Reason = trafficHealthReasonSnapshotUnavailable
	}

	return ordered, evidence
}

func decorateTrafficFallbackDecision(
	decision providerfallback.Decision,
	healthOrder trafficProviderHealthOrder,
) providerfallback.Decision {
	if healthOrder.ConfiguredPrimary == "" {
		return decision
	}

	decision.HealthAware = true
	decision.HealthReordered = healthOrder.Reordered
	decision.HealthOrderingReason = healthOrder.Reason
	decision.PrimaryHealthStatus =
		healthOrder.Statuses[healthOrder.ConfiguredPrimary]
	decision.PrimaryProvider = healthOrder.ConfiguredPrimary

	if decision.SelectedProvider != "" {
		decision.SelectedHealthStatus =
			healthOrder.Statuses[decision.SelectedProvider]
	}

	if healthOrder.Reordered &&
		healthOrder.FirstProvider != healthOrder.ConfiguredPrimary {
		decision.UsedFallback = true

		if decision.SelectedProvider != "" &&
			decision.SelectedProvider != healthOrder.ConfiguredPrimary &&
			decision.Outcome == providerfallback.OutcomePrimarySelected {
			decision.Outcome =
				providerfallback.OutcomeFallbackSelected
		}

		if decision.TriggerReason == "" ||
			decision.TriggerReason ==
				providerbudget.DecisionReasonAllowed {
			decision.TriggerReason =
				providerbudget.DecisionReasonProviderUnavailable
		}
	}

	return decision
}

func trafficProviderEffectiveHealthRank(
	provider providerpolicy.Provider,
	configuredPrimary providerpolicy.Provider,
	status providerhealthdomain.Status,
	recoveryProbeDue bool,
) int {
	if provider == configuredPrimary && recoveryProbeDue {
		return 0
	}

	return trafficProviderHealthRank(status)
}

func trafficProviderRecoveryProbeDue(
	snapshot providerhealthdomain.Snapshot,
) bool {
	if snapshot.Status != providerhealthdomain.StatusUnavailable ||
		snapshot.LastRequestAgeSeconds == nil {
		return false
	}

	return time.Duration(*snapshot.LastRequestAgeSeconds)*time.Second >=
		trafficProviderRecoveryProbeAfter
}

func trafficProviderHealthRank(
	status providerhealthdomain.Status,
) int {
	switch status {
	case providerhealthdomain.StatusHealthy:
		return 0
	case providerhealthdomain.StatusDegraded,
		providerhealthdomain.StatusUnknown:
		return 1
	case providerhealthdomain.StatusUnavailable:
		return 2
	default:
		return 1
	}
}

func validTrafficProviderHealthStatus(
	status providerhealthdomain.Status,
) bool {
	switch status {
	case providerhealthdomain.StatusUnknown,
		providerhealthdomain.StatusHealthy,
		providerhealthdomain.StatusDegraded,
		providerhealthdomain.StatusUnavailable:
		return true
	default:
		return false
	}
}

func trafficProviderHealthStatusLabel(
	status providerhealthdomain.Status,
) string {
	if validTrafficProviderHealthStatus(status) {
		return string(status)
	}

	return fmt.Sprintf("invalid(%s)", status)
}
