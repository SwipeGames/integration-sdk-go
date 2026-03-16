# SwipeGames Platform Architecture

## Overview
Gaming platform with real-time betting experiences and casino integrations built on Go microservices architecture.

## Infrastructure
- **Cloud**: AWS with Kubernetes orchestration
- **Database**: PostgreSQL 15+ with read replicas, pg_partman for partitioning
- **Cache**: Redis 7 with clustering support
- **Workflows**: Temporal for casino API orchestration
- **Monitoring**: Prometheus + AlertManager, Grafana Loki

## Service Architecture

### Core Service (Financial Engine)
- **Location**: `platform/services/core/`
- **Responsibility**: Central coordinator for all gaming operations requiring financial integrity
- **Key Features**:
  - Atomic financial operations with rollback capability
  - Server-side RNG and outcome calculation
  - Session lifecycle management with partner configurations
  - Comprehensive audit trails and transaction logging
  - Risk management and responsible gaming enforcement

### Integration Services (Casino Gateway)
- **Location**: Various (e.g., `platform/services/swipegames-integration/`)
- **Responsibility**: Handle integration between game platform and casino operators
- **Key Features**:
  - Game launch processing via URL redirect
  - Wallet operations (/auth, /withdraw, /deposit, /rollback)
  - HMAC authentication for all casino API calls
  - Temporal workflow orchestration for casino operations
  - DB-first atomic transaction management

### Game Services
- **Catch API**: TypeScript + Fastify (frontend API)
- **Swipe API**: TypeScript + Fastify (frontend API)
- WebSocket handlers for real-time game interactions

### Internal Services
- **Location**: `platform/services/internal/`
- **Shared utilities**: HTTP/gRPC servers, database, Redis, testing frameworks

## Technology Stack

### Backend
- **Language**: Go 1.24.5+
- **Communication**: gRPC with protocol buffers
- **Database**: PostgreSQL 15+ with read replicas
- **Cache**: Redis 7 with clustering
- **Workflows**: Temporal
- **ORM**: go-jet for type-safe SQL generation
- **Migrations**: Goose
- **Logging**: zerolog with structured JSON

### Infrastructure
- **Containerization**: Docker Compose with health checks
- **Orchestration**: Kubernetes deployment via Helm charts
- **Build System**: Make-based with service-specific commands

## API Types

### Public API (Core Service Only)
- Used by external partners for integration
- Signature security check (HMAC)
- Cannot be used from frontend directly
- Designed through OpenAPI specification in `public-api` repository

### Internal API
- Secured by game session ID
- Publicly available to Internet
- Designed through OpenAPI specification in `internal-api` repository

### Private API (gRPC)
- Between-service communication
- Not publicly available
- Designed through gRPC specification in `platform/shared`

### Private HTTP API (Port 8888)
- For cron jobs and internal operations
- Not publicly available
- Always setup on port 8888

## Temporal Worker Framework

### Simplified Architecture
- **Location**: `services/internal/app/temporal.go`
- One worker per service handles all workflows/activities
- Single task queue per service
- Integrated into App lifecycle with graceful shutdown
- No singletons - clean dependency injection

### Usage Pattern
```go
// 1. Setup during app configuration
temporalWorker, err := app.TemporalSetup(cfg.Temporal, "service-task-queue")
app.RegisterTemporalWorker(temporalWorker)

// 2. Register all workflows and activities
temporalWorker.RegisterWorkflowsAndActivities(betUseCase, winUseCase)

// 3. Worker starts/stops automatically with app lifecycle
```

## Logging & Monitoring

### Logging Requirements
- **Structured Logging**: JSON format with consistent fields across all services
- **Correlation IDs**: Track requests across service boundaries
- **Log Levels**: DEBUG, INFO, WARN, ERROR with appropriate filtering
- **Sensitive Data**: Never log secrets, API keys, or financial amounts in plain text
- **Critical Operations**: Log all money-related operations with Info level

### Alerting Rules
**Critical Alerts**:
- Financial transaction failures > 1% over 5 minutes
- Database connection pool exhaustion
- Redis cluster node failures
- Partner API failure rate > 5% over 10 minutes

**Warning Alerts**:
- High transaction latency (p95 > 500ms)
- Memory usage > 80%
- CPU usage > 70% sustained over 10 minutes

## Build & Deployment

### Make Commands
```bash
make up <service_name>           # Start service with dependencies
make down                        # Stop all services
make build <service_name>        # Build service binary
make gen-db <service_name>       # Generate database models
make add-migration <service_name>  # Create new migration
make test <service_name>         # Run service tests
make test-ci <service_name>      # CI tests with Docker
make lint <service_name>         # Code linting
```

## Port Mappings
- **PostgreSQL**: 5432
- **Redis**: 6379
- **Temporal**: 7233
- **Temporal UI**: 8088
- **Service HTTP APIs**: Configured per service
- **Service Private APIs**: Always port 8888
