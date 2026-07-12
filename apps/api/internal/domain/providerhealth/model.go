package providerhealth

import "time"

type Status string

const (
	StatusUnknown     Status = "unknown"
	StatusHealthy     Status = "healthy"
	StatusDegraded    Status = "degraded"
	StatusUnavailable Status = "unavailable"
)

type RequestOutcome string

const (
	RequestOutcomeUnknown         RequestOutcome = "unknown"
	RequestOutcomeSuccess         RequestOutcome = "success"
	RequestOutcomeRateLimited     RequestOutcome = "rate_limited"
	RequestOutcomeUnauthorized    RequestOutcome = "unauthorized"
	RequestOutcomeClientError     RequestOutcome = "client_error"
	RequestOutcomeServerError     RequestOutcome = "server_error"
	RequestOutcomeTimeout         RequestOutcome = "timeout"
	RequestOutcomeNetworkError    RequestOutcome = "network_error"
	RequestOutcomeInvalidResponse RequestOutcome = "invalid_response"
)

type BudgetState string

const (
	BudgetStateUnknown     BudgetState = "unknown"
	BudgetStateAvailable   BudgetState = "available"
	BudgetStateConstrained BudgetState = "constrained"
	BudgetStateExhausted   BudgetState = "exhausted"
)

type ObservationEvidence struct {
	Received int64
	Accepted int64
	Rejected int64
}

type BudgetEvidence struct {
	State     BudgetState
	Limit     int64
	Remaining int64
	ResetsAt  *time.Time
}

type EvaluationInput struct {
	ProviderName string
	EvaluatedAt  time.Time

	FirstRequestAt *time.Time
	LastRequestAt  *time.Time
	LastSuccessAt  *time.Time
	LastFailureAt  *time.Time

	RequestsTotal       int64
	RequestsSuccessful  int64
	ConsecutiveFailures int
	AverageLatency      time.Duration
	LatestOutcome       RequestOutcome

	Observations ObservationEvidence
	Budget       BudgetEvidence
}

type Snapshot struct {
	ProviderName string
	Status       Status
	EvaluatedAt  time.Time

	FirstRequestAt *time.Time
	LastRequestAt  *time.Time
	LastSuccessAt  *time.Time
	LastFailureAt  *time.Time

	RequestsTotal       int64
	RequestsSuccessful  int64
	SuccessRatio        float64
	ConsecutiveFailures int
	AverageLatency      time.Duration
	LatestOutcome       RequestOutcome

	Observations   ObservationEvidence
	RejectionRatio float64
	Budget         BudgetEvidence

	LastRequestAgeSeconds *int64
	LastSuccessAgeSeconds *int64

	Reasons     []string
	Limitations []string
}
