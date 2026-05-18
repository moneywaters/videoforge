module github.com/videoforge/backend/svc-brief

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/joho/godotenv v1.5.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/videoforge/backend/pkg v0.0.0
)

replace github.com/videoforge/backend/pkg => ../pkg