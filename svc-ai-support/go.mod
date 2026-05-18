module svc-ai-support

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/nats-io/nats.go v1.31.0
	github.com/videoforge/backend v0.0.0
)

replace github.com/videoforge/backend => ../backend