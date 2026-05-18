module github.com/videoforge/backend/svc-notification

go 1.23

require (
	github.com/go-chi/chi/v5 v5.0.14
	github.com/gorilla/websocket v1.5.1
	github.com/jackc/pgx/v5 v5.5.5
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/kelseyhightower/envconfig v1.0.0
	github.com/nats-io/nats.go v1.34.1
	github.com/videoforge/backend/pkg v0.0.0
)

replace github.com/videoforge/backend/pkg => ../pkg