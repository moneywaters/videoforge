# Brief Service (svc-brief)

Brief Service handles AI interviews, brief CRUD, matching editors to briefs, and submission limits for the VideoForge platform.

## Features

- **AI Interview**: Conversational endpoint (MVP STUB) that compiles user input into a structured brief
- **Brief CRUD**: Create, update, publish, close briefs
- **Bounty Budget**: Stub payment verification before publishing (bounty_deposited flag)
- **Matching**: Filter briefs by editor skills/tags
- **Submissions Control**: max submissions limit per brief
- **Blind Submissions**: Hide editor submissions from each other (is_blind flag)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/briefs | Create brief (starts as draft) |
| GET | /api/v1/briefs | List briefs (with filtering & pagination) |
| GET | /api/v1/briefs/:id | Get brief by ID |
| PATCH | /api/v1/briefs/:id | Update brief (draft only) |
| POST | /api/v1/briefs/:id/publish | Publish brief (check bounty_deposited) |
| POST | /api/v1/briefs/:id/close | Close brief |
| POST | /api/v1/briefs/:id/interview | AI interview STUB |
| GET | /api/v1/briefs/matching | Get briefs matching editor tags |
| POST | /api/v1/briefs/:id/view | Mark brief as viewed |

## Database Schema

Tables in `brief` schema:

- **briefs**: Core brief data (status, bounty, submissions limit, blind flag)
- **brief_tags**: Tags for matching editors to briefs
- **brief_questions**: AI interview Q&A storage
- **brief_editor_views**: Track which editors viewed which briefs

## Authentication

Uses JWT from `Authorization: Bearer <token>` header.

Configure JWT public key via:
- `JWT_PUBLIC_KEY_PATH` - Path to public key file
- `JWT_PUBLIC_KEY` - Base64 encoded public key
- `JWKS_URL` - URL to fetch JWKS from (stub)

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Server port | 8080 |
| ENVIRONMENT | Environment | development |
| DATABASE_URL | PostgreSQL connection string | postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable |
| JWT_PUBLIC_KEY | JWT public key (base64) | - |
| JWT_PUBLIC_KEY_PATH | Path to JWT public key file | - |

## Running

```bash
# With Docker
docker build -t svc-brief .
docker run -p 8080:8080 svc-brief

# Local development
go run cmd/main.go
```

## Brief Status Flow

```
draft -> published -> closed
         ^           |
         |___________|
```

- Briefs start as `draft`
- Can only be published if `bounty_deposited = true`
- Once published, can be closed but never reopened