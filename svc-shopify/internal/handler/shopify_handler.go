package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/svc-shopify/internal/model"
	"github.com/videoforge/backend/svc-shopify/internal/service"
)

// ShopifyHandler handles HTTP requests for Shopify service
type ShopifyHandler struct {
	service service.ShopifyService
	log    *logger.Logger
}

// NewShopifyHandler creates a new Shopify handler
func NewShopifyHandler(svc service.ShopifyService, log *logger.Logger) *ShopifyHandler {
	return &ShopifyHandler{
		service: svc,
		log:     log,
	}
}

// Routes registers the handler routes
func (h *ShopifyHandler) Routes(mux *http.ServeMux) {
	// Webhook endpoint (public)
	mux.HandleFunc("POST /api/v1/shopify/webhook", h.HandleWebhook)

	// Link endpoints (JWT required)
	mux.HandleFunc("POST /api/v1/links", h.HandleCreateLink)
	mux.HandleFunc("GET /api/v1/links", h.HandleListLinks)
	mux.HandleFunc("GET /api/v1/links/", h.HandleGetLink)

	// Order endpoints (JWT required)
	mux.HandleFunc("GET /api/v1/orders", h.HandleListOrders)
	mux.HandleFunc("GET /api/v1/orders/", h.HandleGetOrder)

	// Attribution endpoints (JWT required)
	mux.HandleFunc("GET /api/v1/attributions", h.HandleListAttributions)
	mux.HandleFunc("GET /api/v1/attributions/summary", h.HandleGetAttributionSummary)
}

// =============================================================================
// Webhook Handler
// =============================================================================

// HandleWebhook handles Shopify webhook requests
func (h *ShopifyHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("Failed to read request body", slog.String("error", err.Error()))
		errors.WriteError(ctx, w, errors.BadRequest("Failed to read request body"))
		return
	}

	// Get shop domain from header or query param
	shopDomain := r.Header.Get("X-Shopify-Shop-Domain")
	if shopDomain == "" {
		shopDomain = r.URL.Query().Get("shop")
	}
	if shopDomain == "" {
		// Try to extract from body
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err == nil {
			if shop, ok := payload["shop_domain"].(string); ok {
				shopDomain = shop
			} else if shop, ok := payload["shop"].(string); ok {
				shopDomain = shop
			}
		}
	}

	// TODO: Add Shopify HMAC verification here
	// For MVP, we accept the payload without verification

	// Process the webhook
	order, err := h.service.ProcessWebhook(ctx, body, shopDomain)
	if err != nil {
		h.log.Error("Failed to process webhook", slog.String("error", err.Error()))
		errors.WriteError(ctx, w, errors.Internal(fmt.Sprintf("failed to process webhook: %v", err)))
		return
	}

	// Return the created order
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order.ToOrderResponse())
}

// =============================================================================
// Link Handlers
// =============================================================================

// HandleCreateLink handles POST /api/v1/links
func (h *ShopifyHandler) HandleCreateLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate request method
	if r.Method != http.MethodPost {
		errors.WriteError(ctx, w, errors.BadRequest("method not allowed"))
		return
	}

	// Parse request body
	var req model.CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("Failed to parse request body", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Validate required fields
	if req.VideoID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("video_id is required"))
		return
	}
	if req.DiscountCode == "" {
		errors.WriteError(ctx, w, errors.BadRequest("discount_code is required"))
		return
	}
	if req.BaseURL == "" {
		errors.WriteError(ctx, w, errors.BadRequest("base_url is required"))
		return
	}

	// Create the link
	link, err := h.service.CreateLink(ctx, &req)
	if err != nil {
		h.log.Error("Failed to create link", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.Internal(fmt.Sprintf("failed to create link: %v", err)))
		return
	}

	// Return the created link
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link.ToLinkResponse())
}

// HandleListLinks handles GET /api/v1/links
func (h *ShopifyHandler) HandleListLinks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate request method
	if r.Method != http.MethodGet {
		errors.WriteError(ctx, w, errors.BadRequest("method not allowed"))
		return
	}

	// Parse query parameters
	videoID := r.URL.Query().Get("video_id")
	campaignID := r.URL.Query().Get("campaign_id")

	// List links
	links, err := h.service.ListLinks(ctx, videoID, campaignID)
	if err != nil {
		h.log.Error("Failed to list links", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.Internal(fmt.Sprintf("failed to list links: %v", err)))
		return
	}

	// Convert to response format
	responses := make([]model.LinkResponse, len(links))
	for i, link := range links {
		responses[i] = link.ToLinkResponse()
	}

	// Return the links
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// HandleGetLink handles GET /api/v1/links/:id
func (h *ShopifyHandler) HandleGetLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate request method
	if r.Method != http.MethodGet {
		errors.WriteError(ctx, w, errors.BadRequest("method not allowed"))
		return
	}

	// Extract ID from path
	id := extractID(r.URL.Path)

	// Get the link
	link, err := h.service.GetLink(ctx, id)
	if err != nil {
		h.log.Error("Failed to get link", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.NotFound(fmt.Sprintf("link not found: %v", err)))
		return
	}

	// Return the link
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(link.ToLinkResponse())
}

// =============================================================================
// Order Handlers
// =============================================================================

// HandleListOrders handles GET /api/v1/orders
func (h *ShopifyHandler) HandleListOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate request method
	if r.Method != http.MethodGet {
		errors.WriteError(ctx, w, errors.BadRequest("method not allowed"))
		return
	}

	// Parse query parameters
	storeID := r.URL.Query().Get("store_id")
	videoID := r.URL.Query().Get("video_id")
	status := r.URL.Query().Get("status")

	// List orders
	orders, err := h.service.ListOrders(ctx, storeID, videoID, status)
	if err != nil {
		h.log.Error("Failed to list orders", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.Internal(fmt.Sprintf("failed to list orders: %v", err)))
		return
	}

	// Convert to response format
	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToOrderResponse()
	}

	// Return the orders
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// HandleGetOrder handles GET /api/v1/orders/:id
func (h *ShopifyHandler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate request method
	if r.Method != http.MethodGet {
		errors.WriteError(ctx, w, errors.BadRequest("method not allowed"))
		return
	}

	// Extract ID from path
	id := extractID(r.URL.Path)

	// Get the order
	order, err := h.service.GetOrder(ctx, id)
	if err != nil {
		h.log.Error("Failed to get order", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.NotFound(fmt.Sprintf("order not found: %v", err)))
		return
	}

	// Return the order
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order.ToOrderResponse())
}

// =============================================================================
// Attribution Handlers
// =============================================================================

// HandleListAttributions handles GET /api/v1/attributions
func (h *ShopifyHandler) HandleListAttributions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate request method
	if r.Method != http.MethodGet {
		errors.WriteError(ctx, w, errors.BadRequest("method not allowed"))
		return
	}

	// Parse query parameters
	videoID := r.URL.Query().Get("video_id")
	campaignID := r.URL.Query().Get("campaign_id")

	// List attributions
	attributions, err := h.service.GetAttributions(ctx, videoID, campaignID)
	if err != nil {
		h.log.Error("Failed to list attributions", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.Internal(fmt.Sprintf("failed to list attributions: %v", err)))
		return
	}

	// Convert to response format
	responses := make([]model.AttributionResponse, len(attributions))
	for i, attr := range attributions {
		responses[i] = attr.ToAttributionResponse()
	}

	// Return the attributions
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// HandleGetAttributionSummary handles GET /api/v1/attributions/summary
func (h *ShopifyHandler) HandleGetAttributionSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate request method
	if r.Method != http.MethodGet {
		errors.WriteError(ctx, w, errors.BadRequest("method not allowed"))
		return
	}

	// Get summaries
	summaries, err := h.service.GetAttributionSummaries(ctx)
	if err != nil {
		h.log.Error("Failed to get attribution summaries", slog.String("error", err.Error())))
		errors.WriteError(ctx, w, errors.Internal(fmt.Sprintf("failed to get attribution summaries: %v", err)))
		return
	}

	// Return the summaries
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// =============================================================================
// Helper Functions
// =============================================================================

// extractID extracts the ID from the URL path
// URL format: /api/v1/resource/:id
func extractID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 5 {
		return parts[4]
	}
	// Try alternate format /api/v1/resource/id
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}