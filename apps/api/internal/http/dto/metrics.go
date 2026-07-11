package dto

import "time"

type ActiveAircraftMetricResponse struct {
	Metric        string                   `json:"metric"`
	Value         int                      `json:"value"`
	WindowMinutes int                      `json:"window_minutes"`
	Scope         MetricScopeResponse      `json:"scope"`
	ObservedFrom  time.Time                `json:"observed_from"`
	ObservedTo    time.Time                `json:"observed_to"`
	CalculatedAt  time.Time                `json:"calculated_at"`
	Confidence    MetricConfidenceResponse `json:"confidence"`
	Sources       []MetricSourceResponse   `json:"sources"`
	Limitations   []string                 `json:"limitations"`
}

type MetricScopeResponse struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

type MetricConfidenceResponse struct {
	Level   string   `json:"level"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

type MetricSourceResponse struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
