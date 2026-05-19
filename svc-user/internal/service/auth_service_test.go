package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/videoforge/backend/svc-user/internal/model"
	"github.com/videoforge/backend/svc-user/internal/repository"
)

// =============================================================================
// Mock Repository
// =============================================================================

// MockUserRepository is a mock implementation of UserRepoInterface for testing
type MockUserRepository struct {
	users           map[string]*model.User
	refreshTokens    map[string]*model.RefreshToken
	createUserFunc  func(ctx context.Context, user *model.User) error
	getUserByEmailFunc func(ctx context.Context, email string) (*model.User, error)
	getUserByIDFunc func(ctx context.Context, id uuid.UUID) (*model.User, error)
	updateUserFunc func(ctx context.Context, user *model.User) error
	updateLastLoginFunc func(ctx context.Context, userID uuid.UUID) error
	createRefreshTokenFunc func(ctx context.Context, token *model.RefreshToken) error
	getRefreshTokenFunc func(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	deleteRefreshTokenFunc func(ctx context.Context, tokenHash string) error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:        make(map[string]*model.User),
		refreshTokens: make(map[string]*model.RefreshToken),
	}
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	if _, ok := m.users[user.Email]; ok {
		return repository.ErrEmailExists
	}
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Status = "active"
	user.Role = "client"
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	user, ok := m.users[email]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(ctx, user)
	}
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	if m.updateLastLoginFunc != nil {
		return m.updateLastLoginFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	return nil, nil
}

func (m *MockUserRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	if m.createRefreshTokenFunc != nil {
		return m.createRefreshTokenFunc(ctx, token)
	}
	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	m.refreshTokens[token.TokenHash] = token
	return nil
}

func (m *MockUserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	if m.getRefreshTokenFunc != nil {
		return m.getRefreshTokenFunc(ctx, tokenHash)
	}
	token, ok := m.refreshTokens[tokenHash]
	if !ok {
		return nil, repository.ErrRefreshTokenNotFound
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, repository.ErrRefreshTokenExpired
	}
	return token, nil
}

func (m *MockUserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	if m.deleteRefreshTokenFunc != nil {
		return m.deleteRefreshTokenFunc(ctx, tokenHash)
	}
	if _, ok := m.refreshTokens[tokenHash]; !ok {
		return repository.ErrRefreshTokenNotFound
	}
	delete(m.refreshTokens, tokenHash)
	return nil
}

func (m *MockUserRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// Helper to generate RSA key for testing
func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return key
}

// =============================================================================
// Auth Service - Register Tests
// =============================================================================

func TestAuthService_Register_ValidInput(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	req := model.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Act
	user, err := svc.Register(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}
	if user.FirstName != req.FirstName {
		t.Errorf("expected first name %s, got %s", req.FirstName, user.FirstName)
	}
	if user.LastName != req.LastName {
		t.Errorf("expected last name %s, got %s", req.LastName, user.LastName)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	req := model.RegisterRequest{
		Email:     "",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Act
	_, err := svc.Register(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty email, got nil")
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	req1 := model.RegisterRequest{
		Email:     "duplicate@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Create first user
	_, err := svc.Register(context.Background(), req1)
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// Try to create duplicate
	req2 := model.RegisterRequest{
		Email:     "duplicate@example.com",
		Password:  "password456",
		FirstName: "Jane",
		LastName:  "Doe",
	}

	// Act
	_, err = svc.Register(context.Background(), req2)

	// Assert
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	req := model.RegisterRequest{
		Email:     "test@example.com",
		Password: "short", // Less than 8 characters
		FirstName: "John",
		LastName:  "Doe",
	}

	// Act
	_, err := svc.Register(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("expected error for weak password, got nil")
	}
}

// =============================================================================
// Auth Service - Login Tests
// =============================================================================

func TestAuthService_Login_ValidCredentials(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	// Create user first
	registerReq := model.RegisterRequest{
		Email:     "login@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}

	// Act
	resp, err := svc.Login(context.Background(), loginReq)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token, got empty string")
	}
	if resp.RefreshToken == "" {
		t.Error("expected refresh token, got empty string")
	}
	if resp.User.Email != registerReq.Email {
		t.Errorf("expected user email %s, got %s", registerReq.Email, resp.User.Email)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	// Create user first
	registerReq := model.RegisterRequest{
		Email:     "invalidpass@example.com",
		Password: "correctpassword",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "invalidpass@example.com",
		Password: "wrongpassword",
	}

	// Act
	_, err = svc.Login(context.Background(), loginReq)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid password, got nil")
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	loginReq := model.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	// Act
	_, err := svc.Login(context.Background(), loginReq)

	// Assert
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestAuthService_Login_BannedUser(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	// Create user first and mark as banned
	registerReq := model.RegisterRequest{
		Email:     "banned@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Manually set status to banned
	user, _ := repo.GetUserByEmail(context.Background(), "banned@example.com")
	user.Status = "banned"
	repo.UpdateUser(context.Background(), user)

	loginReq := model.LoginRequest{
		Email:    "banned@example.com",
		Password: "password123",
	}

	// Act
	_, err = svc.Login(context.Background(), loginReq)

	// Assert
	if err == nil {
		t.Fatal("expected error for banned user, got nil")
	}
}

// =============================================================================
// Auth Service - Refresh Tests
// =============================================================================

func TestAuthService_Refresh_ValidToken(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	// Create user and login to get refresh token
	registerReq := model.RegisterRequest{
		Email:     "refresh@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "refresh@example.com",
		Password: "password123",
	}
	loginResp, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	refreshToken := loginResp.RefreshToken

	// Act
	resp, err := svc.Refresh(context.Background(), refreshToken)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected new access token, got empty string")
	}
	if resp.RefreshToken == "" {
		t.Error("expected new refresh token, got empty string")
	}
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	// Act
	_, err := svc.Refresh(context.Background(), "invalid-token")

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestAuthService_Refresh_ExpiredToken(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	// Create a refresh token that's already expired
	expiredToken := &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "expired-token-hash",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	repo.CreateRefreshToken(context.Background(), expiredToken)

	// Act
	_, err := svc.Refresh(context.Background(), expiredToken.TokenHash)

	// Assert
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

// =============================================================================
// Auth Service - Logout Tests
// =============================================================================

func TestAuthService_Logout_TokenInvalidated(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	// Create user and login to get refresh token
	registerReq := model.RegisterRequest{
		Email:     "logout@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "logout@example.com",
		Password: "password123",
	}
	loginResp, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	user, _ := repo.GetUserByEmail(context.Background(), "logout@example.com")
	refreshToken := loginResp.RefreshToken

	// Act
	err = svc.Logout(context.Background(), user.ID, refreshToken)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify token is invalidated (should fail to refresh)
	_, err = svc.Refresh(context.Background(), refreshToken)
	if err == nil {
		t.Error("expected error after logout, token should be invalidated")
	}
}

// =============================================================================
// Password Hashing Tests
// =============================================================================

func TestAuthService_PasswordHashing_UsesBcrypt(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewAuthService(repo, privateKey, "test-key-id")

	password := "testpassword123"

	req := model.RegisterRequest{
		Email:     "hash@example.com",
		Password: password,
		FirstName: "John",
		LastName:  "Doe",
	}

	// Act
	user, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// Get the stored user
	storedUser, err := repo.GetUserByEmail(context.Background(), "hash@example.com")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	// Assert - verify bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.PasswordHash), []byte(password))
	if err != nil {
		t.Errorf("expected password to match bcrypt hash, got error: %v", err)
	}

	// Verify the hash is different from the original password
	if storedUser.PasswordHash == password {
		t.Error("expected hashed password, got plain text")
	}

	// Verify it's a valid bcrypt hash (starts with $2a$, $2b$, or $2y$)
	if len(storedUser.PasswordHash) < 60 {
		t.Errorf("expected bcrypt hash length >= 60, got %d", len(storedUser.PasswordHash))
	}

	_ = user // Silence unused variable warning
}