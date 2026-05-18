# Architecture Decision Records

## ADR-001: Microservices with Per-Service Database Schemas

**Status**: Accepted

**Context**: VideoForge requires 11 distinct domains (auth, briefs, videos, campaigns, etc.). We needed to decide between monolith and microservices, and how to handle data storage.

**Decision**: Use granular microservices, each with its own PostgreSQL schema. Services communicate via NATS for async events and HTTP for synchronous calls.

**Consequences**:
- ✅ Independent deployability — each service can be deployed, scaled, and updated independently
- ✅ Strong data ownership — no service can access another service's tables directly
- ✅ Technology flexibility per service if needed in the future
- ❌ Operational complexity — 11 services to monitor and manage
- ❌ Distributed transaction complexity — saga patterns needed for cross-service consistency

## ADR-002: NATS for Inter-Service Communication

**Status**: Accepted

**Context**: Services need to communicate asynchronously (events like `video.submitted`, `sale.attributed`). We evaluated NATS, RabbitMQ, and Kafka.

**Decision**: Use NATS (specifically NATS Core with request-reply) hosted on Fly.io.

**Consequences**:
- ✅ Lightweight and fast — single binary, minimal resource usage
- ✅ Native Go client with excellent performance
- ✅ Request-reply pattern maps well to our internal service calls
- ✅ JetStream available for persistence if needed later
- ❌ Less mature ecosystem than Kafka for complex stream processing
- ❌ Self-hosted means we manage availability

## ADR-003: JWT with RS256 for Authentication

**Status**: Accepted

**Context**: All services need to verify user identity. We needed a token format that doesn't require calling the auth service on every request.

**Decision**: Use JWT signed with RS256 (RSA + SHA-256). The User Service holds the private key for signing. All other services fetch the public key (via JWKS endpoint or config) to verify tokens independently.

**Consequences**:
- ✅ Stateless auth verification — no auth service dependency after token issuance
- ✅ Secure asymmetric signing — private key never leaves the auth service
- ✅ Standard format — well-supported across languages and frameworks
- ❌ Token revocation requires additional mechanism (we use refresh token rotation)
- ❌ Token size is larger than simple session IDs

## ADR-004: Go Workspaces for Multi-Module Management

**Status**: Accepted

**Context**: The project has a shared `pkg/` module and 11 service modules. We needed a way to manage dependencies and imports cleanly.

**Decision**: Use Go 1.23 workspaces (`go.work`) with a root module for shared packages and separate modules per service.

**Consequences**:
- ✅ Shared packages via workspace imports — no need to publish pkg/ separately
- ✅ Independent versioning per service
- ✅ Clear module boundaries
- ❌ Slightly more complex initial setup
- ❌ All services must use compatible Go versions

## ADR-005: DodoPayments + Ruul.io for Payment Processing

**Status**: Accepted

**Context**: The platform needs to collect money from clients and pay freelancers. We evaluated Stripe, PayPal, Wise, and others.

**Decision**: 
- **Client → Platform**: DodoPayments as Merchant of Record (handles tax, compliance, checkout)
- **Platform → Freelancers**: Ruul.io Business API as Agent of Record (handles KYC, tax forms, global payouts)

**Consequences**:
- ✅ DodoPayments handles merchant complexity (sales tax, VAT, chargebacks)
- ✅ Ruul.io handles contractor compliance in 190 countries
- ✅ Separation of concerns — client payments and freelancer payouts are decoupled
- ❌ Two integrations to maintain
- ❌ DodoPayments payouts endpoint pays the business, not individuals (mitigated by using Ruul for individual payouts)

## ADR-006: Blind Submissions for Fair Competition

**Status**: Accepted

**Context**: Editors compete to produce the best video for a brief. If they see each other's work, it could lead to copying or discouragement.

**Decision**: Editors can NEVER see each other's video submissions before a campaign ends. They only see leaderboard numbers (rank, sales) after campaigns conclude.

**Consequences**:
- ✅ Fair competition — each editor is judged on their own merit
- ✅ Prevents copycat submissions
- ✅ Focus on results (sales) rather than style imitation
- ❌ Less collaboration between editors
- ❌ Requires careful query filtering in Video Service to enforce blindness
