package dto

type HealthResponse struct {
	Status string `json:"status"`
}

type VersionResponse struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
	BuiltAt  string `json:"built_at"`
}
