# Spec: VideoForge Backend

## Objective

Build a performance-based video marketplace backend using granular microservices on Fly.io. Clients upload raw footage, an AI agent interviews them to build a creative brief, editors compete to produce videos, ad specialists run paid campaigns, and everyone gets paid based on verified Shopify sales — not views.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.23+ |
| Database | Neon (PostgreSQL) — serverless, per-service schemas |
| API Style | REST + WebSockets |
| Hosting | Fly.io |
| Auth | Neon Auth  |
| Message Queue | NATS (Fly.io hosted) |
| File Storage | Storj (S3-compatible, Storj.io) |
| Payments (client→platform) | DodoPayments.com (merchant of record) |
| Payouts (platform→freelancers) | Ruul.io Business API (agent of record) |
| Real-time | WebSockets via API Gateway |

## Microservices Architecture

Each service is independently deployable, owns its own schema, and communicates via NATS + internal REST.

```
┌─────────────────────────────────────────────────────────────┐
│                        Fly.io Edge                          │
│                    ┌──────────────┐                         │
│                    │ API Gateway  │ ← JWT validation, rate  │
│                    │   (Go)       │   limiting, WS upgrade  │
│                    └──────┬───────┘                         │
└───────────────────────────┬─────────────────────────────────┘
                            │ NATS + internal mTLS
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ User Service │  │ Brief Service│  │ Video Service│
│   (Go)       │  │   (Go)       │  │   (Go)       │
│ - Auth       │  │ - AI brief   │  │ - Uploads    │
│ - Profiles   │  │ - Brief CRUD │  │ - Approval   │
│ - Roles      │  │ - Matching   │  │ - Storage    │
└──────────────┘  └──────────────┘  └──────────────┘
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│Campaign Svc  │  │Shopify Svc   │  │Performance   │
│   (Go)       │  │   (Go)       │  │   Service    │
│ - Campaigns  │  │ - Custom     │  │   (Go)       │
│ - Ad accounts│  │   links      │  │ - Sales attr.│
│   (placehold)│  │ - Webhooks   │  │ - Rankings   │
└──────────────┘  └──────────────┘  └──────────────┘
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Payout Svc   │  │ Notification │  │ Admin Svc    │
│   (Go)       │  │   Service    │  │   (Go)       │
│ - Earnings   │  │   (Go)       │  │ - Permissions│
│ - Fee calc   │  │ - WebSockets │  │ - Support    │
│ - Dodo API   │  │ - Email/push │  │ - Moderation │
└──────────────┘  └──────────────┘  └──────────────┘
┌──────────────┐
│ AI Support   │
│   Service    │
│   (Go)       │
│ - Chat agent │
│ - Escalation │
└──────────────┘
```

## Service Details

### 1. API Gateway (`svc-gateway`)
- Public entry point. All external traffic goes here.
- JWT validation (RS256, public key from User Service).
- Rate limiting per client IP + user ID.
- WebSocket upgrade and connection management.
- Routes requests to internal services via NATS request-reply.
- No business logic.

### 2. User Service (`svc-user`)
- **Register/Login:** Email + password, OAuth (Google).
- **Roles:** `client`, `editor`, `ad_specialist`, `admin`, `support_ai`.
- **Admin permissions:** Granular — `users:read`, `users:ban`, `payouts:override`, `campaigns:audit`, `support:escalate`, etc.
- **Profiles:** Role-specific fields (editor portfolio, ad specialist track record, client business info).
- **Onboarding:** Step-by-step flow per role.
- **Schema:** `users`, `roles`, `permissions`, `user_roles`, `refresh_tokens`.

### 3. Brief Service (`svc-brief`)
- **AI Interview:** Conversational endpoint that asks the client about goals, target audience, tone, style preferences, CTA. Compiles into structured brief.
- **Brief CRUD:** Create, update, publish, close.
- **Bounty budget:** Client specifies freelancer bounty budget. Brief is only published after DodoPayments confirms bounty deposit.
- **Matching:** Editors see briefs matching their skills/tags.
- **Submissions control:** Client specifies how many videos they want (e.g., 5). Can request more submissions later from the same pool (no extra budget).
- **Blind submissions:** Editors NEVER see each other's videos before client approval. They only see leaderboard numbers (rank, sales) after campaigns end.
- **Schema:** `briefs`, `brief_questions`, `brief_answers`, `brief_tags`, `brief_submission_limits`, `bounty_deposits`.

### 4. Video Service (`svc-video`)
- **Upload:** Presigned URL to Storj. Editors upload edited videos.
- **Approval workflow:** Client reviews → approve / reject with feedback / request revision.
- **Versioning:** Multiple revisions tracked.
- **Metadata:** Duration, resolution, thumbnail (placeholder).
- **Schema:** `videos`, `video_revisions`, `video_approvals`, `video_feedback`.

### 5. Campaign Service (`svc-campaign`)
- **Campaign creation:** Ad specialist creates and fully owns their campaign for approved videos. They have full control over setup, targeting, and optimization to maximize revenue.
- **Ad account placeholders:** Endpoints for Meta/TikTok API integration (future). For MVP, manual campaign setup with platform-owned ad account.
- **Budget tracking:** Link to client's ad budget.
- **Schema:** `campaigns`, `campaign_videos`, `ad_accounts` (placeholder), `campaign_budgets`.

### 6. Shopify Service (`svc-shopify`)
- **Custom links:** Generate unique discount/affiliate link per video. Placeholder for plugin API integration.
- **Webhook ingestion:** Receive Shopify order webhooks.
- **Attribution:** Match order to video via discount code / UTM parameter.
- **Schema:** `shopify_stores`, `video_links`, `orders`, `attributions`.

### 7. Performance Service (`svc-performance`)
- **Sales aggregation:** Total sales per video, per editor, per ad specialist, per campaign.
- **Rankings:** Leaderboards within cohorts.
- **Analytics:** Time-series data for dashboards.
- **Anomaly detection:** Flag suspicious patterns (placeholder rules).
- **Schema:** `video_sales`, `editor_sales`, `specialist_sales`, `campaign_sales`, `leaderboards`.

### 8. Payout Service (`svc-payout`)
- **Earnings calculation:** 
  - First **$500 in verified sales**: platform fee = 0%.
  - After $500: platform takes **5%** of editor + ad specialist fees only.
  - Client ad spend goes directly to Meta/TikTok (not through platform).
  - Freelancer bounty budget is split proportionally by verified sales.
- **DodoPayments integration:** 
  - DodoPayments is the Merchant of Record for collecting client payments (platform fees + freelancer bounty budget).
  - Client deposits freelancer bounty upfront via DodoPayments checkout when creating a brief.
  - **CRITICAL FINDING:** DodoPayments' `/payouts` endpoint pays out to the *business* (us), not to individual freelancers.
- **Ruul.io integration:**
  - Ruul Business API handles all **platform → freelancer** payouts.
  - Ruul acts as Agent of Record for contractors — they handle KYC/AML, compliance, tax forms, and generate VAT invoices.
  - Supports bulk payouts (pay multiple editors/ad specialists in one batch).
  - Global coverage: 190 countries, 140 currencies.
  - Fast settlement: contractors receive payout in ~1 business day.
  - We create payment requests via API per freelancer earnings batch.
- **Hold periods:** 14-day hold for refunds/chargebacks before freelancer payout eligibility.
- **Schema:** `payouts`, `payout_rules`, `transactions`, `balances`, `ruul_payout_batches`, `ruul_payout_requests`.

### 9. Notification Service (`svc-notification`)
- **WebSockets:** Real-time updates to clients (new submission, approval status, sales milestone).
- **Events consumed:** `video.submitted`, `video.approved`, `sale.attributed`, `payout.released`.
- **Channels:** WS (primary), email (fallback), push (future).
- **Schema:** `notifications`, `user_preferences`, `ws_connections`.

### 10. Admin Service (`svc-admin`)
- **User management:** Search, ban, role assignment.
- **Permission system:** RBAC with granular actions.
- **Dispute resolution:** View evidence, override payouts.
- **Moderation:** Flag content, manual review queue.
- **Schema:** `admin_actions`, `disputes`, `moderation_queue`.

### 11. AI Support Service (`svc-ai-support`)
- **Chat endpoint:** Conversational support for all user types.
- **Context:** Has read access to user's briefs, videos, payouts (via internal API calls).
- **Escalation:** Hands off to human admin when confidence is low or user requests it.
- **Schema:** `conversations`, `messages`, `escalations`.

## Data Flow: Core Loop

```
1. CLIENT uploads raw footage + website URL
        ↓
2. AI BRIEF SERVICE interviews client → structured brief
        ↓
3. CLIENT deposits freelancer bounty budget via DodoPayments checkout
        ↓
4. BRIEF published → editors notified
        ↓
5. EDITORS see brief, submit videos (up to client-specified limit, blind)
        ↓
6. CLIENT reviews → approves N videos for ad testing
        ↓
7. AD SPECIALIST picks approved videos → creates & owns campaign
        ↓
8. SHOPIFY SERVICE generates custom link per video
        ↓
9. Campaign runs (client pays Meta/TikTok directly for ad spend)
        ↓
10. SHOPIFY webhook → order attributed to video
        ↓
11. PERFORMANCE SERVICE aggregates sales
        ↓
12. PAYOUT SERVICE calculates earnings (first $500 free, then 5% fee)
        ↓
13. Platform pays freelancers via Ruul.io bulk payout
        ↓
14. NOTIFICATION SERVICE pushes real-time updates
```

## API Conventions

- Base path: `/api/v1`
- Content-Type: `application/json`
- Errors: RFC 7807 `application/problem+json`
- Pagination: `?page=1&limit=20`, response includes `total`, `has_more`
- Sorting: `?sort=created_at:desc`
- Filtering: `?status=approved&role=editor`
- IDs: UUID v7 (time-sortable)
- Timestamps: RFC 3339

## WebSocket Events

```json
{
  "event": "video.submitted",
  "payload": {
    "brief_id": "uuid",
    "video_id": "uuid",
    "editor_id": "uuid",
    "submitted_at": "2026-01-15T10:00:00Z"
  }
}
```

Events: `video.submitted`, `video.approved`, `video.rejected`, `sale.attributed`, `payout.released`, `campaign.started`, `campaign.ended`.

## Project Structure (Per Service)

```
svc-{name}/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── config/              # Env config
│   ├── handler/             # HTTP handlers
│   ├── service/             # Business logic
│   ├── repository/          # DB queries
│   ├── model/               # Domain models
│   ├── middleware/          # Auth, logging, recovery
│   └── client/              # Internal service clients
├── migrations/              # goose or golang-migrate
├── proto/                   # If gRPC internal (future)
├── Dockerfile
├── fly.toml
└── go.mod
```

## Code Style

```go
// Handler: thin, validates, delegates to service
func (h *VideoHandler) Submit(w http.ResponseWriter, r *http.Request) {
    var req SubmitVideoRequest
    if err := decode(r, &req); err != nil {
        respondError(w, err)
        return
    }

    video, err := h.service.Submit(r.Context(), req)
    if err != nil {
        respondError(w, err)
        return
    }

    respond(w, http.StatusCreated, video)
}

// Service: business logic, no HTTP
func (s *VideoService) Submit(ctx context.Context, req SubmitVideoRequest) (*Video, error) {
    // validate business rules
    // call repository
    // emit event to NATS
}

// Repository: SQL only
func (r *VideoRepo) Create(ctx context.Context, v *Video) error {
    // sqlc-generated or hand-written
}
```

## Testing Strategy

- **Unit:** `go test` per service, mock repositories.
- **Integration:** `testcontainers-go` for Postgres, NATS.
- **Contract:** Shared OpenAPI spec, validated in CI.
- **Coverage target:** 70% business logic, 40% handlers.

## Boundaries

- **Always:** Run `go test ./...` before commit. Use `gofumpt`. Log with `slog`.
- **Ask first:** Add a new dependency. Change DB schema of another service. Add a new microservice.
- **Never:** Commit secrets. Call another service's DB directly. Skip migrations.

## Success Criteria

- [ ] All 11 services deploy independently to Fly.io
- [ ] Client can complete full loop: upload → brief → deposit bounty → submissions → approval → campaign → sale → payout
- [ ] Shopify webhook correctly attributes sales to videos
- [ ] Payouts calculate correctly (free tier + 5% fee)
- [ ] Ruul.io bulk payouts process successfully to freelancers
- [ ] WebSockets deliver real-time events
- [ ] Admin can ban users, override payouts, view disputes
- [ ] AI support handles 80% of queries without escalation

## Open Questions

1. ~~What is X (free sales threshold)?~~ → **$500 global free tier, then 5% fee**
2. ~~DodoPayments API details~~ → **DodoPayments handles client→platform payments. Client deposits freelancer bounty upfront via DodoPayments checkout.**
3. ~~Do editors see each other's submissions?~~ → **Blind. Only leaderboard numbers visible.**
4. ~~How does "request more submissions" work?~~ → **Same pool, no extra budget.**
5. ~~Ad specialist assignment?~~ → **Ad specialist creates and fully owns their own campaign.**
6. ~~Freelancer payout provider?~~ → **Ruul.io Business API for all platform→freelancer payouts.**
7. ~~How does client deposit freelancer bounty budget?~~ → **Upfront via DodoPayments checkout when creating a brief.**
