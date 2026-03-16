# Swipe Games Go SDK

## Overview

Go SDK for Swipe Games integrators. Single-package SDK (`swipegames/`) providing:
- **Outbound**: Core API client (CreateNewGame, GetGames, CreateFreeRounds, CancelFreeRounds)
- **Inbound**: Integration adapter request verification and parsing (balance, bet, win, refund)
- **Helpers**: Response builders, error types, re-exported API types

Public API docs: https://swipegames.github.io/public-api/

## Commands

```bash
go build ./...    # build
go test ./...     # run all tests
go vet ./...      # lint
```

## Architecture

```
config.go      - Environment, ClientConfig
client.go      - Core API client (outbound HTTP calls)
crypto.go      - HMAC-SHA256 signing, RFC 8785 JCS canonicalization
verify.go      - Inbound request signature verification + validation
responses.go   - Response builder helpers (NewBetResponse, etc.)
types.go       - Re-exported types from public-api + SDK-specific param structs
errors.go      - APIError, ValidationError, VerifyError
```

All code is in `package swipegames` at the repo root. Users import as `swipegames "github.com/swipegames/integration-sdk-go"`.

All types in `types.go` are aliases of `github.com/swipegames/public-api` generated types. SDK-specific structs (`CreateNewGameParams`, `CreateFreeRoundsParams`, etc.) are defined at the bottom of `types.go`.

## Key Design Decisions

- **Single package at root**: no internal/, no sub-packages. Everything is `package swipegames` at the repo root.
- **Two API keys**: `APIKey` signs outbound Core API requests; `IntegrationAPIKey` verifies inbound reverse calls.
- **Canonical JSON**: all request bodies are sent as RFC 8785 canonical JSON (sorted keys, no whitespace) to match the signed content exactly.
- **Type aliases**: re-export public-api types so SDK users don't need to import public-api directly.

## Dependency on public-api

The `github.com/swipegames/public-api` module provides all generated types (BetRequest, WinRequest, RefundRequest, response types, enums). When the public API spec changes:

### What updates automatically (via `go get -u`)
- Struct fields on request/response types
- New enum values and their `Valid()` methods
- New types added to the spec

### What requires MANUAL updates in this SDK

1. **Validation in `verify.go`**: each `ParseAndVerify*` method manually checks required fields. If a new required field is added to BetRequest/WinRequest/RefundRequest, add it to the validation check.

2. **Param structs in `types.go`**: `CreateNewGameParams`, `CreateFreeRoundsParams`, `CancelFreeRoundsParams` are SDK-owned. New API fields need to be added here.

3. **Body builders in `client.go`**: `buildCreateNewGameBody()` and `buildCreateFreeRoundsBody()` manually construct the request map. New fields must be added to the builder.

4. **Tests**: add test coverage for any new fields or endpoints.

### Update checklist (when bumping public-api)

```bash
# 1. bump dependency
go get github.com/swipegames/public-api@latest

# 2. check what changed in the public-api types
# compare struct fields in the new version vs what verify.go validates

# 3. update verify.go validation for new required fields
# 4. update types.go param structs for new outbound fields
# 5. update client.go body builders for new outbound fields
# 6. update types.go re-exports if new types were added
# 7. add/update tests
# 8. run tests
go test ./...
```

## Code Style

- Comments start lowercase (except godoc for exported symbols)
- Use `fmt.Sprintf` for string operations, not `+` concatenation
- No named returns in long methods
- Go 1.22+ range loops (no variable capture needed)
- Tests use stdlib `testing` (no testify), table-driven with descriptive names using spaces

## Publishing

Go modules don't use a registry. Push to `github.com/swipegames/integration-sdk-go` and tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Users install with: `go get github.com/swipegames/integration-sdk-go@v1.0.0`

Note: any commit pushed to a public repo is fetchable via pseudo-versions, even without a tag. Tag releases to signal stability.
