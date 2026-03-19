package swipegames

import (
	coreapiv1 "github.com/swipegames/public-api/api/v1.0/core/types"
	integrationapiv1 "github.com/swipegames/public-api/api/v1.0/swipegames-integration/types"
)

// ── Re-exported types from public-api (core) ──

// User represents a player.
type User = coreapiv1.User

// ErrorResponse is the core API error response.
type ErrorResponse = coreapiv1.ErrorResponse

// PlatformType is the game platform (desktop/mobile).
type PlatformType = coreapiv1.PlatformType

// CreateNewGameResponse is returned when a new game session is created.
type CreateNewGameResponse = coreapiv1.CreateNewGameResponse

// CreateFreeRoundsResponse is returned when a free rounds campaign is created.
type CreateFreeRoundsResponse = coreapiv1.CreateFreeRoundsResponse

// GameInfo contains information about a game.
type GameInfo = coreapiv1.GameInfo

// GameInfoImages contains image URLs for a game.
type GameInfoImages = coreapiv1.GameInfoImages

// BetLineInfo contains bet line information for a specific currency.
type BetLineInfo = coreapiv1.BetLineInfo

// BetLineValue represents a single bet line value.
type BetLineValue = coreapiv1.BetLineValue

// ── Re-exported types from public-api (integration) ──

// BetRequest represents an incoming bet request from the platform.
type BetRequest = integrationapiv1.BetRequest

// BetRequestType is the type of a bet transaction (regular/free).
type BetRequestType = integrationapiv1.BetRequestType

// WinRequest represents an incoming win request from the platform.
type WinRequest = integrationapiv1.WinRequest

// WinRequestType is the type of a win transaction (regular/free).
type WinRequestType = integrationapiv1.WinRequestType

// RefundRequest represents an incoming refund request from the platform.
type RefundRequest = integrationapiv1.RefundRequest

// BalanceResponse is the response for a balance request.
type BalanceResponse = integrationapiv1.BalanceResponse

// BetResponse is the response for a bet request.
type BetResponse = integrationapiv1.BetResponse

// WinResponse is the response for a win request.
type WinResponse = integrationapiv1.WinResponse

// RefundResponse is the response for a refund request.
type RefundResponse = integrationapiv1.RefundResponse

// ErrorResponseWithCodeAndAction is an error response with optional code and action fields.
type ErrorResponseWithCodeAndAction = integrationapiv1.ErrorResponseWithCodeAndAction

// ErrorResponseWithCodeAndActionCode is the error code for integration error responses.
type ErrorResponseWithCodeAndActionCode = integrationapiv1.ErrorResponseWithCodeAndActionCode

// ErrorResponseWithCodeAndActionAction is the action for integration error responses.
type ErrorResponseWithCodeAndActionAction = integrationapiv1.ErrorResponseWithCodeAndActionAction

// CoreErrorResponseCode is the error code for core API error responses.
type CoreErrorResponseCode = coreapiv1.ErrorResponseCode

// ── Re-exported constants ──

const (
	// Platform types
	PlatformDesktop PlatformType = coreapiv1.Desktop
	PlatformMobile  PlatformType = coreapiv1.Mobile

	// Bet request types
	BetRequestTypeRegular BetRequestType = integrationapiv1.BetRequestTypeRegular
	BetRequestTypeFree    BetRequestType = integrationapiv1.BetRequestTypeFree

	// Win request types
	WinRequestTypeRegular WinRequestType = integrationapiv1.WinRequestTypeRegular
	WinRequestTypeFree    WinRequestType = integrationapiv1.WinRequestTypeFree

	// Error codes (integration)
	ErrorCodeAccountBlocked        ErrorResponseWithCodeAndActionCode = integrationapiv1.AccountBlocked
	ErrorCodeBetLimit              ErrorResponseWithCodeAndActionCode = integrationapiv1.BetLimit
	ErrorCodeClientConnectionError ErrorResponseWithCodeAndActionCode = integrationapiv1.ClientConnectionError
	ErrorCodeCurrencyNotSupported  ErrorResponseWithCodeAndActionCode = integrationapiv1.CurrencyNotSupported
	ErrorCodeGameNotFound          ErrorResponseWithCodeAndActionCode = integrationapiv1.GameNotFound
	ErrorCodeInsufficientFunds     ErrorResponseWithCodeAndActionCode = integrationapiv1.InsufficientFunds
	ErrorCodeLocaleNotSupported    ErrorResponseWithCodeAndActionCode = integrationapiv1.LocaleNotSupported
	ErrorCodeLossLimit             ErrorResponseWithCodeAndActionCode = integrationapiv1.LossLimit
	ErrorCodeSessionExpired        ErrorResponseWithCodeAndActionCode = integrationapiv1.SessionExpired
	ErrorCodeSessionNotFound       ErrorResponseWithCodeAndActionCode = integrationapiv1.SessionNotFound
	ErrorCodeTimeLimit             ErrorResponseWithCodeAndActionCode = integrationapiv1.TimeLimit

	// Error actions
	ErrorActionRefresh ErrorResponseWithCodeAndActionAction = integrationapiv1.Refresh

	// Core API error codes
	CoreErrorCodeAccountBlocked       CoreErrorResponseCode = coreapiv1.AccountBlocked
	CoreErrorCodeCurrencyNotSupported CoreErrorResponseCode = coreapiv1.CurrencyNotSupported
	CoreErrorCodeGameNotFound         CoreErrorResponseCode = coreapiv1.GameNotFound
	CoreErrorCodeLocaleNotSupported   CoreErrorResponseCode = coreapiv1.LocaleNotSupported
)

// ── SDK-specific types (not in public-api) ──

// CreateNewGameParams contains parameters for creating a new game session.
// CID and ExtCID are automatically added by the client.
type CreateNewGameParams struct {
	GameID          string       `json:"gameID"`
	Demo            bool         `json:"demo"`
	Platform        PlatformType `json:"platform"`
	Currency        string       `json:"currency"`
	Locale          string       `json:"locale"`
	SessionID       string       `json:"sessionID,omitempty"`
	ReturnURL       string       `json:"returnURL,omitempty"`
	DepositURL      string       `json:"depositURL,omitempty"`
	InitDemoBalance string       `json:"initDemoBalance,omitempty"`
	User            *User        `json:"user,omitempty"`
}

// CreateFreeRoundsParams contains parameters for creating a free rounds campaign.
// CID and ExtCID are automatically added by the client.
type CreateFreeRoundsParams struct {
	ExtID      string   `json:"extID"`
	GameIDs    []string `json:"gameIDs,omitempty"`
	UserIDs    []string `json:"userIDs,omitempty"`
	Currency   string   `json:"currency"`
	Quantity   int      `json:"quantity"`
	BetLine    int      `json:"betLine"`
	ValidFrom  string   `json:"validFrom"`
	ValidUntil string   `json:"validUntil,omitempty"`
}

// CancelFreeRoundsParams contains parameters for canceling a free rounds campaign.
// At least one of ID or ExtID must be provided.
// CID and ExtCID are automatically added by the client.
type CancelFreeRoundsParams struct {
	ID    string `json:"id,omitempty"`
	ExtID string `json:"extID,omitempty"`
}

// GetBalanceQuery represents parsed query parameters for a balance request.
type GetBalanceQuery struct {
	SessionID string `json:"sessionID"`
}
