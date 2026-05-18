# VideoForge Notification Service

Real-time notification service for VideoForge backend.

## Features

- WebSocket support for real-time notifications
- NATS event consumer for video, sale, payout, and campaign events
- REST API for notification management
- User notification preferences

## Endpoints

- `GET /health` - Health check
- `GET /ws` - WebSocket endpoint
- `GET /api/v1/notifications` - List notifications
- `POST /api/v1/notifications/:id/read` - Mark as read
- `POST /api/v1/notifications/read-all` - Mark all as read
- `GET /api/v1/notifications/preferences` - Get preferences
- `PUT /api/v1/notifications/preferences` - Update preferences
- `DELETE /api/v1/notifications/:id` - Delete notification

## Environment Variables

- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection URL
- `NATS_URL` - NATS server URL (default: nats://localhost:4222)
- `JWT_SECRET` - JWT secret for authentication
- `ENVIRONMENT` - Environment (development/production)

## Events Consumed

- `video.submitted` - New video submission
- `video.approved` - Video approved
- `video.rejected` - Video rejected
- `sale.attributed` - Sale attributed
- `payout.released` - Payout released
- `campaign.started` - Campaign started
- `campaign.ended` - Campaign ended