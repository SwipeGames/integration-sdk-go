# Auto-Update Go SDK for swipegames/public-api

You are updating `github.com/swipegames/integration-sdk-go` to match a new version of `github.com/swipegames/public-api`.

## Environment

These env vars provide context (all optional — detect from local state if missing):

- `PUBLIC_API_DIFF_FILE` — path to a file containing the git diff between versions
- `PUBLIC_API_REPO_PATH` — path to the public-api repo checkout (default: `../public-api`)
- `CURRENT_VERSION` — currently installed version of `swipegames/public-api`
- `TARGET_VERSION` — target version to update to

## Step 1: Gather context

1. **Detect versions** (if not provided via env vars):
   - Current: `grep 'github.com/swipegames/public-api/api/v1.0/core/types' go.mod | awk '{print $2}' | sed 's/^v//'`
   - Target: `cd public-api && git describe --tags --abbrev=0 | sed 's/^v//'`

2. **Read the public-api diff** to understand what changed:
   - If `PUBLIC_API_DIFF_FILE` is set and exists, read that file
   - Otherwise, in the public-api repo (`PUBLIC_API_REPO_PATH` or `../public-api`), run:
     ```
     git diff v{current}..v{target} -- api/ docs/ ':!**/package-lock.json' ':!**/node_modules' ':!**/*.ts' ':!**/*.js' ':!**/schemas/*.schema.mdx'
     ```
   - If the public-api repo is not available, inspect the Go module cache or `go doc` output to understand the current API surface

3. **Read the public-api source** for full context:
   - In the public-api repo: `api/` and `docs/` directories
   - Pay special attention to `docs/changes-log.md` for a human-readable changelog
   - Look at Go generated types (`.gen.go` files) in `api/v1.0/core/types/` and `api/v1.0/swipegames-integration/types/`

4. **Categorize the changes:**
   - New Core API endpoints (-> new methods on the client in `client.go`)
   - New Integration (reverse-call) endpoints (-> new verify/parse methods + response builders)
   - Changed request/response schemas (-> update existing types, validation, and body builders)
   - New shared types or enums (-> update type aliases in `types.go`)
   - New error codes or actions (-> update constants in `types.go`)
   - Breaking changes (-> document in PR body)

## Step 2: Read the current SDK

Read these files to understand existing patterns:

1. `CLAUDE.md` — project architecture overview
2. `types.go` — re-exported type aliases and SDK-specific param structs
3. `client.go` — outbound API client + body builders
4. `verify.go` — inbound request signature verification + validation
5. `responses.go` — response builder helpers
6. `crypto.go` — HMAC-SHA256 signing internals
7. `errors.go` — error types (APIError, ValidationError, VerifyError)
8. `config.go` — Environment, ClientConfig
9. All `*_test.go` files — test patterns and style

## Step 3: Update the SDK

Apply changes following existing patterns exactly. For each change type:

### New Core API endpoint

1. **`client.go`**: Add a new public method following the pattern of `CreateNewGame` or `GetGames`:
   - Use the appropriate HTTP method
   - Build the request path from the endpoint
   - Sign the request with `c.config.APIKey` using `signRequest`
   - Build the request body using canonical JSON (RFC 8785)
   - Parse the JSON response into the appropriate type
   - Return the typed result and error

2. **`types.go`**: Add a new `{Name}Params` struct at the bottom (SDK-specific param structs section) and add type aliases for any new public-api types

3. **`client.go`**: Add a `build{Name}Body()` function following the pattern of `buildCreateNewGameBody()`

### New Integration (reverse-call) endpoint

1. **`verify.go`**: Add two methods following existing patterns:
   - `Verify{Name}Request(headers http.Header, rawBody []byte) error` — signature verification only
   - `ParseAndVerify{Name}Request(headers http.Header, rawBody []byte) (*{Name}Request, error)` — verify + parse body + validate required fields

2. **`responses.go`**: Add a `New{Name}Response(data)` builder function following existing patterns

3. **`types.go`**: Add type aliases for the new request/response types from public-api

### Changed schemas (new fields, removed fields)

1. Update validation in `verify.go` — each `ParseAndVerify*` method manually checks required fields. If a new required field is added, add it to the validation check
2. Update SDK-specific param structs in `types.go` if outbound request fields changed
3. Update body builders in `client.go` (`buildCreateNewGameBody()`, `buildCreateFreeRoundsBody()`) if new fields need to be included
4. If fields were removed, check for usages in tests and update accordingly

### New error codes or actions

1. **`types.go`**: Add new type aliases or constants
2. Update response builders in `responses.go` if they need to handle new codes

### New shared types

1. **`types.go`**: Add type aliases to re-export from public-api so SDK users don't need to import public-api directly
2. Follow the pattern: `type NewTypeName = publictypes.NewTypeName`

## Step 4: Update tests

Follow existing Go test patterns exactly:

1. Read existing `*_test.go` files to understand the style
2. Use stdlib `testing` package — no testify or external test frameworks
3. Use table-driven tests with descriptive names using spaces (e.g., `"valid bet request with all fields"`)
4. Add tests for every new method or changed behavior
5. Test both success and error paths
6. For new client methods: test request signing, URL construction, response parsing
7. For new verify/parse methods: test valid signatures, invalid signatures, malformed bodies, missing required fields
8. For new response builders: test output shape and field correctness
9. Use Go 1.22+ range loops (no variable capture needed in loop bodies)

## Step 5: Verify

1. Run `go test ./...` — all tests must pass
2. Run `go vet ./...` — no lint issues
3. Run `go build ./...` — build succeeds
4. If any command fails, read the error output carefully, fix the issues, and re-run
5. Repeat until all three pass cleanly

## Step 6: Write changes summary

Write a concise markdown summary of the SDK changes to `/tmp/sdk-changes-summary.md`. This will be included in the PR description. Focus on what changed from the **SDK user's perspective**:
- New methods, types, or response builders added
- Changed method signatures or validation rules
- New fields on param structs or response types
- Any breaking changes

Use bullet points. Do NOT include internal implementation details or test changes.

## Constraints

- **Single package at root**: all code is `package swipegames` at the repo root — no `internal/`, no sub-packages
- **Type aliases for re-exports**: use `type Foo = publictypes.Foo`, not wrapper types
- **RFC 8785 canonical JSON**: all request bodies are sent as canonical JSON (sorted keys, no whitespace) to match the signed content exactly
- **Two-key architecture**: `APIKey` signs outbound Core API requests; `IntegrationAPIKey` verifies inbound reverse calls
- **Do NOT modify `crypto.go`** unless the signing/verification mechanism itself changed
- **Backward compatible**: don't remove or rename existing public API unless the upstream change is breaking
- **No unnecessary changes**: only modify what's needed to support the new public-api version
- **Go 1.22+ range loops**: no variable capture needed in for-range loops
- **Comments start lowercase** (except godoc for exported symbols)
- **Use `fmt.Sprintf`** for string operations, not `+` concatenation
- **stdlib `testing`** only — no testify or external test frameworks
- **Keep imports from `public-api` aligned**: use `coretypes` and `integrationtypes` import aliases as appropriate
