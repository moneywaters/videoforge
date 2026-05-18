# svc-payout

Payout service for VideoForge platform - handles earnings calculation, payments, and freelancer payouts.

## Overview

The Payout Service manages:
- Earnings calculation for editors and ad specialists
- Platform fee calculations with tiered structure
- Balance management (available vs pending)
- Hold period management (14-day refund/chargeback protection)
- Ruul.io integration for freelancer payouts
- DodoPayments integration for client deposits

## Earnings Model

### Platform Fee Structure
- **First $500 in verified sales**: 0% platform fee
- **After $500**: 5% platform fee on editor + specialist earnings only
- Client ad spend goes directly to Meta/TikTok (not through platform)

### Calculation Logic
```
For each user (editor or specialist):
  Calculate total_verified_sales from Performance Service
  
  If total_verified_sales <= $500:
    user_earnings = total_verified_sales * (user_share_of_bounty / total_bounty)
    platform_fee = 0
  Else:
    bounty_portion = total_verified_sales * (user_share_of_bounty / total_bounty)
    platform_fee = bounty_portion * 0.05
    user_earnings = bounty_portion - platform_fee
  
  Add to user's pending balance (with 14-day hold)
  After hold period: move to available balance
```

## Integration

### DodoPayments
- **Role**: Merchant of Record for client payments
- Collects: platform fees + freelancer bounty budget
- **Note**: DodoPayments pays out to *business* (VideoForge), not individual freelancers
- Webhook endpoint: `POST /api/v1/payouts/webhook/dodo`

### Ruul.io
- **Role**: Platform → Freelancer payouts
- **Features**: Bulk payouts support via Business API
- Webhook endpoint: `POST /api/v1/payouts/webhook/ruul`

## API Endpoints

### Payouts
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/payouts` | List payouts (user or admin all) |
| GET | `/api/v1/payouts/:id` | Get payout details |
| GET | `/api/v1/balance` | Get user's balance |
| GET | `/api/v1/earnings` | Get earnings history |
| POST | `/api/v1/payouts/calculate` | Calculate earnings breakdown |

### Ruul Batches
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/payouts/batches` | Create payout batch |
| GET | `/api/v1/payouts/batches` | List all batches |
| GET | `/api/v1/payouts/batches/:id` | Get batch details |
| POST | `/api/v1/payouts/batches/:id/process` | Process batch via Ruul |

### Webhooks
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/payouts/webhook/dodo` | DodoPayments webhook |
| POST | `/api/v1/payouts/webhook/ruul` | Ruul webhook |

## NATS Events

### Subscriptions
- `sale.attributed` - Processing sale attribution events to trigger earnings calculation

### Event Publishing
- `payout.created` - When new payout is created
- `balance.updated` - When balance changes
- `hold.released` - When hold period ends

## Database Schema

### Tables
- `payouts` - Individual payout records
- `payout_rules` - Platform fee rules
- `transactions` - Financial ledger
- `balances` - User balances
- `ruul_payout_batches` - Ruul batch records
- `ruul_payout_requests` - Individual requests in batches

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment | `development` |
| `DATABASE_URL` | PostgreSQL connection | `postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable` |
| `NATS_URL` | NATS server URL | `nats://localhost:4222` |
| `HOLD_PERIOD_DAYS` | Hold period in days | `14` |
| `RUUL_API_KEY` | Ruul.io API key | - |
| `RUUL_BASE_URL` | Ruul.io API URL | `https://api.ruul.io/v1` |
| `DODO_API_KEY` | DodoPayments API key | - |
| `DODO_SECRET` | DodoPayments secret | - |
| `DODO_BASE_URL` | DodoPayments API URL | `https://api.dodopayments.com/v1` |

## Deployment

Built for Fly.io deployment:
```bash
fly deploy
```

## Development

```bash
# Run locally
go run cmd/main.go

# Run tests
go test ./...
```

## Hold Period

14-day hold period protects against:
- Refunds
- Chargebacks
- Dispute resolution

After 14 days:
1. Payout status changes from `pending` → `eligible`
2. Funds move from `pending` → `available` balance
3. Funds become eligible for payout via Ruul