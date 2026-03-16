package swipegames

import "net/http"

// Environment represents the target deployment environment.
type Environment string

const (
	// EnvStaging is the staging environment.
	EnvStaging Environment = "staging"
	// EnvProduction is the production environment.
	EnvProduction Environment = "production"
)

var envURLs = map[Environment]string{
	EnvStaging:    "https://staging.platform.0.swipegames.io/api/v1",
	EnvProduction: "https://prod.platform.1.swipegames.io/api/v1",
}

// ClientConfig holds the configuration for a SwipeGames client.
type ClientConfig struct {
	// CID is the SwipeGames-assigned client ID.
	CID string
	// ExtCID is the external client ID.
	ExtCID string
	// APIKey is used to sign outbound requests to the Swipe Games Core API.
	APIKey string
	// IntegrationAPIKey is used to verify inbound reverse calls from the Swipe Games platform.
	IntegrationAPIKey string
	// Env selects the environment. Defaults to EnvStaging if empty.
	Env Environment
	// BaseURL overrides the environment-based URL when set.
	BaseURL string
	// Debug enables debug logging of requests and responses.
	Debug bool
	// HTTPClient is an optional custom HTTP client. Defaults to http.DefaultClient.
	HTTPClient *http.Client
}
