package handler

import (
	"encoding/json"
	"net/http"

	"github.com/videoforge/backend/pkg/auth"
	"svc-user/internal/model"
	"svc-user/internal/service"

	"github.com/videoforge/backend/pkg/errors"
)

// NeonAuthHandler handles Neon Auth-specific authentication endpoints.
// It provides an optional integration with Neon Auth that falls back to
// the existing JWT-based authentication system.
type NeonAuthHandler struct {
	neonClient *auth.NeonAuthClient
	fallback   service.AuthServiceInterface
}

// NewNeonAuthHandler creates a new Neon Auth handler.
// The fallback service is used when Neon Auth is not configured.
func NewNeonAuthHandler(neonClient *auth.NeonAuthClient, fallback service.AuthServiceInterface) *NeonAuthHandler {
	return &NeonAuthHandler{
		neonClient: neonClient,
		fallback:   fallback,
	}
}

// Register handles user registration via Neon Auth.
// If Neon Auth is configured, it proxies to Neon Auth API.
// Otherwise, it falls back to the existing registration flow.
func (h *NeonAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("email and password are required"))
		return
	}

	// If Neon Auth is configured, use it
	if h.neonClient != nil && h.neonClient.IsConfigured() {
		h.registerWithNeon(w, r, req)
		return
	}

	// Fall back to existing auth system
	h.fallbackRegister(w, r, req)
}

// registerWithNeon registers a user using Neon Auth API.
func (h *NeonAuthHandler) registerWithNeon(w http.ResponseWriter, r *http.Request, req model.RegisterRequest) {
	neonReq := auth.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:    req.FirstName + " " + req.LastName,
		Provider: "email",
	}

	resp, err := h.neonClient.Register(neonReq)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.InternalServerError("registration failed"))
		return
	}

	// Convert Neon response to our response format
	response := model.AuthResponse{
		AccessToken:   resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User: model.UserResponse{
			Email:     resp.User.Email,
			FirstName: resp.User.Name,
		},
	}

	respond(w, http.StatusCreated, response)
}

// fallbackRegister registers a user using the existing auth service.
func (h *NeonAuthHandler) fallbackRegister(w http.ResponseWriter, r *http.Request, req model.RegisterRequest) {
	user, err := h.fallback.Register(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusCreated, user)
}

// Login handles user login via Neon Auth.
// If Neon Auth is configured, it proxies to Neon Auth API.
// Otherwise, it falls back to the existing login flow.
func (h *NeonAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("email and password are required"))
		return
	}

	// If Neon Auth is configured, use it
	if h.neonClient != nil && h.neonClient.IsConfigured() {
		h.loginWithNeon(w, r, req)
		return
	}

	// Fall back to existing auth system
	h.fallbackLogin(w, r, req)
}

// loginWithNeon logs in a user using Neon Auth API.
func (h *NeonAuthHandler) loginWithNeon(w http.ResponseWriter, r *http.Request, req model.LoginRequest) {
	neonReq := auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		Provider: "email",
	}

	resp, err := h.neonClient.Login(neonReq)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid credentials"))
		return
	}

	// Convert Neon response to our response format
	response := model.AuthResponse{
		AccessToken:   resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User: model.UserResponse{
			Email:  resp.User.Email,
			Name:  resp.User.Name,
		},
	}

	respond(w, http.StatusOK, response)
}

// fallbackLogin logs in a user using the existing auth service.
func (h *NeonAuthHandler) fallbackLogin(w http.ResponseWriter, r *http.Request, req model.LoginRequest) {
	response, err := h.fallback.Login(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, response)
}

// GetSession retrieves the current session.
// This proxies to Neon Auth session endpoint if configured.
func (h *NeonAuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// Get session token from Authorization header or cookie
	sessionToken := extractBearerToken(r.Header.Get("Authorization"))
	if sessionToken == "" {
		sessionToken = extractCookieToken(r, "neon_session")
	}

	if sessionToken == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("session token required"))
		return
	}

	// If Neon Auth is configured, use it
	if h.neonClient != nil && h.neonClient.IsConfigured() {
		h.getSessionFromNeon(w, r, sessionToken)
		return
	}

	// Fall back to existing session validation
	h.fallbackGetSession(w, r, sessionToken)
}

// getSessionFromNeon retrieves session from Neon Auth API.
func (h *NeonAuthHandler) getSessionFromNeon(w http.ResponseWriter, r *http.Request, sessionToken string) {
	resp, err := h.neonClient.GetSession(sessionToken)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid or expired session"))
		return
	}

	response := map[string]interface{}{
		"session": resp.Session,
		"user": map[string]string{
			"id":    resp.User.ID,
			"email": resp.User.Email,
			"name":  resp.User.Name,
		},
	}

	respond(w, http.StatusOK, response)
}

// fallbackGetSession validates session using the existing auth service.
func (h *NeonAuthHandler) fallbackGetSession(w http.ResponseWriter, r *http.Request, token string) {
	// Validate the token using the fallback service
	if token == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	// For fallback, we would typically call the auth service to validate
	// For now, return unauthorized since we need a proper token
	errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
	return
}

// Logout handles user logout via Neon Auth.
// If Neon Auth is configured, it invalidates the session via Neon Auth API.
func (h *NeonAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req model.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	// If Neon Auth is configured, use it
	if h.neonClient != nil && h.neonClient.IsConfigured() {
		h.logoutFromNeon(w, r, req)
		return
	}

	// Fall back to existing logout
	h.fallbackLogout(w, r, req)
}

// logoutFromNeon logs out user from Neon Auth.
func (h *NeonAuthHandler) logoutFromNeon(w http.ResponseWriter, r *http.Request, req model.LogoutRequest) {
	neonReq := auth.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}

	if err := h.neonClient.Logout(neonReq); err != nil {
		// Log the error but don't fail the logout
		// This ensures graceful degradation
	}

	respond(w, http.StatusNoContent, nil)
}

// fallbackLogout logs out user from the existing auth system.
func (h *NeonAuthHandler) fallbackLogout(w http.ResponseWriter, r *http.Request, req model.LogoutRequest) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	if err := h.fallback.Logout(r.Context(), nil, req.RefreshToken); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusNoContent, nil)
}

// Refresh handles token refresh.
// This endpoints allows refreshing access tokens via refresh token.
func (h *NeonAuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	if req.RefreshToken == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("refresh token is required"))
		return
	}

	// If Neon Auth is configured, use it
	if h.neonClient != nil && h.neonClient.IsConfigured() {
		h.refreshWithNeon(w, r, req.RefreshToken)
		return
	}

	// Fall back to existing refresh
	h.fallbackRefresh(w, r, req.RefreshToken)
}

// refreshWithNeon refreshes token using Neon Auth.
func (h *NeonAuthHandler) refreshWithNeon(w http.ResponseWriter, r *http.Request, refreshToken string) {
	resp, err := h.neonClient.RefreshToken(refreshToken)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid or expired refresh token"))
		return
	}

	response := model.AuthResponse{
		AccessToken:   resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User: model.UserResponse{
			Email: resp.User.Email,
		},
	}

	respond(w, http.StatusOK, response)
}

// fallbackRefresh refreshes token using the existing auth service.
func (h *NeonAuthHandler) fallbackRefresh(w http.ResponseWriter, r *http.Request, refreshToken string) {
	response, err := h.fallback.Refresh(r.Context(), refreshToken)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, response)
}

// extractBearerToken extracts token from Authorization header.
func extractBearerToken(header string) string {
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}

// extractCookieToken extracts token from a cookie.
func extractCookieToken(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// respond writes a JSON response.
func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}