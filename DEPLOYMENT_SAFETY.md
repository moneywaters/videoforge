# Deployment Safety Guidelines

## Critical Rules (Must Never Break)

### 1. No os.Exit() in Business Logic
- os.Exit(1) is ONLY for catastrophic config loading failures
- Database schema issues are NOT catastrophic - service should start and return 503

### 2. No Runtime Schema Creation
- Never auto-create tables at service startup
- Use separate SQL migration files in `svc-*/migrations/`
- Run migrations manually before deploying code that uses new tables
- Example: `psql $DATABASE_URL < svc-brief/migrations/006_create_brief_files.sql`

### 3. Context Keys Must Be Typed Constants
- Always use typed constants from `pkg/middleware/auth.go`
- Type ContextKey string; const UserIDContextKey ContextKey = "user_id"
- NEVER use plain strings like ctx.Value("user_id")
- Remove dead code in `svc-*/internal/middleware/` that defines duplicate string keys

### 4. Interface-Based Dependency Injection
- All service dependencies must be interfaces
- Constructor signature: NewService(repo RepoInterface, storage StorageInterface, ...)
- Allows nil/optional dependencies for graceful degradation

### 5. Add Features Without Modifying Existing Routes
- New features get NEW routes (POST /api/v1/resource/new-feature)
- NEVER modify existing working handlers
- Always verify existing endpoints still return expected responses

## Deployment Checklist

Before deploying any change:
1. [ ] Run `go build ./...` locally
2. [ ] Run `go test ./svc-brief/...` 
3. [ ] Verify existing POST /api/v1/briefs still returns 201
4. [ ] Verify new endpoints work (if any)
5. [ ] If adding database tables: Run migration manually FIRST
6. [ ] Deploy to staging first
7. [ ] Verify brief creation on staging
8. [ ] Deploy to production

## Safe Feature Addition Pattern

### Phase 1: Database (if needed)
- Create migration file
- Run manually in production
- Verify table exists

### Phase 2: Repository Layer (if needed)
- Add new repository with interface
- Add to service constructor
- Pass nil if not needed yet

### Phase 3: Service Layer (if needed)
- Add new methods
- Don't modify existing methods
- Handle nil repository gracefully

### Phase 4: HTTP Handlers
- Add new routes in main.go
- Don't modify existing route handlers
- Add new handler files if needed

### Phase 5: Frontend
- Add new API client methods
- Don't modify existing API calls
- Feature-flag if needed

### Phase 6: Verification
- Test existing features work
- Test new features work
- Deploy incrementally