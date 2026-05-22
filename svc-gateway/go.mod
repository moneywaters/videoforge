module github.com/videoforge/backend/svc-gateway

go 1.25

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/videoforge/backend/pkg v0.0.0
	golang.org/x/time v0.15.0
)

require (
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
)

replace github.com/videoforge/backend/pkg => ../pkg
