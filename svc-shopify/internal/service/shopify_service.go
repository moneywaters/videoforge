package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/natsclient"
	"github.com/videoforge/backend/svc-shopify/internal/model"
	"github.com/videoforge/backend/svc-shopify/internal/repository"
)

// ShopifyService defines the interface for Shopify business logic
type ShopifyService interface {
	// Link operations
	CreateLink(ctx context.Context, req *model.CreateLinkRequest) (*model.VideoLink, error)
	GetLink(ctx context.Context, id string) (*model.VideoLink, error)
	ListLinks(ctx context.Context, videoID, campaignID string) ([]model.VideoLink, error)

	// Order webhook operations
	ProcessWebhook(ctx context.Context, payload []byte, shopDomain string) (*model.Order, error)
	GetOrder(ctx context.Context, id string) (*model.Order, error)
	ListOrders(ctx context.Context, storeID, videoID, status string) ([]model.Order, error)

	// Attribution operations
	GetAttributions(ctx context.Context, videoID, campaignID string) ([]model.Attribution, error)
	GetAttributionSummaries(ctx context.Context) ([]model.AttributionSummary, error)
}

// PgShopifyService implements ShopifyService using PostgreSQL and NATS
type PgShopifyService struct {
	repo     repository.ShopifyRepository
	nats     *natsclient.Client
	log      *logger.Logger
	storeURL string
}

// NewShopifyService creates a new Shopify service
func NewShopifyService(repo repository.ShopifyRepository, natsClient *natsclient.Client, log *logger.Logger, storeURL string) *PgShopifyService {
	return &PgShopifyService{
		repo:     repo,
		nats:     natsClient,
		log:      log,
		storeURL: storeURL,
	}
}

// =============================================================================
// Link Operations
// =============================================================================

// CreateLink creates a new custom link for a video
func (s *PgShopifyService) CreateLink(ctx context.Context, req *model.CreateLinkRequest) (*model.VideoLink, error) {
	// Validate request
	if req.VideoID == "" {
		return nil, fmt.Errorf("video_id is required")
	}
	if req.DiscountCode == "" {
		return nil, fmt.Errorf("discount_code is required")
	}
	if req.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	// Build the URL with UTM parameters
	linkURL := s.buildLinkURL(req.BaseURL, req)

	// Create the link record
	link := &model.VideoLink{
		VideoID:       req.VideoID,
		CampaignID:   nullString(req.CampaignID),
		DiscountCode: req.DiscountCode,
		UTMSource:    nullString(req.UTMSource),
		UTMMedium:    nullString(req.UTMMedium),
		UTMCampaign:  nullString(req.UTMCampaign),
		URL:          linkURL,
	}

	// Save to database
	if err := s.repo.CreateLink(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	s.log.Info("Created video link",
		slog.String("link_id", link.ID),
		slog.String("video_id", link.VideoID),
		slog.String("discount_code", link.DiscountCode),
	)

	return link, nil
}

// GetLink retrieves a link by ID
func (s *PgShopifyService) GetLink(ctx context.Context, id string) (*model.VideoLink, error) {
	link, err := s.repo.GetLinkByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil {
		return nil, fmt.Errorf("link not found")
	}
	return link, nil
}

// ListLinks lists links with optional filters
func (s *PgShopifyService) ListLinks(ctx context.Context, videoID, campaignID string) ([]model.VideoLink, error) {
	var links []model.VideoLink
	var err error

	if videoID != "" {
		links, err = s.repo.GetLinksByVideo(ctx, videoID)
	} else if campaignID != "" {
		links, err = s.repo.GetLinksByCampaign(ctx, campaignID)
	} else {
		links, err = s.repo.GetAllLinks(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list links: %w", err)
	}

	return links, nil
}

// =============================================================================
// Webhook Processing
// =============================================================================

// ProcessWebhook processes a Shopify order webhook
func (s *PgShopifyService) ProcessWebhook(ctx context.Context, payload []byte, shopDomain string) (*model.Order, error) {
	// Parse the webhook payload
	parsed, err := model.ParseOrderPayload(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Get or create the store
	store, err := s.getOrCreateStore(ctx, shopDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	// Parse the total price
	totalPrice, err := parsePrice(parsed.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total price: %w", err)
	}

	// Determine order status based on financial status
	status := s.determineOrderStatus(parsed.FinancialStatus)

	// Create the order record
	order := &model.Order{
		ShopifyOrderID: fmt.Sprintf("%d", parsed.ID),
		StoreID:       store.ID,
		CustomerEmail: nullString(parsed.Email),
		TotalPrice:    totalPrice,
		Currency:     parsed.Currency,
		DiscountCode: nullString(parsed.DiscountCode),
		UTMSource:    nullString(parsed.UTMSource),
		UTMMedium:   nullString(parsed.UTMMedium),
		UTMCampaign: nullString(parsed.UTMCampaign),
		Status:      status,
	}

	// Save to database
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	s.log.Info("Processed order webhook",
		slog.String("order_id", order.ID),
		slog.String("shopify_order_id", order.ShopifyOrderID),
		slog.Float64("total_price", order.TotalPrice),
	)

	// Try to attribute the order
	go func() {
		if err := s.attributeOrder(context.Background(), order); err != nil {
			s.log.Error("Failed to attribute order",
				slog.String("order_id", order.ID),
				slog.String("error", err.Error()))
		}
	}()

	return order, nil
}

// attributeOrder attempts to attribute an order to a video
func (s *PgShopifyService) attributeOrder(ctx context.Context, order *model.Order) error {
	var matchedLink *model.VideoLink
	var attributionMethod string

	// First, try to match by discount code
	if order.DiscountCode != nil && *order.DiscountCode != "" {
		link, err := s.repo.GetLinkByDiscountCode(ctx, *order.DiscountCode)
		if err == nil && link != nil {
			matchedLink = link
			attributionMethod = "discount_code"
		}
	}

	// If not matched, try to match by UTM parameters
	if matchedLink == nil {
		if utmSource := order.UTMSource; utmSource != nil && *utmSource != "" {
			links, err := s.repo.GetLinksByUTMSource(ctx, *utmSource)
			if err == nil && len(links) > 0 {
				matchedLink = &links[0]
				attributionMethod = "utm"
			}
		}
	}

	// If still not matched, try UTM campaign
	if matchedLink == nil {
		if utmCampaign := order.UTMCampaign; utmCampaign != nil && *utmCampaign != "" {
			links, err := s.repo.GetLinksByUTMCampaign(ctx, *utmCampaign)
			if err == nil && len(links) > 0 {
				matchedLink = &links[0]
				attributionMethod = "utm"
			}
		}
	}

	if matchedLink == nil {
		return fmt.Errorf("no matching link found for attribution")
	}

	// Create attribution record
	attr := &model.Attribution{
		OrderID:           order.ID,
		VideoID:           matchedLink.VideoID,
		CampaignID:        matchedLink.CampaignID,
		AttributedAmount:  order.TotalPrice,
		AttributionMethod: attributionMethod,
	}

	if err := s.repo.CreateAttribution(ctx, attr); err != nil {
		return fmt.Errorf("failed to create attribution: %w", err)
	}

	s.log.Info("Attributed order to video",
		slog.String("order_id", order.ID),
		slog.String("video_id", matchedLink.VideoID),
		slog.String("method", attributionMethod),
		slog.Float64("amount", order.TotalPrice),
	)

	// Emit NATS event
	s.emitSaleAttributedEvent(ctx, order, matchedLink, attributionMethod)

	return nil
}

// emitSaleAttributedEvent emits a NATS event for sale attribution
func (s *PgShopifyService) emitSaleAttributedEvent(ctx context.Context, order *model.Order, link *model.VideoLink, method string) {
	if s.nats == nil {
		return
	}

	event := model.SaleAttributedEvent{
		OrderID:            order.ID,
		VideoID:            link.VideoID,
		CampaignID:         link.CampaignID,
		Amount:            order.TotalPrice,
		AttributionMethod: method,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		s.log.Error("Failed to marshal sale attributed event",
			slog.String("error", err.Error()),
		)
		return
	}

	if err := s.nats.Publish("sale.attributed", eventJSON); err != nil {
		s.log.Error("Failed to publish sale.attributed event",
			slog.String("error", err.Error()),
		)
	}
}

// =============================================================================
// Order Operations
// =============================================================================

// GetOrder retrieves an order by ID
func (s *PgShopifyService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

// ListOrders lists orders with optional filters
func (s *PgShopifyService) ListOrders(ctx context.Context, storeID, videoID, status string) ([]model.Order, error) {
	var orders []model.Order
	var err error

	if storeID != "" {
		orders, err = s.repo.GetOrdersByStore(ctx, storeID)
	} else if videoID != "" {
		orders, err = s.repo.GetOrdersByVideo(ctx, videoID)
	} else if status != "" {
		orders, err = s.repo.GetOrdersByStatus(ctx, status)
	} else {
		orders, err = s.repo.GetAllOrders(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, nil
}

// =============================================================================
// Attribution Operations
// =============================================================================

// GetAttributions retrieves attributions with optional filters
func (s *PgShopifyService) GetAttributions(ctx context.Context, videoID, campaignID string) ([]model.Attribution, error) {
	var attributions []model.Attribution
	var err error

	if videoID != "" {
		attributions, err = s.repo.GetAttributionsByVideo(ctx, videoID)
	} else if campaignID != "" {
		attributions, err = s.repo.GetAttributionsByCampaign(ctx, campaignID)
	} else {
		attributions, err = s.repo.GetAllAttributions(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list attributions: %w", err)
	}

	return attributions, nil
}

// GetAttributionSummaries retrieves attribution summaries
func (s *PgShopifyService) GetAttributionSummaries(ctx context.Context) ([]model.AttributionSummary, error) {
	summaries, err := s.repo.GetAllAttributionSummaries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribution summaries: %w", err)
	}
	return summaries, nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// getOrCreateStore gets or creates a store for the given domain
func (s *PgShopifyService) getOrCreateStore(ctx context.Context, domain string) (*model.ShopifyStore, error) {
	store, err := s.repo.GetStoreByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if store != nil {
		return store, nil
	}

	// Create a new store (for public webhook - client_id will be nil/empty)
	store = &model.ShopifyStore{
		ClientID:    "",
		ShopDomain: domain,
		AccessToken: "",
		Status:     "active",
	}

	if err := s.repo.CreateStore(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// buildLinkURL builds the full URL with UTM parameters
func (s *PgShopifyService) buildLinkURL(baseURL string, req *model.CreateLinkRequest) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	// Add discount code as query param
	q := u.Query()
	q.Set("discount_code", req.DiscountCode)

	// Add UTM parameters if provided
	if req.UTMSource != "" {
		q.Set("utm_source", req.UTMSource)
	}
	if req.UTMMedium != "" {
		q.Set("utm_medium", req.UTMMedium)
	}
	if req.UTMCampaign != "" {
		q.Set("utm_campaign", req.UTMCampaign)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// parsePrice parses a price string to float64
func parsePrice(priceStr string) (float64, error) {
	priceStr = strings.TrimSpace(priceStr)
	if priceStr == "" {
		return 0, nil
	}

	// Handle various formats (remove currency symbols and commas)
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	priceStr = strings.ReplaceAll(priceStr, "$", "")
	priceStr = strings.ReplaceAll(priceStr, "USD", "")
	priceStr = strings.TrimSpace(priceStr)

	// Simple parsing
	var result float64
	_, err := fmt.Sscanf(priceStr, "%f", &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// determineOrderStatus determines the order status based on financial status
func (s *PgShopifyService) determineOrderStatus(financialStatus string) string {
	switch strings.ToLower(financialStatus) {
	case "paid", "captured", "succeeded":
		return "confirmed"
	case "refunded", "voided", "cancelled":
		return "cancelled"
	default:
		return "pending"
	}
}

// nullString returns a pointer to the string if non-empty, otherwise nil
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Ensure PgShopifyService implements ShopifyService
var _ ShopifyService = (*PgShopifyService)(nil)