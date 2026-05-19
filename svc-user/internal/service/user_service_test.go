package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/videoforge/backend/svc-user/internal/model"
	"github.com/videoforge/backend/svc-user/internal/repository"
)

// =============================================================================
// User Service - GetProfile Tests
// =============================================================================

func TestUserService_GetProfile_ValidID(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Create a user first
	req := model.RegisterRequest{
		Email:     "profile@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	createdUser, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Act
	user, err := svc.GetProfile(context.Background(), createdUser.ID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != createdUser.ID {
		t.Errorf("expected ID %s, got %s", createdUser.ID, user.ID)
	}
	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}
}

func TestUserService_GetProfile_InvalidID(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Act
	_, err := svc.GetProfile(context.Background(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid user ID, got nil")
	}
}

// =============================================================================
// User Service - UpdateProfile Tests
// =============================================================================

func TestUserService_UpdateProfile_UpdatesAllowedFields(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Create a user first
	req := model.RegisterRequest{
		Email:     "update@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	createdUser, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	updateReq := model.UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
	}

	// Act
	updatedUser, err := svc.UpdateProfile(context.Background(), createdUser.ID, updateReq)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedUser.FirstName != "Jane" {
		t.Errorf("expected first name 'Jane', got %s", updatedUser.FirstName)
	}
	if updatedUser.LastName != "Smith" {
		t.Errorf("expected last name 'Smith', got %s", updatedUser.LastName)
	}
}

func TestUserService_UpdateProfile_InvalidUserID(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	updateReq := model.UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
	}

	// Act
	_, err := svc.UpdateProfile(context.Background(), uuid.New(), updateReq)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid user ID, got nil")
	}
}

func TestUserService_UpdateProfile_PartialUpdate(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Create a user first
	req := model.RegisterRequest{
		Email:     "partial@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	createdUser, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Only update first name
	updateReq := model.UpdateProfileRequest{
		FirstName: "Jane",
	}

	// Act
	updatedUser, err := svc.UpdateProfile(context.Background(), createdUser.ID, updateReq)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedUser.FirstName != "Jane" {
		t.Errorf("expected first name 'Jane', got %s", updatedUser.FirstName)
	}
	// LastName should remain unchanged
	if updatedUser.LastName != "Doe" {
		t.Errorf("expected last name 'Doe', got %s", updatedUser.LastName)
	}
}

// =============================================================================
// Additional User Service Tests
// =============================================================================

func TestUserService_Login_ValidCredentials(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Create user first
	registerReq := model.RegisterRequest{
		Email:     "userlogin@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "userlogin@example.com",
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
	if resp.User.Email != registerReq.Email {
		t.Errorf("expected user email %s, got %s", registerReq.Email, resp.User.Email)
	}
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Create user first
	registerReq := model.RegisterRequest{
		Email:     "wrongpass@example.com",
		Password: "correctpassword",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "wrongpass@example.com",
		Password: "wrongpassword",
	}

	// Act
	_, err = svc.Login(context.Background(), loginReq)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid password, got nil")
	}
}

func TestUserService_Refresh_ValidToken(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Create user and login to get refresh token
	registerReq := model.RegisterRequest{
		Email:     "userrefresh@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "userrefresh@example.com",
		Password: "password123",
	}
	loginResp, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// Act
	resp, err := svc.Refresh(context.Background(), loginResp.RefreshToken)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected new access token, got empty string")
	}
}

func TestUserService_Logout_TokenInvalidated(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	privateKey := generateTestRSAKey(t)
	svc := NewUserService(repo, privateKey, "test-key-id")

	// Create user and login
	registerReq := model.RegisterRequest{
		Email:     "userlogout@example.com",
		Password: "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	_, err := svc.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := model.LoginRequest{
		Email:    "userlogout@example.com",
		Password: "password123",
	}
	loginResp, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	user, _ := repo.GetUserByEmail(context.Background(), "userlogout@example.com")

	// Act
	err = svc.Logout(context.Background(), user.ID, loginResp.RefreshToken)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify token is invalidated
	_, err = svc.Refresh(context.Background(), loginResp.RefreshToken)
	if err == nil {
		t.Error("expected error after logout, token should be invalidated")
	}
}