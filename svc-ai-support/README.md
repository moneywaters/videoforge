# AI Support Service (svc-ai-support)

AI-powered conversational support service for VideoForge.

## Overview

The AI Support Service provides conversational support for all user types with the ability to escalate to human agents when needed. It has read access to user briefs, videos, and payouts through internal API calls.

## Features

- **Chat endpoint**: Conversational support for all user types
- **Context**: Has read access to user's briefs, videos, payouts (via internal API calls - STUB)
- **Escalation**: Hands off to human admin when confidence is low or user requests it

## API Endpoints

### POST /api/v1/support/chat

Start or continue a conversation.

**Request Body:**
```json
{
  "message": "How do payouts work?",
  "conversation_id": "optional-uuid"
}
```

**Response:**
```json
{
  "conversation_id": "uuid",
  "messages": [...],
  "ai_response": "I can help with payout questions...",
  "should_escalate": false
}
```

### GET /api/v1/support/conversations

List user's conversations.

### GET /api/v1/support/conversations/:id

Get conversation with all messages.

### POST /api/v1/support/conversations/:id/escalate

Escalate to human support.

**Request Body:**
```json
{
  "reason": "optional reason"
}
```

### POST /api/v1/support/escalations/:id/resolve

Admin resolves escalation (admin only).

### GET /api/v1/support/escalations

List escalations (admin only).

## Database Schema

See `migrations/` directory for database schema:

- `001_create_conversations.sql` - conversations table
- `002_create_messages.sql` - messages table
- `003_create_escalations.sql` - escalations table

## AI Response Logic (STUB)

The AI response generation is implemented as a STUB that returns canned responses based on keyword matching:

- "payout"/"payment" → Payout-related response
- "brief"/"project" → Brief-related response
- "video"/"edit" → Video-related response
- "campaign"/"ad" → Campaign-related response
- Otherwise → Default helpful response

### Escalation Triggers

1. User types "human", "agent", "support", "help", "escalate"
2. AI confidence is low (10% random chance when confidence < 0.5)
3. User explicitly clicks escalate button

## Environment Variables

- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `NATS_URL` - NATS server URL (default: nats://localhost:4222)
- `JWT_PUBLIC_KEY` - RSA public key for JWT validation
- `LOG_LEVEL` - Logging level (default: info)

## Running Locally

```bash
# Set environment variables
export DATABASE_URL="postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable"
export NATS_URL="nats://localhost:4222"

# Run the service
go run cmd/main.go
```

## Running with Docker

```bash
docker build -t svc-ai-support .
docker run -p 8080:8080 svc-ai-support
```

## Architecture

```
svc-ai-support/
├── cmd/main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── handler/          # HTTP handlers
│   ├── service/          # Business logic
│   ├── repository/       # Database operations
│   └── model/            # Data models
├── migrations/            # Database migrations
├── Dockerfile
├── fly.toml
└── README.md
```

## Cross-Service Integration (STUB)

For MVP, user context is not loaded from other services. TODO comments are included to implement:

- Get user profile from User Service
- List user's briefs from Brief Service
- List user's videos from Video Service
- Get user's balance from Payout Service

## NATS Events

- `support.escalated` - Emitted when a conversation is escalated