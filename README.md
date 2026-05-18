# VideoForge Backend

A performance-based video marketplace backend built with Go microservices.

## Architecture

The system consists of 11 independently deployable microservices, each owning its own PostgreSQL schema and communicating via NATS + internal REST.

```
┌──────────────┐
│ API Gateway  │ ← Public entry point (JWT, rate limit, routing)
└──────┬───────┘
       │
   ┌───┴───┬─────────┬──────────┬──────────┐
   ▼       ▼         ▼          ▼          ▼
┌──────┐┌──────┐ ┌────────┐ ┌────────┐ ┌────────┐
│ User ││Brief │ │ Video  ││Campaign││Shopify │
└──────┘└──────┘ └────────┘ └────────┘ └────────┘
   ▼       ▼         ▼          ▼          ▼
┌──────────┐ ┌────────────┐ ┌────────────┐
│Performance│ │  Payout    ││Notification│
└──────────┘ └────────────┘ └────────────┘
   ▼
┌─────────┐ ┌────────────┐
│  Admin  │ │ AI Support │
└─────────┘ └────────────┘
```

## Services

| Service | Path | Description |
|---------|------|-------------|
| API Gateway | `svc-gateway/` | Public entry point, JWT validation, rate limiting, routing |
| User Service | `svc-user/` | Auth (register/login/refresh), roles, profiles |
| Brief Service | `svc-brief/` | AI interview stub, brief CRUD, matching, submission limits |
| Video Service | `svc-video/` | Upload stubs, approval workflow, versioning, blind submissions |
| Campaign Service | `svc-campaign/` | Campaign creation, budget tracking, ad account placeholders |
| Shopify Service | `svc-shopify/` | Custom links, webhook ingestion, order attribution |
| Performance Service | `svc-performance/` | Sales aggregation, leaderboards, time-series analytics |
| Payout Service | `svc-payout/` | Earnings calculation (free tier + 5% fee), Ruul.io bulk payouts |
| Notification Service | `svc-notification/` | WebSocket real-time updates, event consumption |
| Admin Service | `svc-admin/` | User management, disputes, moderation, payout overrides |
| AI Support Service | `svc-ai-support/` | Chat agent, escalation to human admins |

## Tech Stack

- **Language**: Go 1.23+
- **Database**: PostgreSQL (per-service schemas)
- **Message Queue**: NATS
- **API Style**: REST + WebSockets
- **Auth**: JWT (RS256)
- **Errors**: RFC 7807 Problem Details
- **Hosting**: Fly.io (infra configs included)
- **Payments**: DodoPayments (client→platform), Ruul.io (platform→freelancers)

## Quick Start

### Prerequisites

- Go 1.23+
- Docker + Docker Compose
- Make

### Local Development

```bash
# Start dependencies
docker-compose up -d

# Run a service
cd svc-user && go run ./cmd/main.go

# Or run all (in separate terminals)
cd svc-gateway && go run ./cmd/main.go
cd svc-user && go run ./cmd/main.go
cd svc-brief && go run ./cmd/main.go
# ... etc
```

### Environment Variables

Each service reads from environment variables. Copy `.env.example` (create your own) or set:

```bash
PORT=8080
DATABASE_URL=postgres://user:pass@localhost/dbname
NATS_URL=nats://localhost:4222
JWT_SECRET=your-jwt-secret
ENVIRONMENT=development
```

### Database Migrations

Each service has its own migrations in `migrations/`:

```bash
cd svc-user
goose postgres "$DATABASE_URL" up
```

## API Conventions

- Base path: `/api/v1`
- Content-Type: `application/json`
- Errors: RFC 7807 `application/problem+json`
- Pagination: `?page=1&limit=20`
- Sorting: `?sort=created_at:desc`
- Filtering: `?status=approved&role=editor`
- IDs: UUID v7 (time-sortable)
- Timestamps: RFC 3339

## Project Structure

```
.
├── pkg/                          # Shared packages
│   ├── config/                   # Environment config loading
│   ├── logger/                   # Structured logging (slog)
│   ├── errors/                   # RFC 7807 Problem Details
│   ├── middleware/               # HTTP middleware + JWT auth
│   ├── natsclient/               # NATS connection manager
│   └── database/                 # PostgreSQL pool setup
├── svc-gateway/                  # API Gateway
├── svc-user/                     # User Service
├── svc-brief/                    # Brief Service
├── svc-video/                    # Video Service
├── svc-campaign/                 # Campaign Service
├── svc-shopify/                  # Shopify Service
├── svc-performance/              # Performance Service
├── svc-payout/                   # Payout Service
├── svc-notification/             # Notification Service
├── svc-admin/                    # Admin Service
├── svc-ai-support/               # AI Support Service
├── docker-compose.yml            # Local dependencies
├── Makefile                      # Build commands
└── go.work                       # Go workspace
```

## Core Data Flow

```
1. CLIENT uploads raw footage + website URL
        ↓
2. AI BRIEF SERVICE interviews client → structured brief
        ↓
3. CLIENT deposits freelancer bounty budget via DodoPayments
        ↓
4. BRIEF published → editors notified
        ↓
5. EDITORS submit videos (blind, up to limit)
        ↓
6. CLIENT reviews → approves N videos
        ↓
7. AD SPECIALIST creates campaign with approved videos
        ↓
8. SHOPIFY SERVICE generates custom link per video
        ↓
9. Campaign runs (client pays Meta/TikTok directly)
        ↓
10. SHOPIFY webhook → order attributed to video
        ↓
11. PERFORMANCE SERVICE aggregates sales, updates leaderboards
        ↓
12. PAYOUT SERVICE calculates earnings ($500 free, then 5% fee)
        ↓
13. Platform pays freelancers via Ruul.io bulk payout
        ↓
14. NOTIFICATION SERVICE pushes real-time updates
```

## Testing

```bash
# Run all tests
make test

# Run tests for a specific service
cd svc-user && go test ./...

# Integration tests (requires docker-compose up)
make test-integration
```

## Deployment

Each service is containerized and deployable to Fly.io:

```bash
# Deploy a service
cd svc-user && fly deploy

# Or deploy all
make deploy-all
```

See individual service `README.md` files for service-specific details.

## License

MIT
