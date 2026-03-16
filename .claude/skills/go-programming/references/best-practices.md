# Go Best Practices and Idioms

## Table of Contents
- Error Handling
- Naming Conventions
- Interface Design
- Struct Composition
- Package Organization
- Common Patterns
- Domain Objects and Mapper Pattern

## Error Handling

### Return Errors, Don't Panic
```go
// Good
func readFile(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading %s: %w", path, err)
    }
    return data, nil
}

// Bad - panic for recoverable errors
func readFile(path string) []byte {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err) // Don't do this
    }
    return data
}
```

### Error Wrapping with %w
Use `%w` to wrap errors for error chain inspection:
```go
if err := doSomething(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Later, check for specific errors
if errors.Is(err, os.ErrNotExist) {
    // Handle not found
}
```

### Custom Error Types
```go
type ValidationError struct {
    Field string
    Value interface{}
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Msg)
}
```

### Sentinel Errors
```go
var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid input")
)

// Use errors.Is() to check
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

## Naming Conventions

### Package Names
- Short, lowercase, single word
- No underscores or mixedCaps
- Import path's last element should match package name
```go
// Good
package user
package http

// Bad
package user_service
package httpUtils
```

### Variable Names
- Short names for short lifetimes
- More descriptive for wider scope
```go
// Good - short scope
for i, v := range items {
    fmt.Println(i, v)
}

// Good - wider scope
var userRepository *Repository
var maxConnectionRetries = 3

// Bad - overly verbose for short scope
for indexNumber, itemValue := range items {
    fmt.Println(indexNumber, itemValue)
}
```

### Interface Names
- Single-method interfaces end in "-er"
- Descriptive names for multi-method interfaces
```go
// Good
type Reader interface {
    Read(p []byte) (n int, err error)
}

type UserRepository interface {
    FindByID(id string) (*User, error)
    Save(user *User) error
}

// Bad
type IUserRepository interface { // Don't use "I" prefix
    // ...
}
```

### Receiver Names
- Use short, consistent names (1-2 letters)
- Same name for all methods on a type
```go
type Client struct {
    name string
}

// Good - consistent "c" receiver
func (c *Client) Name() string {
    return c.name
}

func (c *Client) SetName(name string) {
    c.name = name
}

// Bad - inconsistent receivers
func (client *Client) Name() string { /* ... */ }
func (c *Client) SetName(name string) { /* ... */ }
```

## Interface Design

### Accept Interfaces, Return Structs
```go
// Good - accept interface
func ProcessData(r io.Reader) error {
    // Implementation
}

// Good - return concrete type
func NewClient(url string) *Client {
    return &Client{url: url}
}

// Bad - return interface for no reason
func NewClient(url string) ClientInterface {
    return &Client{url: url}
}
```

### Small Interfaces
Prefer many small interfaces over large ones:
```go
// Good - composable
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type ReadWriter interface {
    Reader
    Writer
}

// Bad - monolithic interface
type Storage interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    Seek(offset int64, whence int) (int64, error)
    Stat() (FileInfo, error)
    // Too many methods
}
```

### Empty Interface Usage
Avoid `interface{}` (or `any`) when possible:
```go
// Bad - loses type safety
func Process(data interface{}) error {
    // Need type assertions everywhere
}

// Good - use generics or specific types
func Process[T any](data T) error {
    // Type safe
}

// Or when specific type is known
func Process(data *User) error {
    // Type safe
}
```

## Struct Composition

### Embed for Behavior, Not State
```go
// Good - embedding for interface satisfaction
type LoggingWriter struct {
    io.Writer
    logger *log.Logger
}

func (lw *LoggingWriter) Write(p []byte) (n int, err error) {
    lw.logger.Printf("Writing %d bytes", len(p))
    return lw.Writer.Write(p)
}

// Bad - embedding for state sharing
type User struct {
    Database // Don't embed for state
    Name string
}
```

### Use Struct Tags
```go
type User struct {
    ID        string    `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

### Zero Values
Design structs to have useful zero values:
```go
// Good - works without initialization
type Buffer struct {
    buf []byte
}

func (b *Buffer) Write(p []byte) (n int, err error) {
    b.buf = append(b.buf, p...)
    return len(p), nil
}

// Usage
var buf Buffer
buf.Write([]byte("hello")) // Works without initialization
```

## Package Organization

### Internal Packages
Use `internal/` for private packages:
```
myproject/
├── cmd/
│   └── server/
│       └── main.go
├── internal/          # Only importable by this module
│   ├── auth/
│   ├── database/
│   └── handlers/
├── pkg/               # Public packages
│   └── client/
└── go.mod
```

### Package-Level Variables
Initialize in `init()` or use `sync.Once`:
```go
// Good - sync.Once for lazy initialization
var (
    instance *Database
    once     sync.Once
)

func GetDatabase() *Database {
    once.Do(func() {
        instance = &Database{/* ... */}
    })
    return instance
}

// Avoid - init() for heavy operations
func init() {
    // Runs at program start, slows down imports
    db = connectToDatabase()
}
```

## Common Patterns

### Functional Options
```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type Option func(*Server)

func WithHost(host string) Option {
    return func(s *Server) {
        s.host = host
    }
}

func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

func NewServer(opts ...Option) *Server {
    s := &Server{
        host:    "localhost",
        port:    8080,
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage
server := NewServer(
    WithHost("0.0.0.0"),
    WithPort(9000),
)
```

### Context Usage
Always pass context as first parameter:
```go
// Good
func FetchUser(ctx context.Context, id string) (*User, error) {
    // Check context
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Implementation
}

// Bad - context not first parameter
func FetchUser(id string, ctx context.Context) (*User, error) {
    // Wrong parameter order
}
```

### Defer for Cleanup
```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // Always closes, even on error

    // Process file
    return nil
}
```

### Table-Driven Tests Preview
For detailed testing patterns, see `testing.md`. Quick example:
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"zero", 0, 5, 5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

## Domain Types (Team Pattern)

### Type Safety with Domain Types

Create strong domain types for all identifiers and values to prevent type-confusion bugs:

```go
package domain

// Domain IDs
type UserID uuid.UUID
type GameID string
type SessionID uuid.UUID
type RoundID uuid.UUID
type CurrencyID string

// Domain values
type Email string
type Currency string
type Locale string

// Validation on domain types
func (e Email) Validate() error {
    if !strings.Contains(string(e), "@") {
        return errors.New("invalid email")
    }
    return nil
}

func (c Currency) String() string {
    return string(c)
}
```

**Benefits**:
- Compile-time type safety (can't pass GameID where UserID expected)
- Self-documenting code
- Easy to add validation methods
- Refactoring-friendly

**Usage in method signatures**:

```go
// Good - type-safe, impossible to mix up parameters
func GetUser(ctx context.Context, userID domain.UserID) (*domain.User, error)
func GetGame(ctx context.Context, gameID domain.GameID) (*domain.Game, error)

// Bad - error-prone, easy to swap parameters
func GetUser(ctx context.Context, userID string) (*domain.User, error)
func GetGame(ctx context.Context, gameID string) (*domain.Game, error)
```

**Prevents mistakes**:

```go
userID := domain.UserID(uuid.New())
gameID := domain.GameID("game-123")

GetUser(ctx, gameID)  // Won't compile - type mismatch!
GetGame(ctx, userID)  // Won't compile - type mismatch!
```

## API Development Standards (Team Pattern)

### Versioning

Use semver for API versions: `v1.0.0`, `v2.0.0`

**Minor versions must be backward-compatible**:
- Server: Update handler to support new version
- Client: Same client code works (new fields optional)
- Achieved through optional request/response fields

### Time Format

All times in API responses:
- **UTC timezone** (ALWAYS)
- **RFC3339 format**: `2024-01-30T15:04:05Z`

```go
// Good
response := UserResponse{
    CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
    UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
}

// Bad - not UTC
response := UserResponse{
    CreatedAt: user.CreatedAt.Format(time.RFC3339),  // May not be UTC
}
```

### Money Format

Money always as **string** in **main currency unit** (not cents):

```go
// Good - API response
{
    "amount": "10.50",     // $10.50 (string, main unit)
    "currency": "USD"
}

// Bad - cents as integer
{
    "amount": 1050,        // Ambiguous
    "currency": "USD"
}
```

**Internal storage**: Use cents (int64) internally, convert at API boundary

```go
// Internal
type Balance struct {
    AmountCents int64
    Currency    domain.Currency
}

// API Response
type BalanceResponse struct {
    Amount   string `json:"amount"`    // "10.50"
    Currency string `json:"currency"`  // "USD"
}

// Conversion
amount := ConvertFromCurrencyCentsAsString(balance.AmountCents, balance.Currency)
response := BalanceResponse{
    Amount:   amount,
    Currency: string(balance.Currency),
}
```

### Request Validation

All validation in controller/handler layer before passing to usecase:

```go
func (h *Handler) CreateUser(ctx echo.Context) error {
    var req api.CreateUserRequest
    if err := ctx.Bind(&req); err != nil {
        return domain.NewBadRequestError("Invalid request", err)
    }

    // Validate here - before usecase
    if req.Email == "" {
        return domain.NewBadRequestError("Email required", nil)
    }

    email := domain.Email(req.Email)
    if err := email.Validate(); err != nil {
        return domain.NewBadRequestError("Invalid email format", err)
    }

    // Pass validated data to usecase
    return h.usecase.CreateUser(ctx.Request().Context(), email, req.Name)
}
```

## Domain Objects and Mapper Pattern

### Golden Rule: Domain Objects in Business Logic

**Always use domain objects when passing data between layers.** Convert at layer boundaries only.

### Controller Layer: API → Domain Conversion

Convert API input to domain objects in controller - validates input and enforces type safety:

```go
func (c *Controller) CreateBet(ctx echo.Context) error {
    var req api.CreateBetRequest
    if err := ctx.Bind(&req); err != nil {
        return domain.NewBadRequestError("invalid request", err)
    }

    // Convert and validate in controller
    cid, err := domain.NewCIDFromString(req.CID)
    if err != nil {
        return domain.NewBadRequestError("invalid CID", err)
    }

    gameID, err := domain.NewGameIDFromString(req.GameID)
    if err != nil {
        return domain.NewBadRequestError("invalid game ID", err)
    }

    // Create domain object
    bet := domain.Bet{
        CID:      cid,
        Amount:   decimal.NewFromFloat(req.Amount),
        GameID:   gameID,
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

**Key principle**: Controller validates and converts, service receives clean domain objects.

### Mapper Pattern with Type Aliases

Use type aliases for reusable, clean conversions between model types:

#### Repository Layer: DB ↔ Domain

```go
// Type aliases for mapping context
type (
    repositoryBet model.Bet      // DB model wrapper
    domainBet     domain.Bet     // Domain model wrapper
)

// DB → Domain conversion
func (r repositoryBet) toDomain() *domain.Bet {
    return &domain.Bet{
        ID:       domain.BetID(r.ID),
        CID:      domain.CID(r.Cid),
        ExtCID:   domain.ExtCID(r.ExtCid),
        Amount:   decimal.NewFromInt(r.AmountCents).Div(decimal.NewFromInt(100)),
        Currency: domain.Currency(r.Currency),
        GameID:   domain.GameID(r.GameID),
        RoundID:  domain.RoundID(r.RoundID),
        Status:   domain.BetStatus(r.Status),
        CreatedAt: r.CreatedAt,
    }
}

// Domain → DB conversion
func (d domainBet) asModel() model.Bet {
    return model.Bet{
        ID:          uuid.UUID(d.ID),
        Cid:         d.CID.UUID(),
        ExtCid:      d.ExtCID.String(),
        AmountCents: d.Amount.Mul(decimal.NewFromInt(100)).IntPart(),
        Currency:    d.Currency.String(),
        GameID:      d.GameID.String(),
        RoundID:     d.RoundID.String(),
        Status:      string(d.Status),
    }
}

// Usage in repository methods
func (r *Repository) Save(ctx context.Context, bet domain.Bet) error {
    m := domainBet(bet).asModel()  // Convert to DB model
    return r.db.Insert(ctx, m)
}

func (r *Repository) FindByID(ctx context.Context, id domain.BetID) (*domain.Bet, error) {
    var m model.Bet
    err := r.db.QueryRow(ctx, "SELECT * FROM bet WHERE id = $1", id).Scan(&m)
    if err != nil {
        return nil, fmt.Errorf("query bet: %w", err)
    }
    return repositoryBet(m).toDomain(), nil  // Convert to domain
}
```

#### Controller Layer: Domain → API

```go
// Type alias for API response conversion
type domainBetResult domain.BetResult

func (d domainBetResult) toAPI() api.BetResponse {
    return api.BetResponse{
        BetID:    d.BetID.String(),
        Balance:  fmt.Sprintf("%.2f", d.Balance.InexactFloat64()),
        Currency: d.Currency.String(),
        Status:   string(d.Status),
        Timestamp: d.CreatedAt.UTC().Format(time.RFC3339),
    }
}

// Collection conversion
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
        Status:   string(d.Status),
    }
}

// Usage in controller
func (c *Controller) GetBets(ctx echo.Context) error {
    bets, err := c.service.GetUserBets(ctx.Request().Context(), userID)
    if err != nil {
        return err
    }
    return ctx.JSON(200, domainBetSlice(bets).toAPI())
}
```

### Handling Optional Fields with Pointers

```go
type (
    repositoryWin model.Win
    domainWin     domain.Win
)

// DB → Domain (pointer casting for optional fields)
func (r repositoryWin) toDomain() *domain.Win {
    return &domain.Win{
        ID:      domain.WinID(r.ID),
        BetID:   domain.BetID(r.BetID),
        Amount:  decimal.NewFromInt(r.AmountCents).Div(decimal.NewFromInt(100)),
        // Pointer type casts for optional fields
        FRID:    (*domain.FreeRoundsID)(r.FrID),
        ExtTxID: (*domain.ExtTxID)(r.ExtTxID),
        // Optional nested struct
        ErrorInfo: r.ErrorCode.toErrorInfo(),
    }
}

// Domain → DB (nil-safe conversions)
func (d domainWin) asModel() model.Win {
    return model.Win{
        ID:          uuid.UUID(d.ID),
        BetID:       uuid.UUID(d.BetID),
        AmountCents: d.Amount.Mul(decimal.NewFromInt(100)).IntPart(),
        // Use helper methods for nil-safe pointer conversions
        FrID:    d.FRID.StringPtr(),    // Returns *string or nil
        ExtTxID: d.ExtTxID.StringPtr(), // Returns *string or nil
        ErrorCode: d.ErrorInfo.toErrorCode(),
    }
}
```

### Standard Method Naming

Consistent naming across the codebase:

- **`toDomain()`** - Convert TO domain model (DB → Domain, API → Domain)
- **`toAPI()`** - Convert TO API response (Domain → API)
- **`asModel()`** - Convert TO DB persistence model (Domain → DB)
- **`fromDB()`** - Alternative to `toDomain()` for clarity (DB → Domain)

### Layer-Specific Responsibilities

#### Controller Layer
```go
// Responsibilities:
// - Bind API request
// - Validate and convert to domain
// - Call service with domain objects
// - Convert domain response to API
func (c *Controller) CreateUser(ctx echo.Context) error {
    var req api.CreateUserRequest
    ctx.Bind(&req)

    // Validate and convert to domain
    email, err := domain.NewEmail(req.Email)
    if err != nil {
        return domain.NewBadRequestError("invalid email", err)
    }

    user := domain.User{
        Email: email,
        Name:  req.Name,
    }

    // Pass domain object to service
    created, err := c.service.CreateUser(ctx.Request().Context(), user)
    if err != nil {
        return err
    }

    // Convert domain to API
    return ctx.JSON(201, domainUser(*created).toAPI())
}
```

#### Service/Usecase Layer
```go
// Responsibilities:
// - Accept ONLY domain objects
// - Execute business logic
// - Return ONLY domain objects
// - NO knowledge of API or DB models
func (s *Service) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
    // Business logic with domain objects
    if err := s.validator.ValidateUser(user); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Call repository with domain object
    created, err := s.repo.Save(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("save user: %w", err)
    }

    // Return domain object
    return created, nil
}
```

#### Repository Layer
```go
// Responsibilities:
// - Accept ONLY domain objects
// - Convert domain → DB for persistence
// - Convert DB → domain for retrieval
// - NO knowledge of API models
func (r *Repository) Save(ctx context.Context, user domain.User) (*domain.User, error) {
    // Convert domain to DB model
    m := domainUser(user).asModel()

    // Persist DB model
    if err := r.db.Insert(ctx, m); err != nil {
        return nil, fmt.Errorf("insert user: %w", err)
    }

    // Convert DB model back to domain
    return repositoryUser(m).toDomain(), nil
}

func (r *Repository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
    var m model.User
    err := r.db.Get(ctx, &m, "SELECT * FROM users WHERE id = $1", id)
    if err != nil {
        return nil, fmt.Errorf("query user: %w", err)
    }

    // Convert DB to domain
    return repositoryUser(m).toDomain(), nil
}
```

### Benefits of This Pattern

1. **Type Safety at Boundaries**
   - Domain constructors validate input when converting from API
   - Compile-time checks prevent mixing wrong types
   - Example: Can't pass `GameID` where `UserID` expected

2. **Clean Separation of Concerns**
   - Each layer only knows about its own models + domain
   - API changes don't affect repository layer
   - DB schema changes don't affect API layer

3. **Reusable Conversions**
   - Type aliases + methods = DRY principle
   - Same conversion logic used everywhere
   - Easy to find: `toDomain()`, `toAPI()`, `asModel()`

4. **Testability**
   - Test conversions independently
   - Mock domain objects in tests
   - No need to mock API or DB models in business logic tests

5. **Maintainability**
   - Changes to one model type isolated to mapper methods
   - Clear boundaries between layers
   - Easy to trace data flow through system

### Complete Example: End-to-End Flow

```go
// 1. Controller receives API request
func (c *Controller) PlaceBet(ctx echo.Context) error {
    var req api.PlaceBetRequest
    ctx.Bind(&req)

    // Convert API → Domain
    cid, _ := domain.NewCIDFromString(req.CID)
    bet := domain.Bet{
        CID:    cid,
        Amount: decimal.NewFromFloat(req.Amount),
        GameID: domain.GameID(req.GameID),
    }

    // 2. Service with domain objects
    result, err := c.service.PlaceBet(ctx.Request().Context(), bet)
    if err != nil {
        return err
    }

    // Convert Domain → API
    return ctx.JSON(200, domainBetResult(result).toAPI())
}

// 3. Service - pure domain logic
func (s *Service) PlaceBet(ctx context.Context, bet domain.Bet) (*domain.BetResult, error) {
    // Business logic with domain objects
    if err := s.validateBet(bet); err != nil {
        return nil, err
    }

    // Repository with domain objects
    saved, err := s.repo.SaveBet(ctx, bet)
    if err != nil {
        return nil, err
    }

    balance, _ := s.ledger.Deduct(ctx, bet.CID, bet.Amount)

    return &domain.BetResult{
        BetID:   saved.ID,
        Balance: balance,
    }, nil
}

// 4. Repository - Domain ↔ DB
func (r *Repository) SaveBet(ctx context.Context, bet domain.Bet) (*domain.Bet, error) {
    // Domain → DB
    m := domainBet(bet).asModel()

    err := r.db.Insert(ctx, m)
    if err != nil {
        return nil, err
    }

    // DB → Domain
    return repositoryBet(m).toDomain(), nil
}
```

**Data flow**: API Request → Domain → Service (Domain) → Repository (Domain → DB → Domain) → Service (Domain) → Controller (Domain → API Response)
```
