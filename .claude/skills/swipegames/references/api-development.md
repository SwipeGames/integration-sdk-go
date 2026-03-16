# API Development Guide

## API Versioning

### Semantic Versioning
Use semver for API versions: `v1.0.0`

### Backward Compatibility
- Minor versions always backward-compatible
- Server updates handler to support new version
- Client uses same code with new version (new changes ignored)
- Achieve through optional fields in requests and responses

## API Standards

### Time Formatting
- Always UTC timezone
- RFC3339 format: `2024-01-30T15:04:05Z`

### Money Formatting
- Return and receive as string
- Main currency unit: `10.50` for $10.50
- Never use floating point for money

### Request Validation
- All validation in controller/handler layer
- Before passing to service/usecase layer
- Return clear validation errors

## OpenAPI Specifications

### operationId Naming Convention

Format: `{httpMethod}{PathWithoutSlashes}` in camelCase

**Rules**:
- Use HTTP method + path without slashes
- Don't add extra terms not in actual path
- Keep consistent across all YAML files

**Examples**:
```yaml
GET /level → operationId: getLevel
POST /config/rotate → operationId: postConfigRotate
GET /userstats → operationId: getUserstats
POST /bet → operationId: postBet
GET /game/session → operationId: getGameSession
```

### Generated Files
- Don't review generated files (`*.gen.go`) in code review
- Regenerate after OpenAPI spec changes

## API Types

### Public API
- **Location**: `public-api` repository
- **Used by**: External partners for integration
- **Security**: HMAC signature check
- **Access**: Cannot be used from frontend (requires secret key)
- **Design**: OpenAPI specification

### Internal API
- **Location**: `internal-api` repository
- **Used by**: Frontend applications, internal clients
- **Security**: Game session ID
- **Access**: Publicly available to Internet
- **Design**: OpenAPI specification

### Private API (gRPC)
- **Location**: `platform/shared` repository
- **Used by**: Service-to-service communication
- **Security**: Internal network only
- **Access**: Not publicly available
- **Design**: gRPC/Protocol Buffers specification

### Private HTTP API
- **Port**: Always 8888
- **Used by**: Cron jobs, internal operations
- **Access**: Within cluster only
- **Design**: Directly in service code

## Code Generation

### Generate API Client
After updating OpenAPI spec:

**For TypeScript/Frontend**:
```bash
npm run generate  # Uses Orval
```

**For Go/Backend**:
```bash
make gen-api <service_name>  # Uses oapi-codegen
```

### Generate Database Models
After schema changes:
```bash
make gen-db <service_name>  # Uses go-jet
```

## Response Patterns

### Success Response
```json
{
  "data": { ... },
  "timestamp": "2024-01-30T15:04:05Z"
}
```

### Error Response
```json
{
  "error": {
    "code": "game_not_found",
    "message": "Game unavailable",
    "action": "refresh"
  }
}
```

### Balance Updates
- Always return current balance after operation
- Use string format for amounts
- Include currency code

## Authentication Headers

### HMAC Authentication (Public API)
```http
X-SG-Client-ID: <provider_identifier>
X-SG-Client-TS: <unix_timestamp>
X-SG-Client-Signature: <hmac_sha256_signature>
```

### Session Authentication (Internal API)
```http
Authorization: Bearer <session_token>
```

## Best Practices

### API Design
- RESTful endpoints where appropriate
- Clear, descriptive endpoint names
- Consistent response structures
- Proper HTTP status codes

### Versioning Strategy
- Start with v1
- Increment minor version for backward-compatible changes
- Increment major version for breaking changes
- Support multiple versions simultaneously

### Documentation
- Keep OpenAPI specs up to date
- Document all fields with descriptions
- Provide example requests/responses
- Include error scenarios

### Testing
- Test all API endpoints
- Test authentication/authorization
- Test error handling
- Test backward compatibility

### Error Handling
- Use domain.Error for consistency
- Provide clear error messages
- Include error codes for client handling
- Never expose internal errors to clients
