package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/svc-payout/internal/model"
	"github.com/videoforge/backend/svc-payout/internal/service"
)

// PayoutHandler handles payout HTTP requests
type PayoutHandler struct {
	svc *service.PayoutService
}

// NewPayoutHandler creates a new payout handler
func NewPayoutHandler(svc *service.PayoutService) *PayoutHandler {
	return &PayoutHandler{svc: svc}
}

// GetPayouts handles GET /api/v1/payouts
func (h *PayoutHandler) GetPayouts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get current user from context (set by auth middleware)
	// For now, try to get from query param or header
	userIDStr := r.URL.Query().Get("user_id")
	var userID uuid.UUID
	var err error

	if userIDStr != "" {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			errors.Write(r.Context(), w, errors.BadRequest("invalid user_id"))
			return
		}
	}

	// If admin, return all payouts
	if userIDStr == "" {
		payouts, err := h.svc.GetAllPayouts(ctx)
		if err != nil {
			errors.Write(r.Context(), w, errors.Internal(err.Error()))
			return
		}
		success(w, r, payouts)
		return
	}

	payouts, err := h.svc.GetPayoutsForUser(ctx, userID)
	if err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}
	success(w, r, payouts)
}

// GetPayoutByID handles GET /api/v1/payouts/:id
func (h *PayoutHandler) GetPayoutByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	payoutID, ok := vars["id"]
	if !ok {
		errors.Write(r.Context(), w, errors.BadRequest("payout id required"))
		return
	}

	id, err := uuid.Parse(payoutID)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid payout id"))
		return
	}

	payout, err := h.svc.GetPayoutByID(ctx, id)
	if err != nil {
		errors.Write(r.Context(), w, errors.NotFound("payout not found"))
		return
	}

	success(w, r, payout)
}

// GetBalance handles GET /api/v1/balance
func (h *PayoutHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from query param
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		errors.Write(r.Context(), w, errors.BadRequest("user_id required"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid user_id"))
		return
	}

	balance, err := h.svc.GetBalance(ctx, userID)
	if err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	success(w, r, balance)
}

// GetEarnings handles GET /api/v1/earnings
func (h *PayoutHandler) GetEarnings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		errors.Write(r.Context(), w, errors.BadRequest("user_id required"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid user_id"))
		return
	}

	// Parse period from query params
	periodStart := r.URL.Query().Get("period_start")
	periodEnd := r.URL.Query().Get("period_end")

	startTime := time.Now().AddDate(0, -1, 0) // Default: last month
	endTime := time.Now()

	if periodStart != "" {
		startTime, err = time.Parse(time.RFC3339, periodStart)
		if err != nil {
			errors.Write(r.Context(), w, errors.BadRequest("invalid period_start format"))
			return
		}
	}

	if periodEnd != "" {
		endTime, err = time.Parse(time.RFC3339, periodEnd)
		if err != nil {
			errors.Write(r.Context(), w, errors.BadRequest("invalid period_end format"))
			return
		}
	}

	breakdown, err := h.svc.CalculateEarnings(ctx, userID, startTime, endTime)
	if err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	success(w, r, breakdown)
}

// CalculateEarnings handles POST /api/v1/payouts/calculate
func (h *PayoutHandler) CalculateEarnings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("failed to read request body"))
		return
	}
	defer r.Body.Close()

	var req model.CalculateEarningsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	if req.UserID == uuid.Nil {
		errors.Write(r.Context(), w, errors.BadRequest("user_id required"))
		return
	}

	breakdown, err := h.svc.CalculateEarnings(ctx, req.UserID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	success(w, r, breakdown)
}

// CreateBatch handles POST /api/v1/payouts/batches
func (h *PayoutHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("failed to read request body"))
		return
	}
	defer r.Body.Close()

	var req model.CreateBatchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	if len(req.UserIDs) == 0 {
		errors.Write(r.Context(), w, errors.BadRequest("user_ids required"))
		return
	}

	batch, err := h.svc.CreateRuulBatch(ctx, req.UserIDs, req.Description)
	if err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	success(w, r, batch)
}

// GetBatches handles GET /api/v1/payouts/batches
func (h *PayoutHandler) GetBatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	batches, err := h.svc.GetAllRuulBatches(ctx)
	if err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	success(w, r, batches)
}

// GetBatchByID handles GET /api/v1/payouts/batches/:id
func (h *PayoutHandler) GetBatchByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	batchID, ok := vars["id"]
	if !ok {
		errors.Write(r.Context(), w, errors.BadRequest("batch id required"))
		return
	}

	id, err := uuid.Parse(batchID)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid batch id"))
		return
	}

	batch, err := h.svc.GetRuulBatch(ctx, id)
	if err != nil {
		errors.Write(r.Context(), w, errors.NotFound("batch not found"))
		return
	}

	success(w, r, batch)
}

// ProcessBatch handles POST /api/v1/payouts/batches/:id/process
func (h *PayoutHandler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	batchID, ok := vars["id"]
	if !ok {
		errors.Write(r.Context(), w, errors.BadRequest("batch id required"))
		return
	}

	id, err := uuid.Parse(batchID)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid batch id"))
		return
	}

	batch, err := h.svc.ProcessRuulBatch(ctx, id)
	if err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	success(w, r, batch)
}

// HandleDodoWebhook handles POST /api/v1/payouts/webhook/dodo
func (h *PayoutHandler) HandleDodoWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("failed to read request body"))
		return
	}
	defer r.Body.Close()

	var event model.DodoWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid webhook body"))
		return
	}

	if err := h.svc.HandleDodoWebhook(ctx, event); err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

// HandleRuulWebhook handles POST /api/v1/payouts/webhook/ruul
func (h *PayoutHandler) HandleRuulWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("failed to read request body"))
		return
	}
	defer r.Body.Close()

	var event model.RuulWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		errors.Write(r.Context(), w, errors.BadRequest("invalid webhook body"))
		return
	}

	if err := h.svc.HandleRuulWebhook(ctx, event); err != nil {
		errors.Write(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

// success writes a success response
func success(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":   data,
	})
}

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HandleHealth handles GET /health
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"service": "svc-payout",
		"status":  "healthy",
	})
}

// AdminMiddleware checks if user is admin
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Verify admin role from JWT claims
		next.ServeHTTP(w, r)
	})
}