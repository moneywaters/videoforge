module svc-user

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/videoforge/backend v0.0.0
	golang.org/x/crypto v0.21.0
)

replace github.com/videoforge/backend => ../backend