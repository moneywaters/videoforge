# VideoForge User Service

User authentication and profile management service for VideoForge.

## Overview

The User Service (`svc-user`) provides authentication, authorization, and user profile management for the VideoForge platform. It is built with Go 1.23 and uses PostgreSQL for data persistence.

## Features

- User registration with email/password
- JWT-based authentication (RS256)
- Refresh token rotation
- User profile management
- Role-based access control

## API Endpoints

### Authentication Endpoints

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| POST | `/api/v1/auth/register` | Register a new user | No |
| POST | `/api/v1/auth/login` | Login and get tokens | No |
| POST | `/api/v1/auth/refresh` | Refresh access token | No |
| POST | `/api/v1/auth/logout` | Logout and invalidate token | Yes |

### Profile Endpoints

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| GET | `/api/v1/users/me` | Get current user profile | Yes |
| PUT | `/api/v1/users/me` | Update current user profile | Yes |

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/.well-known/jwks.json` | Get JSON Web Key Set |
| GET | `/health` | Health check |

## Request/Response Examples

### Register

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "client",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Login

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123"
  }'
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "550e8400-e29b-41d4-a716-446655440000",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "client",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Refresh Token

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "660e8400-e29b-41d4-a716-446655440001",
  "user": {...}
}
```

### Get Profile

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "client",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "last_login_at": "2024-01-15T10:30:00Z"
}
```

### Update Profile

**Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane"
  }'
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "first_name": "Jane",
  "last_name": "Doe",
  "role": "client",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

## JWT Configuration

The service uses RS256 for JWT signing. For development, a key pair is automatically generated. For production, set the following environment variables:

- `JWT_PRIVATE_KEY` - RSA private key in PEM format
- `JWT_PUBLIC_KEY` - RSA public key in PEM format
- `JWT_KEY_ID` - Key identifier (default: "videoforge-key-001")

## Error Responses

Errors follow RFC 7807 Problem Details format:

```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "email is required"
}
```

## Database Schema

### Users Table
- `id` (UUID) - Primary key
- `email` (VARCHAR) - Unique email address
- `password_hash` (VARCHAR) - Bcrypt hash of password
- `first_name` (VARCHAR) - User's first name
- `last_name` (VARCHAR) - User's last name
- `role` (VARCHAR) - User role (client, editor, ad_specialist, admin, support_ai)
- `status` (VARCHAR) - Account status (active, banned)
- `created_at` (TIMESTAMP) - Creation timestamp
- `updated_at` (TIMESTAMP) - Last update timestamp
- `last_login_at` (TIMESTAMP) - Last login timestamp

### Refresh Tokens Table
- `id` (UUID) - Primary key
- `user_id` (UUID) - Foreign key to users
- `token_hash` (VARCHAR) - SHA256 hash of token
- `expires_at` (TIMESTAMP) - Expiration timestamp
- `created_at` (TIMESTAMP) - Creation timestamp

### Roles Table
- `id` (UUID) - Primary key
- `name` (VARCHAR) - Unique role name
- `description` (TEXT) - Role description

### Permissions Table
- `id` (UUID) - Primary key
- `name` (VARCHAR) - Unique permission name
- `description` (TEXT) - Permission description

### User Roles Table
- `user_id` (UUID) - Foreign key to users
- `role_id` (UUID) - Foreign key to roles

## Running Locally

```bash
# Set environment variables
export DATABASE_URL="postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable"
export PORT=8080

# Run migrations (using goose)
goose -dir migrations postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable up

# Run the service
go run cmd/main.go
```

## Docker

```bash
# Build the Docker image
docker build -t videoforge/svc-user .

# Run the container
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://videoforge:password@host.docker.internal:5432/videoforge?sslmode=disable" \
  videoforge/svc-user
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `DATABASE_URL` | PostgreSQL connection URL | `postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable` |
| `LOG_LEVEL` | Logging level | info |
| `JWT_PRIVATE_KEY` | RSA private key (PEM) | Auto-generated |
| `JWT_PUBLIC_KEY` | RSA public key (PEM) | Auto-generated |
| `JWT_KEY_ID` | Key identifier | videoforge-key-001 |