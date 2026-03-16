# Integration Adapters

## Overview
Integration adapters connect the gaming platform with external casino operators and payment providers.

## Architecture
Use 3-layer architecture for all integrations:
- **Controller**: Request handling and validation
- **Service**: Business logic and orchestration
- **Repository**: Database operations

Every integration should have mock service for testing.

## Integration Types

### Direct Integration
- Public API integration
- No custom integration adapter needed
- Partners call our Public API directly

### Reverse Integration
- Custom integration adapter required
- Called by integration/core service
- Adapter translates to partner's API

## Routing
Requests routed from core to specific integration via Istio based on CID.
Every integration needs related infrastructure settings.

## Integration Client Pattern (integration.go)

Every integration adapter has an `integration.go` file that acts as the client for making requests to the external integration.

### Critical Rule: Use Raw JSON Requests

**DO NOT use generated API clients.** Instead, construct raw JSON requests and sign them manually.

### Implementation Pattern

```go
// 1. Get API key/token from secret service
token, err := s.secretService.Get(s.intCfg.APIKeyNameForCID(d.CID))
if err != nil {
    return -1, fmt.Errorf("failed to get API key for CID: %s :%w", d.CID, err)
}

// 2. Construct raw JSON request using fmt.Sprintf
req := fmt.Sprintf(`{"account_id":"%s","currency":"%s","game_id":"%s","session_id":"%s"}`,
    d.ExtCID, d.Currency, d.GameID, d.ExtSessionID)

// 3. Sign the request
rSigner := utils.NewRequestSigner(token)
sign, err := rSigner.Sign(req)
if err != nil {
    return -1, fmt.Errorf("failed to sign request: %w", err)
}

// 4. Make HTTP request with signature
var sucResp apiv2.PlayerBalanceResponse
var errResp apiv2.ErrorResponse
client := service3.NewRestyClient(s.intCfg)
client.SetRetryCount(3)

resp, err := client.R().
    SetContext(ctx).
    SetHeader("X-REQUEST-SIGN", sign).
    SetBody(req).
    SetResult(&sucResp).
    SetError(&errResp).
    Post(tUrl)
```

### Why Raw JSON?

1. **Full control** over request format and signing
2. **Simpler debugging** - see exact JSON being sent
3. **Easier request signing** - sign raw payload before HTTP call
4. **No code generation dependency** - reduces complexity
5. **Flexible** - easy to adjust request format per integration needs

### Key Components

- **Secret Service**: Manages API keys per CID
- **Request Signer**: Signs requests using HMAC or other algorithms
- **Resty Client**: HTTP client with retry logic
- **SetHeader("X-REQUEST-SIGN", sign)**: Adds signature to request headers

## DB-First Transaction Pattern

**Critical Pattern**: Save to database FIRST, then make external API calls.

### Philosophy
Prevent money loss scenarios by persisting transactions before external calls.

### Implementation Flow

1. **Validate request** - Check all parameters before any operations
2. **Save to DB with pending status** - Status: `pending_api_call`
3. **Make external API call** - Call partner's API
4. **Update transaction with API response** - Save success/error
5. **Commit transaction** - Only after successful DB update

### Error Handling

**Refundable errors** (5xx, timeouts):
- Commit transaction with error
- Background service will retry or refund

**Non-refundable errors** (4xx):
- Rollback entire transaction
- No DB record created

### Background Processing
Failed transactions handled by dedicated processing services:
- Retry successful API calls that failed to update DB
- Refund transactions that cannot be completed

## Resiliency and Retries

When bet/win/refund performed:
1. Validate request
2. Save with `pending_api_call` status
3. Call integration API
4. If fails or service crashes, worker refunds or retries later

**DB-first approach prevents**:
- Money loss if service crashes between API call and DB update
- Lost transactions when API succeeds but DB update fails

## BetWin Operations

Unified `BetWin` call reduces API calls to integration and improves speed.

### Logic

**If bet part fails**:
- Return error to caller
- Round marked as failed
- Bet stored in DB for management (refund if needed)
- Win part not processed or saved

**If win part fails**:
- Don't return error to caller
- Win will be retried in future
- Round considered successful
- Allows caller to determine round success

### Unified BetWin Struct
Combines bet and win for atomic operations:
```go
type BetWin struct {
    BetAmount  decimal.Decimal
    WinAmount  decimal.Decimal
    RoundID    string
    // ... other fields
}
```

## Idempotency

Win and BetWin operations support idempotency via optional `ext_tx_id` field.

### Implementation

**Duplicate Detection**:
- If win with same `ext_tx_id` exists for CID, return existing balance
- No new transaction processed

**Database Constraint**:
- Unique partial index: `(cid, ext_tx_id, created_at)` where `ext_tx_id IS NOT NULL`

**Outbox Pattern**:
- Idempotency check returns existing results
- No transaction state validation needed
- Outbox pattern handles eventual consistency

**Optional Field**:
- Only enforced when `ext_tx_id` provided
- Wins without field processed normally

Allows integrations to safely retry win/betwin requests without duplicate payouts.

## Transaction and Round IDs

### Transaction ID
- Unique per CID
- Different CIDs can have same transaction IDs
- When providing transaction info, always include CID

### Round ID
- Unique per Game ID
- Different Game IDs can have same Round IDs
- Different CIDs can have same Round ID

### Free Rounds Constraints
- Free Round ID unique per Round ID
- Since Round ID unique per Game ID, Free Round ID unique per Game ID
- `fr_id` in database is CID-related (external free rounds ID)
- Free Round ID unique per CID

## Refunds

- Only for bets (not wins)
- Refund whole bet amount (no partial refunds)
- No external refund requests accepted
- Internal integration-related functionality only

## Error Cases and Sentinel Values

When user balance cannot be determined:
- Return balance as `-1`
- Sentinel value means balance unknown
- Client should not update user balance

## HMAC Authentication

All casino API calls use HMAC authentication headers:

```
X-SG-Client-ID: <provider_identifier>
X-SG-Client-TS: <unix_timestamp>
X-SG-Client-Signature: <hmac_sha256_signature>
```

## Temporal Workflows

Casino API operations orchestrated via Temporal workflows:
- Automatic retry with exponential backoff
- Compensation logic for failures
- Activity timeouts and error handling
- One worker per service pattern

### Temporal Worker Setup

```go
// Setup during app configuration
temporalWorker, err := app.TemporalSetup(cfg.Temporal, "service-task-queue")
app.RegisterTemporalWorker(temporalWorker)

// Register workflows and activities
temporalWorker.RegisterWorkflowsAndActivities(betUseCase, winUseCase)
```

## Testing

### Integration Testing
- Use WireMock for external API simulation
- Test both success and failure scenarios
- Verify DB-first transaction pattern
- Test idempotency behavior
- Verify background processing

### Mock Services
- Every integration needs mock service
- Used in integration tests
- Simulates partner API behavior
