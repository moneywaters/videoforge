# Deployment and Launch Guide

## Pre-Launch Checklist

### Code Quality
- [ ] All tests pass (`make test`)
- [ ] Build succeeds for all services (`make build`)
- [ ] Lint passes (`make lint`)
- [ ] Code reviewed across all services
- [ ] No unresolved TODO comments in production paths
- [ ] Error handling covers expected failure modes

### Security
- [x] JWT RS256 with asymmetric signing (private key in User Service only)
- [ ] Rate limiting configured on Gateway (100 req/s per IP, 10 burst)
- [ ] CORS configured to specific origins (not `*` in production)
- [ ] Input validation on all user-facing endpoints
- [ ] SQL queries use parameterized statements (pgx)
- [ ] Secrets loaded from environment variables (never in code)
- [ ] `npm audit` / `govulncheck` run on dependencies

### Infrastructure
- [ ] PostgreSQL databases provisioned (one per service on Neon)
- [ ] NATS server provisioned (Fly.io or managed)
- [ ] All migrations applied (`make migrate-up`)
- [ ] Environment variables set for all services
- [ ] DNS configured for Gateway
- [ ] SSL/TLS configured
- [ ] Health check endpoints responding (`GET /health`)

### Monitoring
- [ ] Structured logging configured (JSON in production)
- [ ] Error tracking service connected (Sentry or similar)
- [ ] Metrics dashboards created (request rate, latency, errors)
- [ ] Alerting configured for error rate spikes
- [ ] Database connection pool monitoring

## Deployment Steps

### 1. Provision Infrastructure

```bash
# Create Neon PostgreSQL databases (one per service)
# Or use a single Neon project with separate schemas

# Deploy NATS to Fly.io
fly apps create videoforge-nats
fly deploy --config nats/fly.toml

# Set secrets for all services
fly secrets set DATABASE_URL="postgres://..." --app videoforge-user
fly secrets set JWT_SECRET="..." --app videoforge-user
# ... repeat for each service
```

### 2. Deploy User Service First

```bash
cd svc-user
fly deploy

# Verify health check
curl https://videoforge-user.fly.dev/health
```

### 3. Deploy Remaining Services

```bash
# Deploy in dependency order:
# Gateway → User → Brief → Video → Campaign → Shopify → Performance → Payout → Notification → Admin → AI Support

make deploy-all
```

### 4. Apply Database Migrations

```bash
# For each service:
cd svc-user && goose postgres "$DATABASE_URL" up
cd svc-brief && goose postgres "$DATABASE_URL" up
# ... etc
```

### 5. Verify End-to-End Flow

```bash
# 1. Register a client
curl -X POST https://videoforge.fly.dev/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"client@example.com","password":"...","role":"client"}'

# 2. Login and get token
TOKEN=$(curl -X POST https://videoforge.fly.dev/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"client@example.com","password":"..."}' \
  | jq -r '.access_token')

# 3. Create a brief
curl -X POST https://videoforge.fly.dev/api/v1/briefs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Brief","description":"...","bounty_budget":1000}'

# 4. Verify full loop works...
```

## Rollback Plan

### Trigger Conditions
- Error rate > 2x baseline for any service
- P95 latency > 500ms for Gateway
- Database connection pool exhaustion
- Critical bug affecting payments or auth

### Rollback Steps

**Service Rollback:**
```bash
# Re-deploy previous version
cd svc-user
git checkout <previous-commit>
fly deploy
```

**Database Rollback:**
```bash
# Each migration should have a corresponding down migration
cd svc-user && goose postgres "$DATABASE_URL" down
```

**Feature Flag Rollback:**
If using feature flags (future enhancement):
```bash
# Disable feature immediately
# No deployment needed
```

### Rollback Time Targets
- Service redeploy: < 5 minutes
- Database rollback: < 15 minutes (per service)
- Full system rollback: < 30 minutes

## Post-Launch Monitoring

### First Hour
- [ ] Health checks return 200 for all services
- [ ] No new error types in error tracking
- [ ] Gateway latency within normal range
- [ ] Database connections stable
- [ ] NATS messages flowing (check subscriptions)

### First 24 Hours
- [ ] Error rate stable (< 0.1%)
- [ ] No memory leaks (check container metrics)
- [ ] Log volume normal (not excessive)
- [ ] User registration and login working
- [ ] Brief creation and video submission working

### First Week
- [ ] Performance metrics within targets
- [ ] No critical bugs reported
- [ ] Payout calculations verified with test data
- [ ] WebSocket connections stable
- [ ] Admin dashboard accessible

## Environment Configuration

### Development
```bash
docker-compose up -d  # Starts Postgres + NATS
# Run services individually with `go run ./cmd/main.go`
```

### Staging
- Deploy all services to staging Fly.io apps
- Use staging Neon database
- Run full integration test suite

### Production
- Deploy to production Fly.io apps
- Use production Neon database
- Enable all monitoring and alerting
- Restrict CORS to specific origins
- Enable rate limiting

## Known Limitations / TODOs

1. **DodoPayments integration**: Webhook handler exists but actual API calls are stubs
2. **Ruul.io integration**: Payout batch creation is stubbed
3. **AI Support**: Chat is keyword-based stub, not actual LLM integration
4. **Storj integration**: Upload URLs are mock/stub
5. **Meta/TikTok Ads API**: Campaign management uses placeholder ad accounts
6. **Email notifications**: Email channel is stubbed, only WebSocket is implemented
7. **Image/video processing**: Thumbnails and metadata extraction are placeholders
8. **Search**: Full-text search on briefs/users uses simple SQL LIKE, not proper search engine
