package service

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/videoforge/backend/svc-user/internal/model"
	"github.com/videoforge/backend/svc-user/internal/repository"

	"github.com/videoforge/backend/pkg/errors"
)

var (
	// ErrInvalidPassword is returned when password is invalid
	ErrInvalidPassword = errors.Unauthorized("invalid credentials")
	// ErrUserBanned is returned when user is banned
	ErrUserBanned = errors.Forbidden("user account is banned")
	// ErrWeakPassword is returned when password is too weak
	ErrWeakPassword = errors.BadRequest("password must be at least 8 characters")
)

// UserService handles user business logic
type UserService struct {
	repo      repository.UserRepoInterface
	privateKey *rsa.PrivateKey
	keyID     string
}

// UserServiceInterface defines the interface for user service
type UserServiceInterface interface {
	Register(ctx context.Context, req model.RegisterRequest) (model.UserResponse, error)
	Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (model.AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (model.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.UserResponse, error)
}

// NewUserService creates a new UserService
func NewUserService(repo repository.UserRepoInterface, privateKey *rsa.PrivateKey, keyID string) *UserService {
	return &UserService{
		repo:       repo,
		privateKey: privateKey,
		keyID:      keyID,
	}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, req model.RegisterRequest) (model.UserResponse, error) {
	// Validate email format
	if req.Email == "" {
		return model.UserResponse{}, errors.BadRequest("email is required")
	}

	// Validate password strength
	if len(req.Password) < 8 {
		return model.UserResponse{}, ErrWeakPassword
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
func (s *UserService) Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return model.AuthResponse{}, ErrInvalidPassword
		}
		return model.AuthResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return model.AuthResponse{}, ErrInvalidPassword
	}

	// Check if user is banned
	if user.Status == "banned" {
		return model.AuthResponse{}, ErrUserBanned
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
func (s *UserService) Refresh(ctx context.Context, refreshToken string) (model.AuthResponse, error) {
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
		return model.AuthResponse{}, ErrUserBanned
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
func (s *UserService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	// Delete the refresh token
	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		if err == repository.ErrRefreshTokenNotFound {
			return nil // Already deleted, consider it success
		}
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// GetProfile returns the user profile
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (model.UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return model.UserResponse{}, errors.NotFound("user not found")
		}
		return model.UserResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToResponse(), nil
}

// UpdateProfile updates the user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) (model.UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return model.UserResponse{}, errors.NotFound("user not found")
		}
		return model.UserResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return model.UserResponse{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user.ToResponse(), nil
}

// generateJWT generates a new JWT access token using RS256
func (s *UserService) generateJWT(userID uuid.UUID, role string) (string, error) {
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
func (s *UserService) generateRefreshToken() (string, error) {
	token := uuid.Must(uuid.NewV7())
	return token.String(), nil
}