# VideoForge Deployment Guide

## Overview
This project uses a multi-service Go backend (deployed to Fly.io) and a Next.js frontend (deployed to Vercel).

## Prerequisites

### Required Accounts & Tools
- **GitHub account** with SSH key configured
- **Fly.io account** with `flyctl` installed
- **Vercel account** with Vercel CLI installed
- **Docker Desktop** running locally (for local builds)

### Environment Variables Setup

#### Repository Secrets (GitHub)
Set these in your GitHub repository Settings > Secrets and variables > Actions:

```
FLY_API_TOKEN        # Get from: fly auth token (or fly.io dashboard)
VERCEL_TOKEN         # Get from: Vercel dashboard > Account Settings > Tokens
VERCEL_ORG_ID        # From .vercel/project.json or Vercel dashboard
VERCEL_PROJECT_ID    # From .vercel/project.json or Vercel dashboard
NEXT_PUBLIC_API_URL  # e.g. https://videoforge-gateway.fly.dev/api/v1
```

#### Local Environment
Create `.env` in workspace root:
```env
DATABASE_URL=postgresql://...
JWT_SECRET=your-jwt-secret
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=https://videoforge-gateway.fly.dev/api/v1/auth/google/callback
FRONTEND_URL=https://cutthroatreels.com
```

## Deployment Steps

### 1. Backend Services (Fly.io)

Each Go service uses the shared `Dockerfile` at workspace root with `--build-arg SERVICE=<name>`.

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh
export PATH="$HOME/.fly/bin:$PATH"

# Login (if needed)
fly auth login

# Create per-service fly.toml files pointing to root Dockerfile
# Example for svc-user:
cat > fly-user.toml << TOML
app = "videoforge-user"
primary_region = "lax"

[build]
  dockerfile = "Dockerfile"
  args.SERVICE = "user"

[env]
  PORT = "8080"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

[vm]
  memory = "256mb"
  cpu_kind = "shared"
  cpus = 1

[experimental]
  auto_rollback = true
TOML

# Deploy (uses local Docker, then pushes to Fly.io registry)
fly deploy --config fly-user.toml --local-only
```

**Important**: `--local-only` works if Docker Desktop is running. The default remote-only builders struggle with workspace-root Dockerfiles.

**Repeat for each service**:
- `gateway` → `videoforge-gateway` (proxies all traffic, must route to internal `.internal`: URLs)
- `user` → `videoforge-user`
- `brief` → `videoforge-brief`
- `video` → `videoforge-video`
- `campaign` → `videoforge-campaign`
- `shopify` → `videoforge-shopify`
- `performance` → `videoforge-performance`
- `payout` → `videoforge-payout`
- `notification` → `videoforge-notification`
- `ai-support` → `videoforge-ai-support`
- `admin` → `videoforge-admin`

### 2. Frontend (Vercel)

```bash
# Install Vercel CLI (already installed via npm)
npm i -g vercel

# Link project (first time only)
cd frontend
vercel link

# Deploy to production
vercel deploy --prod
```

Or use the CI/CD workflow — it deploys automatically on every push to `main`.

### 3. Git Integration

```bash
# Add remote (if not exists)
git remote add origin git@github.com:moneywaters/videoforge.git

# Push
git push -u origin main
```

## GitHub Actions Automated Deployment

The file `.github/workflows/deploy.yml` handles everything on every `git push origin main`.

### How it works:
1. **Matrix build** — builds Docker images for all 11 services in parallel
2. **Pushes to GHCR** — stores images in GitHub Container Registry
3. **Deploys to Fly.io** — uses `flyctl deploy --image <ghcr-image>` with rolling strategy
4. **Deploys frontend** — installs deps, builds Next.js, deploys to Vercel

### To enable:
1. Create repo `moneywaters/videoforge` on GitHub
2. Set `FLY_API_TOKEN` in GitHub Secrets
3. Set `VERCEL_TOKEN`, `VERCEL_ORG_ID`, `VERCEL_PROJECT_ID` in GitHub Secrets
4. Set `NEXT_PUBLIC_API_URL` in GitHub Secrets

## Troubleshooting

### Fly.io build fails with "/pkg not found"
**Cause**: Fly's remote builder can't access workspace-root paths when `fly.toml` is in a subdirectory.  
**Fix**: Move/copy `fly.toml` to workspace root and set `dockerfile = "Dockerfile"` or use `--local-only`.

### Fly.io "failed to compute cache key"
**Cause**: Remote builder context doesn't include parent directory files.  
**Fix**: Use `--local-only` (requires Docker Desktop running locally).

### Vercel "Element type is invalid" (500 errors)
**Cause**: SSR boundary issue with shadcn/ui `SidebarProvider`/`AppSidebar` on App Router.  
**Fix**: Replace complex shadcn/ui sidebar layout with a simpler custom sidebar (`<aside>` + `<main>`). See `frontend/src/app/dashboard/layout.tsx` for working implementation.

### Registration returns 400
**Cause**: Backend `RegisterRequest` struct doesn't match frontend payload.  
**Fix**: Ensure `svc-user/cmd/main.go` has `first_name`, `last_name`, and `role` fields in `RegisterRequest`.

### Docker not running on macOS
```bash
open -a Docker
# wait 10 seconds for daemon to start
```

## Rollback

### Fly.io
```bash
fly releases --app <APP_NAME>
fly deploy --app <APP_NAME> --image registry.fly.io/<APP_NAME>:deployment-<ID>
```

### Vercel
Use Vercel dashboard > Deployments > select previous deployment > "Promote to Production".

## Architecture
```
GitHub repo
  ├── Dockerfile (shared multi-service build)
  ├── go.mod (workspace root)
  ├── pkg/ (shared packages)
  ├── svc-*/ (Go microservices)
  ├── frontend/ (Next.js 16 + shadcn/ui)
  ├── .github/workflows/deploy.yml (CI/CD)
  └── fly-*.toml (per-service Fly.io configs)
```
