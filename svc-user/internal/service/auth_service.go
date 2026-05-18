package service

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"svc-user/internal/model"
	"svc-user/internal/repository"

	"github.com/videoforge/backend/pkg/errors"
)

// AuthService handles authentication business logic
type AuthService struct {
	repo        repository.UserRepoInterface
	privateKey  *rsa.PrivateKey
	keyID       string
}

// AuthServiceInterface defines the interface for auth service
type AuthServiceInterface interface {
	Register(ctx context.Context, req model.RegisterRequest) (model.UserResponse, error)
	Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (model.AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
}

// NewAuthService creates a new AuthService
func NewAuthService(repo repository.UserRepoInterface, privateKey *rsa.PrivateKey, keyID string) *AuthService {
	return &AuthService{
		repo:       repo,
		privateKey: privateKey,
		keyID:      keyID,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.UserResponse, error) {
	// Validate email format
	if req.Email == "" {
		return model.UserResponse{}, errors.BadRequest("email is required")
	}

	// Validate password strength
	if len(req.Password) < 8 {
		return model.UserResponse{}, errors.BadRequest("password must be at least 8 characters")
	}

	// Validate name fields
	if req.FirstName == "" || req.LastName == "" {
		return model.UserResponse{}, errors.BadRequest("first name and last name are required")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.UserResponse{}, errors.Internal("failed to process password")
	}

	// Create user
	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if err == repository.ErrEmailExists {
			return model.UserResponse{}, errors.Conflict("email already exists")
		}
		return model.UserResponse{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ToResponse(), nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return model.AuthResponse{}, errors.Unauthorized("invalid credentials")
		}
		return model.AuthResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return model.AuthResponse{}, errors.Unauthorized("invalid credentials")
	}

	// Check if user is banned
	if user.Status == "banned" {
		return model.AuthResponse{}, errors.Forbidden("user account is banned")
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to update last login: %w", err)
	}

	// Generate tokens
	accessToken, err := s.generateJWT(user.ID, user.Role)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.repo.CreateRefreshToken(ctx, &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

// Refresh refreshes the access token using a refresh token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (model.AuthResponse, error) {
	// Get refresh token from DB
	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		if err == repository.ErrRefreshTokenNotFound || err == repository.ErrRefreshTokenExpired {
			return model.AuthResponse{}, errors.Unauthorized("invalid refresh token")
		}
		return model.AuthResponse{}, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return model.AuthResponse{}, errors.NotFound("user not found")
		}
		return model.AuthResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is banned
	if user.Status == "banned" {
		return model.AuthResponse{}, errors.Forbidden("user account is banned")
	}

	// Delete old refresh token (rotation)
	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to delete old refresh token: %w", err)
	}

	// Generate new tokens
	accessToken, err := s.generateJWT(user.ID, user.Role)
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store new refresh token
	if err := s.repo.CreateRefreshToken(ctx, &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: newRefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		return model.AuthResponse{}, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user.ToResponse(),
	}, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	// Delete the refresh token
	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		if err == repository.ErrRefreshTokenNotFound {
			return nil // Already deleted, consider it success
		}
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// generateJWT generates a new JWT access token using RS256
func (s *AuthService) generateJWT(userID uuid.UUID, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"role": role,
		"iat":  now.Unix(),
		"exp":  now.Add(15 * time.Minute).Unix(),
		"iss":  "videoforge",
		"aud":  "videoforge",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID

	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// generateRefreshToken generates a new refresh token
func (s *AuthService) generateRefreshToken() (string, error) {
	token := uuid.Must(uuid.NewV7())
	return token.String(), nil
}