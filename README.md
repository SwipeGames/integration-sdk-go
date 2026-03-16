# @swipegames/integration-sdk-go

Go SDK for Swipe Games integrators. Provides a ready-made client for the [Core API](https://swipegames.github.io/public-api/core), typed request/response interfaces for the [Integration Adapter API](https://swipegames.github.io/public-api/swipegames-integration) (reverse calls), and response builder helpers.

For full API details, see the [Swipe Games Public API documentation](https://swipegames.github.io/public-api/).

## Requirements

- Go 1.22+

## Installation

```bash
go get github.com/swipegames/integration-sdk-go
```

---

## Table of Contents

1. [Client Setup](#client-setup)
2. [Core API (Integrator → Swipe Games)](#core-api-integrator--swipe-games)
3. [Integration Adapter API (Reverse Calls)](#integration-adapter-api-reverse-calls)
4. [Error Handling](#error-handling)
5. [Types Reference](#types-reference)
6. [Debug Mode](#debug-mode)
7. [Development](#development)

---

## Client Setup

The SDK uses two separate API keys:

- **`APIKey`** — used to sign requests **you send** to the Swipe Games Core API
- **`IntegrationAPIKey`** — used to verify reverse calls **you receive** from the Swipe Games platform

```go
import "github.com/swipegames/integration-sdk-go/swipegames"

client, err := swipegames.NewClient(swipegames.ClientConfig{
	CID:               "your-cid-uuid",   // Swipe Games-assigned Client ID (CID)
	ExtCID:            "your-ext-cid",     // Your External Client ID (ExtCID)
	APIKey:            "your-api-key",     // Signs outbound requests to Core API
	IntegrationAPIKey: "your-int-key",     // Verifies inbound reverse calls from platform
	Env:               swipegames.EnvStaging, // EnvStaging | EnvProduction
})
```

### Configuration options

| Option              | Type              | Required | Description                                               |
| ------------------- | ----------------- | -------- | --------------------------------------------------------- |
| `CID`               | `string`          | Yes      | Swipe Games-assigned Client ID (CID)                      |
| `ExtCID`            | `string`          | Yes      | Your External Client ID (ExtCID)                          |
| `APIKey`            | `string`          | Yes      | API key for signing outbound requests to the Core API     |
| `IntegrationAPIKey` | `string`          | Yes      | API key for verifying inbound reverse calls from platform |
| `Env`               | `Environment`     | No       | Environment (defaults to `EnvStaging`)                    |
| `BaseURL`           | `string`          | No       | Custom base URL (overrides `Env`)                         |
| `Debug`             | `bool`            | No       | Enable request/response logging (default `false`)         |
| `HTTPClient`        | `*http.Client`    | No       | Custom HTTP client (defaults to `http.DefaultClient`)     |

### Using a custom base URL

If you need to point to a non-standard environment use `BaseURL` instead of `Env`:

```go
client, err := swipegames.NewClient(swipegames.ClientConfig{
	CID:               "your-cid-uuid",
	ExtCID:            "your-ext-cid",
	APIKey:            "your-api-key",
	IntegrationAPIKey: "your-int-key",
	BaseURL:           "https://customenvironment.platform.0.swipegames.io/api/v1",
})
```

---

## Core API (Integrator → Swipe Games)

The client signs all outbound requests automatically via the `X-REQUEST-SIGN` header using `APIKey`.

### Launch a game

```go
resp, err := client.CreateNewGame(ctx, swipegames.CreateNewGameParams{
	GameID:    "sg_catch_97",     // required
	Demo:      false,             // required
	Platform:  swipegames.PlatformDesktop, // required: PlatformDesktop | PlatformMobile
	Currency:  "USD",             // required
	Locale:    "en_us",           // required
	SessionID: "session-123",     // optional
	ReturnURL: "https://...",     // optional: redirect after game ends
	DepositURL: "https://...",    // optional: redirect for deposits
	InitDemoBalance: "1000",      // optional: starting balance for demo mode
	User: &swipegames.User{      // optional
		Id:        "player-123",
		FirstName: strPtr("John"),
		LastName:  strPtr("Doe"),
		NickName:  strPtr("johnny"),
		Country:   strPtr("US"),
	},
})
// resp.GameURL → URL to launch the game
// resp.GsID   → game session ID
```

### List available games

```go
games, err := client.GetGames(ctx)
// Returns []GameInfo — see Types Reference below
```

### Create a free rounds campaign

See [Free Rounds](https://swipegames.github.io/public-api/free-rounds) for details on campaign configuration and behavior.

```go
resp, err := client.CreateFreeRounds(ctx, swipegames.CreateFreeRoundsParams{
	ExtID:      "campaign-1",                    // required: your campaign ID
	Currency:   "USD",                           // required
	Quantity:   10,                              // required: number of free rounds
	BetLine:    5,                               // required
	ValidFrom:  "2026-01-01T00:00:00.000Z",     // required: ISO 8601
	ValidUntil: "2026-02-01T00:00:00.000Z",     // optional: ISO 8601
	GameIDs:    []string{"sg_catch_97"},         // optional: restrict to specific games
	UserIDs:    []string{"player-123"},          // optional: restrict to specific users
})
```

### Cancel a free rounds campaign

```go
// Cancel by Swipe Games ID
err := client.CancelFreeRounds(ctx, swipegames.CancelFreeRoundsParams{ID: "fr-123"})

// Or cancel by your external ID
err := client.CancelFreeRounds(ctx, swipegames.CancelFreeRoundsParams{ExtID: "campaign-1"})
```

---

## Integration Adapter API (Reverse Calls)

When a game session is active, Swipe Games makes [reverse calls](https://swipegames.github.io/public-api/swipegames-integration) to your server for balance checks and wallet operations. You must implement 4 endpoints:

| Endpoint   | Method | Purpose               |
| ---------- | ------ | --------------------- |
| `/balance` | GET    | Get player balance    |
| `/bet`     | POST   | Deduct bet amount     |
| `/win`     | POST   | Credit win amount     |
| `/refund`  | POST   | Refund a previous bet |

During [free rounds](https://swipegames.github.io/public-api/free-rounds), bet/win requests arrive with `type: "free"` and an `frID` — see the docs for how these should be handled.

All reverse calls are signed by the platform with your `IntegrationAPIKey`. The client provides typed methods to verify and parse them — no need to pass keys manually.

### Parse & verify (recommended)

The `ParseAndVerify*` methods verify the signature, parse the body, validate required fields, and return a typed result.

To build your responses, use the response helpers:

```go
swipegames.NewBalanceResponse(balance)
swipegames.NewBetResponse(balance, txID)
swipegames.NewWinResponse(balance, txID)
swipegames.NewRefundResponse(balance, txID)
swipegames.NewErrorResponse(swipegames.ErrorResponseOpts{...})
```

#### GET /balance

```go
func handleGetBalance(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{}
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}
	sig := r.Header.Get("X-REQUEST-SIGN")

	query, verifyErr := client.ParseAndVerifyBalanceRequest(params, sig)
	if verifyErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(verifyErr.Response())
		return
	}

	// Your logic: look up the player's balance using the session ID.
	balance := getPlayerBalance(query.SessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swipegames.NewBalanceResponse(balance))
}
```

#### POST /bet

```go
func handleBet(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	sig := r.Header.Get("X-REQUEST-SIGN")

	bet, verifyErr := client.ParseAndVerifyBetRequest(string(body), sig)
	if verifyErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(verifyErr.Response())
		return
	}

	// Your logic: deduct the bet amount and record the transaction.
	newBalance := deductFromWallet(bet.SessionID, bet.Amount)
	partnerTxID := saveTransaction(bet.TxID, bet.RoundID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swipegames.NewBetResponse(newBalance, partnerTxID))
}
```

#### POST /win

```go
func handleWin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	sig := r.Header.Get("X-REQUEST-SIGN")

	win, verifyErr := client.ParseAndVerifyWinRequest(string(body), sig)
	if verifyErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(verifyErr.Response())
		return
	}

	// Your logic: credit the win amount and record the transaction.
	newBalance := creditToWallet(win.SessionID, win.Amount)
	partnerTxID := saveTransaction(win.TxID, win.RoundID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swipegames.NewWinResponse(newBalance, partnerTxID))
}
```

#### POST /refund

```go
func handleRefund(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	sig := r.Header.Get("X-REQUEST-SIGN")

	refund, verifyErr := client.ParseAndVerifyRefundRequest(string(body), sig)
	if verifyErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(verifyErr.Response())
		return
	}

	// Your logic: refund the original transaction and record the refund.
	newBalance := refundToWallet(
		refund.SessionID,
		refund.OrigTxID,
		refund.Amount,
	)
	partnerTxID := saveRefundTransaction(refund.TxID, refund.OrigTxID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swipegames.NewRefundResponse(newBalance, partnerTxID))
}
```

### Verify only (lower-level)

If you need just the boolean check without parsing (e.g. you already parsed the body yourself):

```go
// POST body verification (/bet, /win, /refund)
valid := client.VerifyBetRequest(rawBody, signatureHeader)
valid := client.VerifyWinRequest(rawBody, signatureHeader)
valid := client.VerifyRefundRequest(rawBody, signatureHeader)

// GET /balance query param verification
valid := client.VerifyBalanceRequest(queryParams, signatureHeader)
```

---

## Error Handling

### Core API errors

The SDK returns two error types:

- **`*APIError`** — API returned a non-2xx response
- **`*ValidationError`** — Request params failed client-side validation before the request was sent

```go
import "errors"

resp, err := client.CreateNewGame(ctx, params)
if err != nil {
	var valErr *swipegames.ValidationError
	if errors.As(err, &valErr) {
		fmt.Println(valErr.Message) // Validation error summary
		fmt.Println(valErr.Field)   // Field that failed validation (if applicable)
	}

	var apiErr *swipegames.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode) // HTTP status code (e.g. 401, 404, 500)
		fmt.Println(apiErr.Message)    // Error message from the platform
		fmt.Println(apiErr.Code)       // Optional error code
		fmt.Println(apiErr.Details)    // Optional additional details
	}
}
```

### Reverse call error responses

When you need to return an error from your reverse call handlers, use `NewErrorResponse()`:

```go
// Simple error
swipegames.NewErrorResponse(swipegames.ErrorResponseOpts{
	Message: "Player not found",
})
// → {"message": "Player not found"}

// Error with code
swipegames.NewErrorResponse(swipegames.ErrorResponseOpts{
	Message: "Insufficient funds",
	Code:    swipegames.ErrorCodeInsufficientFunds,
})
// → {"message": "Insufficient funds", "code": "insufficient_funds"}

// Error with code and action
swipegames.NewErrorResponse(swipegames.ErrorResponseOpts{
	Message: "Session has expired",
	Code:    swipegames.ErrorCodeSessionExpired,
	Action:  swipegames.ErrorActionRefresh,
})
// → {"message": "Session has expired", "code": "session_expired", "action": "refresh"}
```

#### Available error codes

| Code                      | Description                           |
| ------------------------- | ------------------------------------- |
| `game_not_found`          | Game does not exist                   |
| `currency_not_supported`  | Currency not supported                |
| `locale_not_supported`    | Locale not supported                  |
| `account_blocked`         | Player account is blocked             |
| `bet_limit`               | Bet limit exceeded                    |
| `loss_limit`              | Loss limit exceeded                   |
| `time_limit`              | Time limit exceeded                   |
| `insufficient_funds`      | Not enough balance                    |
| `session_expired`         | Session has expired                   |
| `session_not_found`       | Session does not exist                |
| `client_connection_error` | Connection error to integrator system |

---

## Types Reference

All request/response types are derived from the [`github.com/swipegames/public-api`](https://github.com/swipegames/public-api) package and re-exported from this SDK as type aliases. You can import them from either package:

```go
// Via the SDK (recommended)
import "github.com/swipegames/integration-sdk-go/swipegames"
var user swipegames.User

// Or directly from public-api
import apiv1 "github.com/swipegames/public-api/api/v1.0"
var user apiv1.User
```

See [`swipegames/types.go`](swipegames/types.go) for the full list of re-exported types and constants.

---

## Debug Mode

Enable debug logging to see all Core API requests and responses:

```go
client, err := swipegames.NewClient(swipegames.ClientConfig{
	CID:               "your-cid-uuid",
	ExtCID:            "your-ext-cid",
	APIKey:            "your-api-key",
	IntegrationAPIKey: "your-int-key",
	Env:               swipegames.EnvStaging,
	Debug:             true,
})
```

When enabled, the SDK logs with a `[SwipeGamesSDK]` prefix:

```
[SwipeGamesSDK] POST https://staging.platform.0.swipegames.io/api/v1/create-new-game
[SwipeGamesSDK] Body: {"cID":"...","extCID":"...","gameID":"sg_catch_97",...}
[SwipeGamesSDK] POST https://staging.platform.0.swipegames.io/api/v1/create-new-game -> 200
```

On errors:

```
[SwipeGamesSDK] GET https://staging.platform.0.swipegames.io/api/v1/games -> 401
[SwipeGamesSDK] ERROR: GET ... error: {message: "Invalid signature"}
```

---

## Development

```bash
go test ./...     # Run tests
go build ./...    # Build
go vet ./...      # Lint
```
