---
name: swipegames
description: |
  SwipeGames platform development guidance for Go microservices gaming platform with PostgreSQL, Redis, and Temporal workflows. Use when working in the swipegames directory (at any depth) for: (1) Writing new Go code (services, controllers, usecases, repositories), (2) Writing tests (integration or unit tests), (3) Fixing bugs or debugging issues, (4) Refactoring code, (5) Working with databases (migrations, queries, partitioning), (6) Developing APIs (OpenAPI, gRPC, HTTP), (7) Building integration adapters, (8) Implementing game logic, (9) Working with financial transactions, (10) Any code or architecture tasks in swipegames projects.
---

# SwipeGames Platform Development

This skill provides comprehensive guidance for developing within the SwipeGames gaming platform - a Go microservices architecture with real-time betting, casino integrations, and financial transaction processing.

## Platform Overview

**Architecture**: Go microservices with gRPC communication, PostgreSQL with partitioning, Redis clustering, Temporal workflows

**Key Services**:
- **Core Service**: Financial engine with atomic operations and server-side RNG
- **Integration Services**: Casino operator adapters with DB-first transactions
- **Game Services**: Frontend APIs with WebSocket real-time communication
- **Internal Services**: Shared utilities and frameworks

**Technology Stack**: Go 1.24.5+, PostgreSQL 15+, Redis 7, Temporal, go-jet, goose, zerolog

## Quick Reference - Critical Patterns

### DB-First Transaction Pattern (MUST FOLLOW)
For all financial operations in integration services:

1. **Validate request** - Check all parameters first
2. **Save to DB with `pending_api_call` status** - Persist BEFORE external API call
3. **Make external API call** - Call partner's API
4. **Update transaction with API response** - Save result
5. **Commit transaction** - Only after successful DB update

**Error handling**:
- Refundable errors (5xx, timeouts): Commit with error, background service retries
- Non-refundable errors (4xx): Rollback entire transaction

**Why**: Prevents money loss if service crashes between API call and DB update.

### Error Handling Pattern (MUST FOLLOW)
Use `domain.Error` for all service communication:

```go
// Creating errors
err := domain.NewNotFoundError("Session not found", dbErr).
    WithCode(domain.ErrorCodeSessionNotFound)

// Proxying gRPC errors (don't wrap, pass as-is)
if parsedErr, parseErr := domain.ParseError(err.Error()); parseErr == nil {
    return nil, parsedErr  // Pass through unchanged
}
```

**Rules**:
- Lower layers specify root cause
- Top layer (controller) returns errors as-is (no wrapping)
- Don't log when returning errors (middleware logs)
- Between services always use `domain.Error` (even gRPC)

### Domain Types Pattern (MUST FOLLOW)
Use domain types for all parameters:

```go
type GameID string
type SessionID uuid.UUID

func GetGame(gameID GameID) (*Game, error)  // NOT string
```

**Benefits**: Type safety, prevents mixing up parameters, more readable code

### Input Validation Pattern (MUST FOLLOW)
All validation happens in controller layer or value objects before business logic:

**CID validation** - Use domain types from common library:
```go
import domain2 "github.com/swipegames/platform-lib-common/domain"

cid, err := domain2.NewCIDFromString(req.Cid.String())
if err != nil {
    return domain2.NewBadRequestError("CID is required", nil)
}
```

**String/ExtCID validation** - Use utils from common library:
```go
import utils2 "github.com/swipegames/platform-lib-common/utils"

if utils2.IsEmpty(req.ExtCID) {
    return domain2.NewBadRequestError("ExtCID is required", nil)
}
```

**Common library location**: `platform-lib-common` repository contains all shared code, utilities, and domain types.

### Layered Architecture (MUST FOLLOW)
Clear separation of concerns:

```
controller/v1/  → HTTP/gRPC handlers, request validation
service/        → Business logic orchestration
repository/     → Database operations
usecases/v1/    → Complex business logic (optional)
domain/         → Domain models
```

**Versioning**: Use `v1`, `v2` in controller and usecases for API version support

## Reference Documentation

Load specific references based on your task:

### Architecture & Infrastructure
See **[references/architecture.md](references/architecture.md)** for:
- Service architecture and responsibilities
- Technology stack details
- Temporal worker framework
- Build commands and deployment
- Monitoring and alerting

### Code Style & Conventions
See **[references/code-style-guide.md](references/code-style-guide.md)** for:
- Directory structure and file naming
- Layered architecture patterns
- Domain types and model separation
- Import naming conventions
- Common libraries usage
- Testing organization

### Error Handling
See **[references/error-handling.md](references/error-handling.md)** for:
- domain.Error structure and usage
- Serialization format
- Service communication patterns
- Creating and proxying errors
- Standard error codes
- Client actions

### Database
See **[references/database.md](references/database.md)** for:
- Table naming and conventions
- Partitioning with pg_partman
- Partition pruning strategies
- Migrations and fixtures
- Currency handling
- Indexes and constraints

### Integration Adapters
See **[references/integrations.md](references/integrations.md)** for:
- Integration client pattern (integration.go with raw JSON requests)
- DB-first transaction pattern (detailed)
- BetWin operations
- Idempotency implementation
- HMAC authentication
- Temporal workflows
- Testing integrations

### Game Development
See **[references/games.md](references/games.md)** for:
- Game naming strategy
- Game launch and demo mode
- Game config and round concepts
- Game sessions and settings
- Free rounds implementation
- Bonus balance processing

### API Development
See **[references/api-development.md](references/api-development.md)** for:
- API versioning and backward compatibility
- OpenAPI operationId naming
- Request/response formatting
- Code generation
- Authentication headers

## Common Development Tasks

### Writing New Service Code

1. **Validate in controller** - Use domain2.NewCIDFromString for CIDs, utils2.IsEmpty for strings
2. Follow layered architecture (controller → service → repository)
3. Use domain types for all parameters
4. Separate API, domain, and persistence models
5. Import common libs from `platform-lib-common` with `2` suffix
6. **Critical operations**: Log with Info level, include correlation IDs
7. **After big changes**: Run `make lint <service_name>` or use VSCode MCP to check linter errors

### Code Quality Workflow

**After making significant changes:**

1. **Run linter** - Check for code quality issues:
   ```bash
   make lint <service_name>
   ```
   Or use VSCode MCP tasks to check for linter errors

2. **Run affected tests** - If you changed a method with tests:
   - Check if unit tests exist in same package (`*_test.go`)
   - Check if integration tests exist in `tests/` directory
   - **Run the test** to ensure it still passes

3. **Before integration tests** - Services must be running:
   - Check if service already running: `docker ps | grep <service_name>`
   - If not running: `make up <service_name>`
   - Or ask user: "Should I start the service with `make up <service_name>` first?"

**Example workflow:**
```bash
# Made changes to BetService.PlaceBet method
# 1. Check linter
make lint platform

# 2. Run related tests
# - Unit test: services/platform/internal/service/bet_test.go
# - Integration test: services/platform/tests/bet_test.go

# 3. Start service if needed (for integration tests)
make up platform

# 4. Run integration test
# Use VSCode MCP task or make test platform
```

### Writing Tests

**Integration tests** (preferred):
```go
// Base on IntegrationTestSuite from services/internal/tests
type MyTestSuite struct {
    tests.IntegrationTestSuite
}
```

**Use fixtures** from `db/fixtures/` for static test data
**Clean up** dynamic test data after tests

**E2E API Test Entity Naming:**

When creating new tests for e2e-api-tests, always use the `test_` prefix for all test entities to clearly distinguish them from real requests:

```go
// Test entity IDs - use test_ prefix
accountID := "test_account_123"
userID := "test_user_456"
roundID := "test_round_789"
sessionID := "test_session_abc"
extCID := "test_ext_cid_xyz"

// This makes it easy to:
// - Identify test data in logs and databases
// - Filter test requests from real requests
// - Clean up test data after tests
```

**Always use `test_` prefix for:**
- Account IDs
- User IDs
- Round IDs
- Session IDs
- External CIDs
- Any other identifiers created for testing purposes

**Running Tests:**

1. **Before integration tests** - Ensure service is running:
   - Check: `docker ps | grep <service_name>`
   - Start if needed: `make up <service_name>`
   - Or verify with user before running

2. **After changing code** - Run affected tests:
   - Changed method in `service/bet.go`? Run `tests/bet_test.go`
   - Changed controller? Run related controller tests
   - Always run tests to verify changes don't break existing functionality

3. **Use VSCode MCP** - Preferred method for running tests:
   - Automatically handles service dependencies
   - Shows test output inline
   - Easier to debug failures

### Working with Database

```bash
# Create new migration
make add-migration <service_name>

# Apply migrations (requires full reset)
make down
make up <service_name>

# Generate models after schema changes
make gen-db <service_name>
```

**Partitioning**: Always add `created_at` filter for partition pruning

### Fixing Bugs

1. Check if financial operation - verify DB-first pattern
2. Check error handling - ensure domain.Error used correctly
3. Check logging - critical ops must log with Info level
4. Check tests - add test case reproducing bug
5. Never log secrets, API keys, or plain-text money amounts

### Building Integration Adapter

1. Use 3-layer architecture (controller, service, repository)
2. **Create integration.go** - Use raw JSON requests (NOT generated clients), sign with RequestSigner
3. **Must implement** DB-first transaction pattern
4. Add idempotency support (ext_tx_id field)
5. Implement HMAC authentication
6. Create Temporal workflows for API calls
7. Add mock service for testing
8. Configure Istio routing

## Key Commands

```bash
# Start service with dependencies
make up <service_name>

# Build service
make build <service_name>

# Run tests
make test <service_name>

# Run linting
make lint <service_name>

# Generate database models
make gen-db <service_name>

# Create migration
make add-migration <service_name>

# Stop all services
make down
```

## Important Notes

### Security
- Never log secrets, API keys, or financial amounts in plain text
- All external API calls must include correlation IDs
- Server-side outcome calculation is mandatory
- Use Redis for session management, not PostgreSQL

### Financial Operations
- All money transactions processed atomically
- Double-entry ledger principles
- Comprehensive transaction logging with idempotency
- DB-first pattern for integration services

### Temporal Workflows
- One worker per service
- Single task queue per service
- Automatic retry with exponential backoff
- Register workflows/activities on worker before app start

### Testing
- WireMock for external API simulation
- Fixtures for static test data
- Integration tests preferred over unit tests
- Every service needs main test suite
- **Before integration tests**: Verify service is running with `docker ps` or start with `make up <service>`
- **After code changes**: Always run affected tests to verify functionality
- **Use VSCode MCP**: Preferred for running tests (handles dependencies automatically)
- **E2E API tests**: Always use `test_` prefix for all test entity IDs (accountID, userID, roundID, etc.)

## When to Load References

- **Starting new service**: Load architecture.md and code-style-guide.md
- **Working with errors**: Load error-handling.md
- **Database migrations**: Load database.md
- **Building integration**: Load integrations.md
- **Game development**: Load games.md
- **API changes**: Load api-development.md

## Critical Reminders

1. **Input Validation**: Always validate in controller using domain2.NewCIDFromString for CIDs and utils2.IsEmpty for strings
2. **DB-First Pattern**: Always save to database before external API calls in integrations
3. **Domain Errors**: Use domain.Error for all inter-service communication
4. **Domain Types**: Create typed wrappers for IDs and business concepts
5. **Partition Pruning**: Always filter by created_at for partitioned tables
6. **Logging**: Info level for money operations, never log secrets
7. **Linting**: Run `make lint <service>` after big changes, use VSCode MCP to check errors
8. **Testing**: Run tests after changing methods, ensure services running for integration tests
9. **Service Dependencies**: Before integration tests, check if service is running or start with `make up <service>`
10. **Import Naming**: Use `2` suffix for common lib imports (platform-lib-common)
11. **Test Entity Naming**: Use `test_` prefix for all test entity IDs in e2e-api-tests (accountID, userID, roundID, etc.)
