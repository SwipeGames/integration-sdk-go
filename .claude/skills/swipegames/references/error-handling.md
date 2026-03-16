# Error Handling with domain.Error

## Overview
The `domain.Error` type provides standardized error handling for communication between services via HTTP and gRPC, preserving error context across service boundaries.

## Error Structure

```go
type Error struct {
    HttpStatus int           // HTTP status code (e.g., 400, 404, 500)
    Code       *ErrorCode    // Optional domain-specific error code
    Message    string        // Public error message (safe to expose to clients)
    Err        error         // Internal error details (NOT exposed to clients)
    Action     *ClientAction // Optional client action (refresh, redirect)
}
```

## Serialization Format

Domain errors serialize to structured string format for transmission.

**Pattern**: `[status|code]message(details)|action|data`

**Examples**:
- `[404]Session not found`
- `[500|game_not_found]Game not available(database connection failed)`
- `[401]Session expired|refresh`
- `[403|account_blocked]Account restricted(compliance check failed)|redirect|/blocked`

## Service Communication Flow

### HTTP Communication

```go
// Service A creates error
err := domain.NewNotFoundError("Session not found", dbErr).
    WithCode(domain.ErrorCodeSessionNotFound)

// Error converts to HTTP response with status code and JSON body
```

### gRPC Communication

```go
// Service A returns domain.Error
return nil, domain.NewError("Operation failed", err).
    WithCode(domain.ErrorCodeGameNotFound)

// Error serializes to string format and returns as gRPC error
```

### Receiving gRPC Errors

```go
// Service B receives gRPC error from Service A
err := serviceAClient.DoSomething(ctx, req)

// Parse error string to reconstruct domain.Error
if parsedErr, parseErr := domain.ParseError(err.Error()); parseErr == nil {
    // Access original HttpStatus, Code, Message, Action
    if parsedErr.Code != nil && *parsedErr.Code == domain.ErrorCodeSessionExpired {
        // Handle session expiration
    }
}

// Or use errors.As
var domainErr *domain.Error
if errors.As(err, &domainErr) {
    // Handle domain error
}
```

### Proxying Errors (gRPC Server)

```go
// Service B calls Service C and receives error
err := serviceCClient.DoSomething(ctx, req)

// Check if error is already a domain.Error (from downstream service)
if parsedErr, parseErr := domain.ParseError(err.Error()); parseErr == nil {
    // Pass error as-is (proxy mode)
    return nil, parsedErr
}

// Otherwise, wrap as new domain.Error
return nil, domain.NewError("Operation failed", err)
```

## Creating Errors

### Basic Error
```go
domain.NewError("Internal error", err)
```

### HTTP Status-Specific Constructors
```go
domain.NewBadRequestError("Invalid input", err)
domain.NewUnauthorizedError("Authentication required", err)
domain.NewForbiddenError("Access denied", err)
domain.NewNotFoundError("Resource not found", err)
domain.NewAlreadyExistsError("Duplicate entry", err)
```

### With Error Code
```go
domain.NewNotFoundError("Game unavailable", err).
    WithCode(domain.ErrorCodeGameNotFound)
```

### With Client Action
```go
err := domain.Error{
    HttpStatus: 401,
    Code:       &domain.ErrorCodeSessionExpired,
    Message:    "Session expired",
    Action:     &domain.ClientAction{Type: domain.ClientActionTypeRefresh},
}
```

## Standard Error Codes

Predefined error codes ensure consistency:

- `game_not_found` - Requested game doesn't exist
- `currency_not_supported` - Currency not supported by game/platform
- `locale_not_supported` - Locale not supported
- `account_blocked` - User account is blocked
- `bet_limit` / `loss_limit` / `time_limit` - Responsible gaming limits
- `insufficient_funds` - Not enough balance
- `session_expired` / `session_not_found` - Session issues
- `client_connection_error` - Client connectivity problems

## Client Actions

Direct client behavior on error:

- `ClientActionTypeRefresh` - Client should refresh session/token
- `ClientActionTypeRedirect` - Client should redirect (URL in Data field)

## Best Practices

### Error Wrapping
- Use error wrapping to pass errors through layers
- Specify root error cause on lower layer
- Top layer (controller) returns errors as-is (no wrapping)

### Logging
- Don't log errors when returning them
- Errors logged by middleware layer
- Use `.Ctx(ctx)` for distributed tracing

### gRPC Error Proxying
When proxifying errors from integration service to client (e.g., core → integration → core → client):
- Core service must not attach any information to error
- Client should understand what happened on integration service side
- Pass error as-is to preserve original context
