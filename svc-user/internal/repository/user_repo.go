package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"svc-user/internal/model"

	"github.com/videoforge/backend/pkg/errors"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found", 404)
	// ErrEmailExists is returned when the email already exists
	ErrEmailExists = errors.New("email already exists", 409)
	// ErrInvalidCredentials is returned for invalid login credentials
	ErrInvalidCredentials = errors.New("invalid credentials", 401)
	// ErrRefreshTokenNotFound is returned when refresh token is not found
	ErrRefreshTokenNotFound = errors.New("refresh token not found", 404)
	// ErrRefreshTokenExpired is returned when refresh token has expired
	ErrRefreshTokenExpired = errors.New("refresh token expired", 401)
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// UserRepoInterface defines the interface for user repository
type UserRepoInterface interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error)
	CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	// Check if email exists
	exists, err := r.emailExists(ctx, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailExists
	}

	user.ID = uuid.Must(uuid.NewV7())
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Status = "active"
	user.Role = "client"

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`
	var user model.User
	var lastLoginAt pgtype.NullTime
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`
	var user model.User
	var lastLoginAt pgtype.NullTime
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, updated_at = $3, last_login_at = $4
		WHERE id = $5
	`
	result, err := r.db.Exec(ctx, query,
		user.FirstName,
		user.LastName,
		user.UpdatedAt,
		user.LastLoginAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// GetUserRoles retrieves roles for a user
func (r *UserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	query := `
		SELECT r.id, r.name, r.description
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// emailExists checks if an email already exists
func (r *UserRepository) emailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}
	return exists, nil
}

// CreateRefreshToken creates a new refresh token
func (r *UserRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	token.ID = uuid.Must(uuid.NewV7())
	token.CreatedAt = time.Now()

	// Hash the token before storing
	hash := sha256.Sum256([]byte(token.TokenHash))
	token.TokenHash = hex.EncodeToString(hash[:])

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by hash
func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	// Hash the token for lookup
	hash := sha256.Sum256([]byte(tokenHash))
	storedHash := hex.EncodeToString(hash[:])

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var token model.RefreshToken
	err := r.db.QueryRow(ctx, query, storedHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	return &token, nil
}

// DeleteRefreshToken deletes a refresh token
func (r *UserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	hash := sha256.Sum256([]byte(tokenHash))
	storedHash := hex.EncodeToString(hash[:])

	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	result, err := r.db.Exec(ctx, query, storedHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

// DeleteUserRefreshTokens deletes all refresh tokens for a user
func (r *UserRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}

	return nil
}