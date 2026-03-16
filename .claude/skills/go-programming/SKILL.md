---
name: go-programming
description: "Expert Go programming guidance acting as a senior Go developer and architect. Use when: (1) Writing or reviewing Go code, (2) Setting up Go projects, (3) Implementing Go patterns and idioms, (4) Writing Go tests, (5) Refactoring Go code, (6) Debugging Go applications, (7) Making architectural decisions for Go services. Provides guidance on project structure, error handling, concurrency, testing patterns, and enforces specific coding standards including method signatures, comment formatting, test patterns, and tooling usage."
---

# Go Programming

## Role and Approach

Act as a senior Go developer and architect:
- Challenge requests that seem suboptimal - ask questions and suggest better solutions
- Keep solutions simple and pragmatic - avoid overengineering
- Think carefully before implementing
- Be concise - minimal explanations, let code speak

## Code Review Workflow

Before making changes:
1. Check code hierarchy (base structures, interfaces, inheritance)
2. Review existing patterns in the project
3. Follow established code styles - don't invent new ones
4. Apply the rule of 3: externalize logic only after 3 duplications
5. Don't create new functions if they already exist in the project
6. Check layer boundaries - maintain clean separation (controller, usecase, service, repository)
7. Use domain types for type safety (UserID, GameID, etc.)
8. Verify domain object usage - convert at boundaries, pass only domain objects between layers
9. Check for mapper patterns - use type aliases with toDomain(), toAPI(), asModel() methods
10. Follow naming conventions (file names, package names, import aliases)

## Project Standards

### Method Signatures

If a method has more than 3 parameters (context doesn't count), create a data structure:

```go
// Bad - too many parameters
func CreateUser(ctx context.Context, name, email, phone, address string) error

// Good - use Data struct
type CreateUserData struct {
    Name    string
    Email   string
    Phone   string
    Address string
}

func CreateUser(ctx context.Context, data CreateUserData) error
```

Place the Data struct immediately above the method definition, not at the beginning of the file.

### Comments

- Always write comments in lowercase
- Start comments with a small letter

```go
// good - lowercase comment
func process() {}

// Bad - capitalized comment
func process() {}
```

### Formatting

**NEVER format code manually.** Never use gofmt or any formatting tool. Always rely on IDE formatting. Your job is to write code, not format it.

### Range Loops

Starting with Go 1.22, range loop variables no longer need to be captured. The old pattern of capturing range variables is now obsolete:

```go
// Old pattern (pre-Go 1.22) - DO NOT USE
for _, extCID := range TestExtCIDs {
    extCID := extCID // capture range variable - NO LONGER NEEDED
    // use extCID...
}

// Modern pattern (Go 1.22+) - CORRECT
for _, extCID := range TestExtCIDs {
    // use extCID directly - variable is properly scoped per iteration
}
```

This change eliminates the common gotcha where loop variables were shared across iterations when used in goroutines or closures.

### String Operations

Always use `fmt` for string operations - NEVER use the `+` operator for concatenation:

```go
// Bad
result := "hello" + " " + "world"

// Good
result := fmt.Sprintf("%s %s", "hello", "world")
```

### Logging

Use zerolog with structured logging:

```go
// Debug for development flow tracing
log.Debug().Msgf("processing user: %s", userID)

// Info for critical operations (game/money flow)
log.Info().Ctx(ctx).
    Str("game_session_id", gsID).
    Int64("bet_cents", amount).
    Msgf("bet placed for user %s", userID)

// Error with context for tracing
log.Error().Ctx(ctx).Err(err).Msgf("failed to process: %s", id)
```

**Critical rules:**
- Use `.Ctx(ctx)` for trace ID injection (distributed tracing)
- Log game/money operations at `Info` level with all relevant IDs
- **DON'T log and return errors** - middleware handles error logging
- Use `Debug` for flow tracing (omitted in production)

### API Responses

API response rules:

```go
// Times - always UTC in RFC3339 format
response := UserResponse{
    CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
    UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
}

// Money - always string in main currency unit
response := BalanceResponse{
    Amount:   "10.50",  // NOT 1050 cents
    Currency: "USD",
}
```

### Naming Conventions

File and package naming:

```go
// File names - omit folder name
controller/v1/user.go       // NOT user-controller.go
service/user.go             // NOT user-service.go

// Package names - short version of path
package controllerv1        // From controller/v1
package usecasev1           // From usecase/v1

// Import aliases - numbered suffix for common lib
import (
    domain2 "github.com/swipegames/platform-lib-common/domain"  // Common lib
    domain "myservice/internal/domain"                          // Service
)
```

### Error Handling

Use `domain.Error` for structured errors:

```go
// Create errors with status-specific constructors
domain.NewBadRequestError("Invalid input", err)
domain.NewNotFoundError("User not found", err)
domain.NewForbiddenError("Access denied", err)

// Add error codes for client handling
domain.NewNotFoundError("Game unavailable", err).
    WithCode(domain.ErrorCodeGameNotFound)
```

**Layer-specific rules:**

```go
// Repository - specify root cause
return nil, fmt.Errorf("failed to query user: %w", dbErr)

// Service - wrap with context
return nil, fmt.Errorf("failed to get user: %w", err)

// Controller - NO wrapping, return as-is
return err  // Middleware logs and formats
```

**NEVER log and return:**

```go
// Bad
if err != nil {
    log.Error().Err(err).Msg("error")  // Don't log here
    return err
}

// Good
if err != nil {
    return err  // Middleware logs
}
```

### Domain Types

Use domain types for type safety:

```go
// Define domain types
type UserID uuid.UUID
type GameID string
type Email string

// Use in signatures
func GetUser(ctx context.Context, id domain.UserID) (*domain.User, error)

// Prevents mistakes like:
GetUser(ctx, gameID)  // Won't compile - type mismatch
```

### Model Separation

Maintain strict separation between three model types:

```go
// API model (OpenAPI spec) - request/response
package api
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Domain model (business logic)
package domain
type User struct {
    ID      UserID
    Email   Email
    Balance decimal.Decimal
}

// Persistence model (database)
package repository
type UserModel struct {
    ID      uuid.UUID `db:"id"`
    Email   string    `db:"email"`
    Balance int64     `db:"balance_cents"`
}
```

Never mix these models across layers.

### Domain Objects and Mapper Pattern

**Golden Rule:** Always use domain objects when passing data between layers. Convert at boundaries.

#### Controller Layer: Validate and Convert to Domain

Convert API input to domain objects in controller - this validates input and enforces type safety:

```go
// Controller receives API request, converts to domain, passes to service
func (c *Controller) CreateBet(ctx echo.Context) error {
    var req api.CreateBetRequest
    if err := ctx.Bind(&req); err != nil {
        return err
    }

    // Convert to domain in controller (validates input)
    cid, err := domain.NewCIDFromString(req.CID)
    if err != nil {
        return domain.NewBadRequestError("invalid CID", err)
    }

    bet := domain.Bet{
        CID:      cid,
        Amount:   req.Amount,
        GameID:   domain.GameID(req.GameID),
        RoundID:  domain.RoundID(req.RoundID),
    }

    // Pass ONLY domain object to service
    result, err := c.service.PlaceBet(ctx.Request().Context(), bet)
    if err != nil {
        return err
    }

    // Convert domain result to API response
    return ctx.JSON(200, domainBetResult(result).toAPI())
}
```

#### Mapper Pattern with Type Aliases

Use type aliases for clean, reusable conversions between models:

```go
// Repository layer - DB model ↔ Domain model
type (
    repositoryBet model.Bet      // Type alias for DB model
    domainBet     domain.Bet     // Type alias for domain model
)

// DB → Domain
func (r repositoryBet) toDomain() *domain.Bet {
    return &domain.Bet{
        ID:       domain.BetID(r.ID),
        CID:      domain.CID(r.Cid),
        ExtCID:   domain.ExtCID(r.ExtCid),
        Amount:   r.Amount,
        Currency: domain.Currency(r.Currency),
        GameID:   domain.GameID(r.GameID),
        RoundID:  domain.RoundID(r.RoundID),
    }
}

// Domain → DB
func (d domainBet) asModel() model.Bet {
    return model.Bet{
        ID:       uuid.UUID(d.ID),
        Cid:      d.CID.UUID(),
        ExtCid:   d.ExtCID.String(),
        Amount:   d.Amount,
        Currency: d.Currency.String(),
        GameID:   d.GameID.String(),
        RoundID:  d.RoundID.String(),
    }
}

// Usage in repository
func (r *Repository) Save(ctx context.Context, bet domain.Bet) error {
    m := domainBet(bet).asModel()  // Convert domain to DB model
    return r.db.Insert(m)
}

func (r *Repository) FindByID(ctx context.Context, id domain.BetID) (*domain.Bet, error) {
    var m model.Bet
    err := r.db.Get(&m, "SELECT * FROM bet WHERE id = $1", id)
    if err != nil {
        return nil, err
    }
    return repositoryBet(m).toDomain(), nil  // Convert DB to domain
}
```

#### Controller Layer: Domain → API Response

```go
// Type alias for API response conversion
type domainBetResult domain.BetResult

func (d domainBetResult) toAPI() api.BetResponse {
    return api.BetResponse{
        BetID:    d.BetID.String(),
        Balance:  fmt.Sprintf("%.2f", d.Balance.InexactFloat64()),
        Currency: d.Currency.String(),
        Status:   string(d.Status),
    }
}

// Type alias for collection conversion
type domainBetSlice []domain.Bet

func (d domainBetSlice) toAPI() []api.BetInfo {
    result := make([]api.BetInfo, len(d))
    for i, bet := range d {
        result[i] = domainBet(bet).toAPI()
    }
    return result
}

func (d domainBet) toAPI() api.BetInfo {
    return api.BetInfo{
        ID:       d.ID.String(),
        Amount:   fmt.Sprintf("%.2f", d.Amount.InexactFloat64()),
        GameID:   d.GameID.String(),
        RoundID:  d.RoundID.String(),
    }
}
```

#### Handling Optional Fields with Pointers

```go
type (
    repositoryWin model.Win
    domainWin     domain.Win
)

func (r repositoryWin) toDomain() *domain.Win {
    return &domain.Win{
        ID:      domain.WinID(r.ID),
        BetID:   domain.BetID(r.BetID),
        Amount:  r.Amount,
        // Pointer type casts for optional fields
        FRID:    (*domain.FreeRoundsID)(r.FrID),
        ExtTxID: (*domain.ExtTxID)(r.ExtTxID),
    }
}

func (d domainWin) asModel() model.Win {
    return model.Win{
        ID:     uuid.UUID(d.ID),
        BetID:  uuid.UUID(d.BetID),
        Amount: d.Amount,
        // Use helper methods for pointer conversions
        FrID:    d.FRID.StringPtr(),    // nil-safe conversion
        ExtTxID: d.ExtTxID.StringPtr(), // nil-safe conversion
    }
}
```

#### Standard Method Names

- **`toDomain()`** - Convert from any model TO domain model (DB → Domain, API → Domain)
- **`toAPI()`** - Convert from domain TO API response model (Domain → API)
- **`asModel()`** - Convert from domain TO DB persistence model (Domain → DB)
- **`fromDB()`** - Alternative to `toDomain()` when clarity needed (DB → Domain)

#### Layer-Specific Rules

**Controller Layer:**
- Receives API request models (generated from OpenAPI spec)
- Converts to domain objects immediately after binding
- Validates input during conversion (use domain constructors)
- Passes ONLY domain objects to service/usecase
- Converts domain results to API responses using `toAPI()`

**Service/Usecase Layer:**
- Accepts ONLY domain objects as parameters
- Returns ONLY domain objects
- No knowledge of API or DB models
- Pure business logic with domain types

**Repository Layer:**
- Accepts ONLY domain objects as parameters
- Returns ONLY domain objects
- Converts domain to DB models using `asModel()` before persistence
- Converts DB models to domain using `toDomain()` after retrieval
- No knowledge of API models

#### Benefits

1. **Type Safety:** Domain constructors validate input at boundaries
2. **Clean Separation:** Each layer works only with appropriate models
3. **Reusable Conversions:** Type aliases + methods = DRY conversions
4. **Testability:** Easy to test conversions independently
5. **Maintainability:** Changes to one model type don't ripple through layers

### General Code Philosophy

- Don't provide extensive explanations when changing code - just change it
- Don't describe what you changed - diffs show that
- Minimal comments/summaries
- Don't do work that wasn't requested - no "helpful" fixes
- Keep it simple, stupid, and working with minimal requirements
- DON'T add comments everywhere - only when truly needed
- **Don't use named returns in long methods** - makes code less readable

## Testing Standards

### Before Writing Tests

1. **Check existing tests first** - understand patterns used in the project
2. **Follow the same pattern** - don't invent new test styles
3. **NEVER touch the code when writing tests** - if code changes are needed, ask first

### Test Framework

Use `testify/suite` for integration tests:

```go
import tests2 "github.com/swipegames/platform-lib-common/tests"

type UserTestSuite struct {
    tests2.IntegrationTestSuite  // Extends base suite
    // Add service-specific fields
}

func TestUserSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}
```

### Test Data

- **Use fixtures from `db/fixtures/`** - managed by Goose
- Don't create new test data unless absolutely necessary
- No magic strings or numbers where you arrange then assert
- Use constants instead
- If constants are used in current test only, use lowercase (package private)
- If used across tests, use uppercase

```go
// Good - constants for test values
func TestCreateUser(t *testing.T) {
    const (
        testEmail = "test@example.com"
        testName  = "Test User"
    )

    user := CreateUser(testName, testEmail)

    if user.Email != testEmail {
        t.Errorf("expected %s, got %s", testEmail, user.Email)
    }
}

// Exception - one-time setup values can be inline
func TestConfig(t *testing.T) {
    cfg := Config{Port: 8080} // OK - used once for setup
    // ...
}
```

### Table Tests

- Don't use underscores in test names - use spaces
- Name should describe what the test does (arrange-act-assert style)
- **Don't use description field** - just use name

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string  // Use descriptive names with spaces
        a, b    float64
        want    float64
        wantErr bool
    }{
        {name: "divides two positive numbers", a: 10, b: 2, want: 5, wantErr: false},
        {name: "returns error when dividing by zero", a: 10, b: 0, want: 0, wantErr: true},
        {name: "divides negative numbers", a: -10, b: 2, want: -5, wantErr: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Errorf("Divide() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Divide() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Coverage

Always cover:
- Positive cases
- Negative cases
- Edge cases (empty values, nil values, max/min values)
- Database state after operations (if applicable)

### Running Tests

**Always use VSCode MCP tasks to run tests** - never use command line.

When tests fail:
1. Check the logs
2. If not enough logs, add more logs
3. Don't guess what's going on

Once you change an existing test, run it to ensure it still works.

### Test Code Changes

If you've worked on an issue for more than 5 minutes, don't revert changes - the time and cost is already spent.

## References

Read references in this order for best understanding:

1. **actual-patterns.md** - Real production code examples showing error handling, database queries, testing, concurrency, and logging patterns. **START HERE** to see how patterns are actually implemented.

2. **project-structure.md** - Service organization patterns (monorepo/multi-repo), layered architecture (controller → usecase → service → repository), directory structure, database organization (migrations, fixtures, partitioning), module system with import aliases. Complete project organization guide.

3. **best-practices.md** - Core Go idioms, error handling, naming conventions, interfaces, struct composition, domain types for type safety, API development standards (money format, time format, versioning), common patterns.

4. **testing.md** - Comprehensive testing patterns including table-driven tests, subtests, test helpers, mocking, benchmarking, integration tests with testify/suite.

5. **concurrency.md** - Goroutines, channels, select statements, worker pools, synchronization primitives, context usage, error handling in concurrent code.

## Quick Reference

### Architecture Layers

```go
// Controller - API → Domain, call service, Domain → API
func (c *Controller) GetUser(ctx echo.Context) error {
    // Convert API input to domain
    userID, err := domain.NewUserIDFromString(ctx.Param("id"))
    if err != nil {
        return domain.NewBadRequestError("invalid user ID", err)
    }

    // Pass domain object to service
    user, err := c.service.GetUser(ctx.Request().Context(), userID)
    if err != nil {
        return err  // Return as-is, middleware logs
    }

    // Convert domain to API response
    return ctx.JSON(200, domainUser(*user).toAPI())
}

// Service - business logic with domain objects, wrap errors
func (s *Service) GetUser(ctx context.Context, id domain.UserID) (*domain.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}

// Repository - Domain → DB, query, DB → Domain
func (r *Repository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
    var m model.User
    if err := r.db.Get(&m, query, id); err != nil {
        return nil, fmt.Errorf("failed to query user: %w", err)
    }
    return repositoryUser(m).toDomain(), nil
}
```

### Domain Error Handling

```go
// Create domain errors
domain.NewNotFoundError("User not found", err).
    WithCode(domain.ErrorCodeUserNotFound)

// Check error codes
if parsedErr, parseErr := domain.ParseError(err.Error()); parseErr == nil {
    if parsedErr.Code != nil && *parsedErr.Code == domain.ErrorCodeSessionExpired {
        // Handle session expiration
    }
}
```

### Import Naming

```go
import (
    domain2 "github.com/org/lib-common/domain"
    config2 "github.com/org/lib-common/config"
    domain "github.com/org/service/internal/domain"
    config "github.com/org/service/internal/config"
)
```

### Logging with Context

```go
// Critical operations with trace IDs
log.Info().Ctx(ctx).
    Str("session_id", sessionID).
    Int64("amount", amount).
    Msgf("operation completed for user %s", userID)

// Don't log and return
if err != nil {
    return err  // Middleware logs
}
```

### Domain Types

```go
type UserID uuid.UUID
type ItemID string

func GetUser(ctx context.Context, id domain.UserID) (*domain.User, error)
```

### Mapper Pattern

```go
// Type aliases for conversions
type (
    repositoryUser model.User
    domainUser     domain.User
)

// DB → Domain
func (r repositoryUser) toDomain() *domain.User {
    return &domain.User{
        ID:    domain.UserID(r.ID),
        Email: domain.Email(r.Email),
    }
}

// Domain → DB
func (d domainUser) asModel() model.User {
    return model.User{
        ID:    uuid.UUID(d.ID),
        Email: d.Email.String(),
    }
}

// Domain → API
type domainUserResponse domain.User

func (d domainUserResponse) toAPI() api.UserResponse {
    return api.UserResponse{
        ID:    d.ID.String(),
        Email: d.Email.String(),
    }
}

// Usage in layers
// Controller: API → Domain → Service → Repository
bet := domain.Bet{CID: cid, Amount: req.Amount}
result, err := c.service.PlaceBet(ctx, bet)
return ctx.JSON(200, domainBetResult(result).toAPI())

// Repository: Domain → DB, DB → Domain
m := domainUser(user).asModel()
return repositoryUser(m).toDomain()
```

For detailed patterns and examples, see the reference documentation files.
