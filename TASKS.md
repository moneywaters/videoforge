# Implementation Plan: VideoForge Backend

## Overview
Build a performance-based video marketplace backend using granular microservices on Fly.io. 11 services total, starting with foundational infrastructure and building up to the full core loop (brief→video→campaign→sales→payout).

## Architecture Decisions
- **Start with foundational services**: Gateway + User Service first, since everything depends on auth.
- **Shared packages**: Extract common code (logging, errors, NATS client, config, middleware) into a shared module.
- **Vertical slices**: Build one service end-to-end before moving to the next.
- **Go 1.23+**: Use latest stable Go.
- **Database per service**: Each service owns its own schema.
- **NATS for async, REST for sync**: Request-reply pattern within the cluster.

## Task List

### Phase 1: Foundation — Shared Infrastructure

#### Task 1: Shared Packages
**Description:** Create common packages that every service will use.

**Acceptance criteria:**
- [ ] `pkg/config` — Environment variable loading with defaults
- [ ] `pkg/logger` — Structured logging wrapper around slog
- [ ] `pkg/errors` — RFC 7807 Problem Details error types
- [ ] `pkg/middleware` — Request ID, recover, CORS, basic auth middleware stubs
- [ ] `pkg/natsclient` — NATS connection manager with reconnect logic
- [ ] `pkg/database` — PostgreSQL connection pool setup
- [ ] `go.mod` at root with shared module path

**Verification:**
- [ ] `go build ./pkg/...` passes
- [ ] Unit tests for config and error packages

**Dependencies:** None
**Estimated scope:** Medium (3-5 files per package)

---

#### Task 2: API Gateway Scaffolding
**Description:** Public entry point service. Thin routing layer with JWT validation stubs.

**Acceptance criteria:**
- [ ] `svc-gateway/cmd/main.go` — HTTP server setup
- [ ] `svc-gateway/internal/handler` — Proxy/routing handlers
- [ ] `svc-gateway/internal/middleware` — Rate limiting (token bucket stub), JWT validation stub
- [ ] `svc-gateway/internal/config` — Service-specific config
- [ ] `svc-gateway/Dockerfile` — Multi-stage Go build
- [ ] `svc-gateway/fly.toml` — Fly.io deployment config
- [ ] Healthcheck endpoint at `GET /health`
- [ ] Returns 404 for undefined routes with JSON error

**Verification:**
- [ ] `go test ./svc-gateway/...` passes
- [ ] Docker build succeeds
- [ ] Manual: `curl http://localhost:8080/health` returns 200 OK

**Dependencies:** Task 1
**Estimated scope:** Medium

---

#### Task 3: User Service Scaffolding + Database
**Description:** Auth and user management service with PostgreSQL schema.

**Acceptance criteria:**
- [ ] `svc-user/cmd/main.go` — HTTP server + NATS subscriber setup
- [ ] `svc-user/internal/handler` — Register, Login, GetProfile endpoints (stubs)
- [ ] `svc-user/internal/service` — Business logic layer
- [ ] `svc-user/internal/repository` — SQL queries for users, roles, permissions
- [ ] `svc-user/internal/model` — Domain types
- [ ] `svc-user/migrations/` — goose migrations:
  - `001_create_users_table.sql`
  - `002_create_roles_and_permissions.sql`
  - `003_create_refresh_tokens.sql`
- [ ] Password hashing with bcrypt (stub in service layer)
- [ ] JWT generation (RS256 stub) and validation helpers

**Verification:**
- [ ] `go test ./svc-user/...` passes (unit tests with mocked repo)
- [ ] Migrations run successfully with `goose up`
- [ ] Docker build succeeds

**Dependencies:** Task 1
**Estimated scope:** Large — may split further if needed

---

#### Task 4: NATS Infrastructure
**Description:** NATS server setup and basic pub/sub patterns.

**Acceptance criteria:**
- [ ] Docker Compose file with NATS server
- [ ] `pkg/natsclient` updated with publisher/subscriber interfaces
- [ ] Example event publishing in User Service (e.g., `user.registered`)
- [ ] NATS health check integration in services

**Verification:**
- [ ] `docker-compose up nats` starts successfully
- [ ] Services can connect and publish dummy events

**Dependencies:** Task 1
**Estimated scope:** Small

---

### Checkpoint: Foundation
- [ ] All shared packages build cleanly
- [ ] Gateway and User Service containers build
- [ ] NATS messaging works between services
- [ ] Review with human before proceeding

---

### Phase 2: Core Features — User & Auth

#### Task 5: User Registration & Login
**Description:** Full implementation of auth flow.

**Acceptance criteria:**
- [ ] `POST /api/v1/auth/register` — Email + password, validate uniqueness, hash password
- [ ] `POST /api/v1/auth/login` — Verify password, issue JWT (RS256) + refresh token
- [ ] `POST /api/v1/auth/refresh` — Rotate refresh token, issue new JWT
- [ ] `GET /api/v1/users/me` — Return current user profile
- [ ] Proper error responses (RFC 7807)
- [ ] OAuth Google stub (endpoint exists, logic placeholder)

**Verification:**
- [ ] Integration tests with testcontainers (Postgres)
- [ ] Manual: Register → Login → GetProfile flow works

**Dependencies:** Task 2, Task 3
**Estimated scope:** Medium

---

#### Task 6: Gateway JWT Integration
**Description:** Wire gateway to validate JWTs using User Service public key.

**Acceptance criteria:**
- [ ] Gateway fetches JWT public key from User Service on startup
- [ ] JWT middleware validates `Authorization: Bearer <token>` headers
- [ ] Valid tokens populate `user_id` and `roles` in request context
- [ ] Invalid/expired tokens return 401 with RFC 7807 error

**Verification:**
- [ ] Unit tests for JWT middleware
- [ ] Manual: Request without token → 401, with valid token → 200

**Dependencies:** Task 5
**Estimated scope:** Small

---

### Checkpoint: Core Auth
- [ ] Full auth flow works end-to-end
- [ ] Gateway protects routes with JWT
- [ ] All tests pass

---

### Phase 3: Brief & Video Services

#### Task 7: Brief Service Scaffolding
**Description:** Create Brief Service with AI interview stub and CRUD.

**Acceptance criteria:**
- [ ] `svc-brief/` service skeleton (same structure as User Service)
- [ ] Migrations for `briefs`, `brief_questions`, `brief_answers`, `brief_tags`
- [ ] `POST /api/v1/briefs` — Create brief with AI interview stub
- [ ] `GET /api/v1/briefs/:id` — Get brief by ID
- [ ] `GET /api/v1/briefs` — List briefs (with filtering)
- [ ] `PATCH /api/v1/briefs/:id` — Update brief
- [ ] `POST /api/v1/briefs/:id/publish` — Publish brief (stub payment check)

**Verification:**
- [ ] Unit tests for handlers and service
- [ ] Integration tests with testcontainers
- [ ] Manual: CRUD flow works via curl

**Dependencies:** Task 6
**Estimated scope:** Large

---

#### Task 8: Video Service Scaffolding
**Description:** Video upload, approval workflow, versioning.

**Acceptance criteria:**
- [ ] `svc-video/` service skeleton
- [ ] Migrations for `videos`, `video_revisions`, `video_approvals`, `video_feedback`
- [ ] `POST /api/v1/videos` — Create video entry (metadata only)
- [ ] `POST /api/v1/videos/:id/submit` — Submit for approval
- [ ] `POST /api/v1/videos/:id/approve` — Client approves (emits event)
- [ ] `POST /api/v1/videos/:id/reject` — Client rejects with feedback
- [ ] `POST /api/v1/videos/:id/revise` — Request revision
- [ ]盲 submissions: editors can't see each other's work

**Verification:**
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual: Submit → Approve/Reject flow

**Dependencies:** Task 7
**Estimated scope:** Large

---

### Checkpoint: Brief & Video
- [ ] Brief CRUD works
- [ ] Video submission and approval flow works
- [ ] Events are emitted to NATS

---

### Phase 4: Campaign, Shopify, Performance

#### Task 9: Campaign Service
**Description:** Campaign creation for approved videos.

**Acceptance criteria:**
- [ ] `svc-campaign/` skeleton
- [ ] Migrations for `campaigns`, `campaign_videos`, `campaign_budgets`
- [ ] `POST /api/v1/campaigns` — Create campaign with approved videos
- [ ] Ad account placeholder endpoints
- [ ] Campaign status lifecycle (draft → active → ended)

**Verification:**
- [ ] Unit tests
- [ ] Manual: Create campaign flow

**Dependencies:** Task 8
**Estimated scope:** Medium

---

#### Task 10: Shopify Service
**Description:** Custom links and webhook ingestion.

**Acceptance criteria:**
- [ ] `svc-shopify/` skeleton
- [ ] Migrations for `shopify_stores`, `video_links`, `orders`, `attributions`
- [ ] `POST /api/v1/shopify/webhook` — Receive order webhooks
- [ ] Attribution logic: match discount code / UTM to video
- [ ] `GET /api/v1/links/:video_id` — Generate custom link

**Verification:**
- [ ] Unit tests for attribution logic
- [ ] Manual: Send test webhook → verify attribution

**Dependencies:** Task 9
**Estimated scope:** Medium

---

#### Task 11: Performance Service
**Description:** Sales aggregation and leaderboards.

**Acceptance criteria:**
- [ ] `svc-performance/` skeleton
- [ ] Migrations for `video_sales`, `editor_sales`, etc.
- [ ] Consumes `sale.attributed` events from NATS
- [ ] `GET /api/v1/leaderboards/:brief_id` — Return rankings
- [ ] Time-series aggregation endpoints

**Verification:**
- [ ] Unit tests
- [ ] Manual: Attributed sale appears in leaderboard

**Dependencies:** Task 10
**Estimated scope:** Medium

---

### Checkpoint: Campaign Loop
- [ ] Full campaign creation flow works
- [ ] Shopify webhooks attribute correctly
- [ ] Performance data aggregates and shows leaderboards

---

### Phase 5: Payout & Notifications

#### Task 12: Payout Service
**Description:** Calculate earnings and integrate with Ruul.io.

**Acceptance criteria:**
- [ ] `svc-payout/` skeleton
- [ ] Migrations for `payouts`, `transactions`, `balances`
- [ ] Earnings calculation: free tier ($500) then 5% fee
- [ ] 14-day hold logic
- [ ] Ruul.io bulk payout creation (API stub)
- [ ] DodoPayments webhook handling (deposit confirmation)

**Verification:**
- [ ] Unit tests for fee calculations
- [ ] Manual: Verify earning math with sample data

**Dependencies:** Task 11
**Estimated scope:** Large

---

#### Task 13: Notification Service
**Description:** Real-time WebSocket notifications.

**Acceptance criteria:**
- [ ] `svc-notification/` skeleton
- [ ] WebSocket upgrade endpoint
- [ ] Consumes events from NATS, pushes to connected clients
- [ ] `POST /api/v1/notifications/preferences` — User preferences
- [ ] Fallback email queue (placeholder)

**Verification:**
- [ ] Integration test: WebSocket client receives event
- [ ] Manual: Connect WS, trigger event, receive message

**Dependencies:** Task 6
**Estimated scope:** Medium

---

### Checkpoint: Payouts & Notifications
- [ ] Payout math is correct
- [ ] WebSocket events flow end-to-end
- [ ] All core loops are complete

---

### Phase 6: Admin & AI Support

#### Task 14: Admin Service
**Description:** Admin dashboard endpoints.

**Acceptance criteria:**
- [ ] `svc-admin/` skeleton
- [ ] Migrations for `admin_actions`, `disputes`
- [ ] RBAC middleware (check granular permissions)
- [ ] `POST /api/v1/admin/users/:id/ban` — Ban user
- [ ] `GET /api/v1/admin/disputes` — List disputes
- [ ] `POST /api/v1/admin/payouts/:id/override` — Override payout
- [ ] `GET /api/v1/admin/moderation-queue` — Content review

**Verification:**
- [ ] Unit tests
- [ ] Manual: Admin operations work

**Dependencies:** Task 5
**Estimated scope:** Medium

---

#### Task 15: AI Support Service
**Description:** Conversational support chat.

**Acceptance criteria:**
- [ ] `svc-ai-support/` skeleton
- [ ] Migrations for `conversations`, `messages`
- [ ] `POST /api/v1/support/chat` — Chat endpoint (stub AI response)
- [ ] Context loading: fetch user's briefs/videos (internal API calls)
- [ ] Escalation logic: route to admin when confidence low

**Verification:**
- [ ] Unit tests
- [ ] Manual: Chat → Escalation flow

**Dependencies:** Task 14
**Estimated scope:** Medium

---

### Checkpoint: Complete MVP
- [ ] All 11 services running
- [ ] Full end-to-end flow tested
- [ ] Admin and AI support functional
- [ ] Ready for production review

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Ruul.io API changes | High | Abstract behind interface, mock in tests |
| DodoPayments API changes | High | Abstract behind interface, mock in tests |
| NATS operational complexity | Medium | Start with single NATS node, docker-compose local |
| Schema drift across services | Medium | Strict migrations, never skip |
| 11 services is over-scoped for initial build | High | **Prioritize foundation + Phase 2-3 first**, defer Phase 4-6 |

## Open Questions

1. Do you want me to implement **all 11 services** now, or start with a **smaller subset** (e.g., gateway + user + brief + video)?
2. Should I create **OpenAPI specs** before implementing endpoints?
3. Do you have **Fly.io** and **Neon** accounts set up already, or should deployment configs be stubs?

---

**Recommendation:** Given the 11-service scope, I suggest implementing Phases 1-3 first (Foundation + Auth + Brief/Video), which gets you to a working "upload → brief → video submission" flow. Then iterate on Phases 4-6. This keeps each checkpoint verifiable and avoids building too much before validation.
