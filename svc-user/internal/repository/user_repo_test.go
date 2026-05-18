package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"svc-user/internal/model"
)

// =============================================================================
// In-Memory Repository for Testing
// =============================================================================

// InMemoryUserRepository is an in-memory implementation for testing
type InMemoryUserRepository struct {
	users         map[uuid.UUID]*model.User
	usersByEmail map[string]*model.User
	refreshTokens map[string]*model.RefreshToken
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:         make(map[uuid.UUID]*model.User),
		usersByEmail: make(map[string]*model.User),
		refreshTokens: make(map[string]*model.RefreshToken),
	}
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	if _, ok := r.usersByEmail[user.Email]; ok {
		return ErrEmailExists
	}
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Status = "active"
	user.Role = "client"
	r.users[user.ID] = user
	r.usersByEmail[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, ok := r.usersByEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	r.usersByEmail[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	return nil, nil
}

func (r *InMemoryUserRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	r.refreshTokens[token.TokenHash] = token
	return nil
}

func (r *InMemoryUserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	token, ok := r.refreshTokens[tokenHash]
	if !ok {
		return nil, ErrRefreshTokenNotFound
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}
	return token, nil
}

func (r *InMemoryUserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	if _, ok := r.refreshTokens[tokenHash]; !ok {
		return ErrRefreshTokenNotFound
	}
	delete(r.refreshTokens, tokenHash)
	return nil
}

func (r *InMemoryUserRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	for hash, token := range r.refreshTokens {
		if token.UserID == userID {
			delete(r.refreshTokens, hash)
		}
	}
	return nil
}

// =============================================================================
// Repository Tests - CreateUser
// =============================================================================

func TestUserRepository_CreateUser(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashedpassword",
		FirstName:    "John",
		LastName:     "Doe",
	}

	// Act
	err := repo.CreateUser(context.Background(), user)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID == uuid.Nil {
		t.Error("expected user ID to be set")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if user.Status != "active" {
		t.Errorf("expected status 'active', got %s", user.Status)
	}
	if user.Role != "client" {
		t.Errorf("expected role 'client', got %s", user.Role)
	}
}

func TestUserRepository_CreateUser_DuplicateEmail(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user1 := &model.User{
		Email:        "duplicate@example.com",
		PasswordHash: "$2a$10$hash1",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user1)
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	user2 := &model.User{
		Email:        "duplicate@example.com",
		PasswordHash: "$2a$10$hash2",
		FirstName:    "Jane",
		LastName:     "Doe",
	}

	// Act
	err = repo.CreateUser(context.Background(), user2)

	// Assert
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if err != ErrEmailExists {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

// =============================================================================
// Repository Tests - GetUserByEmail
// =============================================================================

func TestUserRepository_GetUserByEmail(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "getbyemail@example.com",
		PasswordHash: "$2a$10$hash",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Act
	foundUser, err := repo.GetUserByEmail(context.Background(), "getbyemail@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if foundUser.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, foundUser.ID)
	}
	if foundUser.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, foundUser.Email)
	}
}

func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()

	// Act
	_, err := repo.GetUserByEmail(context.Background(), "nonexistent@example.com")

	// Assert
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// =============================================================================
// Repository Tests - GetUserByID
// =============================================================================

func TestUserRepository_GetUserByID(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "getbyid@example.com",
		PasswordHash: "$2a$10$hash",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Act
	foundUser, err := repo.GetUserByID(context.Background(), user.ID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if foundUser.Email != "getbyid@example.com" {
		t.Errorf("expected email getbyid@example.com, got %s", foundUser.Email)
	}
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()

	// Act
	_, err := repo.GetUserByID(context.Background(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// =============================================================================
// Repository Tests - UpdateUser
// =============================================================================

func TestUserRepository_UpdateUser(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "update@example.com",
		PasswordHash: "$2a$10$hash",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Update user
	user.FirstName = "Jane"
	user.LastName = "Smith"

	// Act
	err = repo.UpdateUser(context.Background(), user)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify update
	updatedUser, _ := repo.GetUserByID(context.Background(), user.ID)
	if updatedUser.FirstName != "Jane" {
		t.Errorf("expected first name 'Jane', got %s", updatedUser.FirstName)
	}
	if updatedUser.LastName != "Smith" {
		t.Errorf("expected last name 'Smith', got %s", updatedUser.LastName)
	}
	if updatedUser.UpdatedAt.Before(user.CreatedAt) {
		t.Error("expected updated_at to be after created_at")
	}
}

// =============================================================================
// Repository Tests - Refresh Token Operations
// =============================================================================

func TestUserRepository_CreateRefreshToken(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "token@example.com",
		PasswordHash: "$2a$10$hash",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: "test-token-hash",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// Act
	err = repo.CreateRefreshToken(context.Background(), token)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token.ID == uuid.Nil {
		t.Error("expected token ID to be set")
	}
}

func TestUserRepository_GetRefreshToken(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "gettoken@example.com",
		PasswordHash: "$2a$10$hash",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: "token-to-get",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = repo.CreateRefreshToken(context.Background(), token)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Act
	foundToken, err := repo.GetRefreshToken(context.Background(), "token-to-get")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if foundToken.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, foundToken.UserID)
	}
}

func TestUserRepository_GetRefreshToken_Expired(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "expiredtoken@example.com",
		PasswordHash: "$2a$10$hash",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
	}
	err = repo.CreateRefreshToken(context.Background(), token)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Act
	_, err = repo.GetRefreshToken(context.Background(), "expired-token")

	// Assert
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if err != ErrRefreshTokenExpired {
		t.Errorf("expected ErrRefreshTokenExpired, got %v", err)
	}
}

func TestUserRepository_DeleteRefreshToken(t *testing.T) {
	// Arrange
	repo := NewInMemoryUserRepository()
	user := &model.User{
		Email:        "deltoken@example.com",
		PasswordHash: "$2a$10$hash",
		FirstName:    "John",
		LastName:     "Doe",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: "token-to-delete",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = repo.CreateRefreshToken(context.Background(), token)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Act
	err = repo.DeleteRefreshToken(context.Background(), "token-to-delete")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deletion
	_, err = repo.GetRefreshToken(context.Background(), "token-to-delete")
	if err != ErrRefreshTokenNotFound {
		t.Errorf("expected ErrRefreshTokenNotFound after deletion, got %v", err)
	}
}