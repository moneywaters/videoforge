package model

import (
	"encoding/json"
	"time"
)

// =============================================================================
// Request/Response DTOs
// =============================================================================

// CreateLinkRequest represents the request body for creating a video link
type CreateLinkRequest struct {
	VideoID      string `json:"video_id"`
	CampaignID   string `json:"campaign_id,omitempty"`
	DiscountCode string `json:"discount_code"`
	UTMSource    string `json:"utm_source,omitempty"`
	UTMMedium    string `json:"utm_medium,omitempty"`
	UTMCampaign  string `json:"utm_campaign,omitempty"`
	BaseURL      string `json:"base_url"`
}

// LinkResponse represents a video link in responses
type LinkResponse struct {
	ID             string `json:"id"`
	VideoID        string `json:"video_id"`
	CampaignID     string `json:"campaign_id,omitempty"`
	DiscountCode  string `json:"discount_code"`
	UTMSource     string `json:"utm_source,omitempty"`
	UTMMedium    string `json:"utm_medium,omitempty"`
	UTMCampaign  string `json:"utm_campaign,omitempty"`
	URL           string `json:"url"`
	CreatedAt     string `json:"created_at"`
}

// OrderResponse represents an order in responses
type OrderResponse struct {
	ID              string  `json:"id"`
	ShopifyOrderID  string  `json:"shopify_order_id"`
	StoreID         string  `json:"store_id"`
	CustomerEmail  string  `json:"customer_email,omitempty"`
	TotalPrice     float64 `json:"total_price"`
	Currency       string  `json:"currency"`
	DiscountCode   string  `json:"discount_code,omitempty"`
	UTMSource      string  `json:"utm_source,omitempty"`
	UTMMedium     string  `json:"utm_medium,omitempty"`
	UTMCampaign   string  `json:"utm_campaign,omitempty"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// AttributionResponse represents an attribution in responses
type AttributionResponse struct {
	ID                string  `json:"id"`
	OrderID          string  `json:"order_id"`
	VideoID          string  `json:"video_id"`
	CampaignID       string  `json:"campaign_id,omitempty"`
	AttributedAmount float64 `json:"attributed_amount"`
	AttributionMethod string `json:"attribution_method"`
	CreatedAt        string  `json:"created_at"`
}

// AttributionSummary represents summary statistics for attributions
type AttributionSummary struct {
	VideoID          string  `json:"video_id"`
	CampaignID      string  `json:"campaign_id,omitempty"`
	TotalSales      float64 `json:"total_sales"`
	TotalOrders     int     `json:"total_orders"`
	AttributedAmount float64 `json:"attributed_amount"`
}

// WebhookOrderPayload represents an order received from Shopify webhook
type WebhookOrderPayload struct {
	ID               int64    `json:"id"`
	OrderNumber     int64    `json:"order_number"`
	Email           string   `json:"email"`
	TotalPrice      string   `json:"total_price"`
	Currency        string   `json:"currency"`
	DiscountCodes  []string `json:"discount_codes"`
	DiscountCode   string   `json:"discount_code"`
	UTMSource       string   `json:"utm_source"`
	UTMMedium      string   `json:"utm_medium"`
	UTMCampaign    string   `json:"utm_campaign"`
	FinancialStatus string   `json:"financial_status"`
	CreatedAt       string   `json:"created_at"`
}

// =============================================================================
// Database Models (for internal use)
// =============================================================================

// ShopifyStore represents a connected Shopify store
type ShopifyStore struct {
	ID           string    `json:"id"`
	ClientID    string    `json:"client_id"`
	ShopDomain  string    `json:"shop_domain"`
	AccessToken string    `json:"access_token"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VideoLink represents a custom link for a video
type VideoLink struct {
	ID            string    `json:"id"`
	VideoID       string    `json:"video_id"`
	CampaignID   *string   `json:"campaign_id"`
	DiscountCode string    `json:"discount_code"`
	UTMSource    *string   `json:"utm_source"`
	UTMMedium    *string   `json:"utm_medium"`
	UTMCampaign  *string   `json:"utm_campaign"`
	URL           string    `json:"url"`
	CreatedAt    time.Time `json:"created_at"`
}

// Order represents a Shopify order
type Order struct {
	ID              string    `json:"id"`
	ShopifyOrderID  string    `json:"shopify_order_id"`
	StoreID         string    `json:"store_id"`
	CustomerEmail  *string   `json:"customer_email"`
	TotalPrice     float64   `json:"total_price"`
	Currency       string    `json:"currency"`
	DiscountCode   *string   `json:"discount_code"`
	UTMSource      *string   `json:"utm_source"`
	UTMMedium     *string   `json:"utm_medium"`
	UTMCampaign   *string   `json:"utm_campaign"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Attribution represents an attribution record
type Attribution struct {
	ID                string    `json:"id"`
	OrderID          string    `json:"order_id"`
	VideoID          string    `json:"video_id"`
	CampaignID       *string   `json:"campaign_id"`
	AttributedAmount float64   `json:"attributed_amount"`
	AttributionMethod string  `json:"attribution_method"`
	CreatedAt        time.Time `json:"created_at"`
}

// =============================================================================
// NATS Event Payloads
// =============================================================================

// SaleAttributedEvent is emitted when a sale is attributed to a video
type SaleAttributedEvent struct {
	OrderID            string  `json:"order_id"`
	VideoID            string  `json:"video_id"`
	CampaignID         *string `json:"campaign_id,omitempty"`
	Amount             float64 `json:"amount"`
	AttributionMethod string  `json:"attribution_method"`
}

// =============================================================================
// Helper Methods
// =============================================================================

// ToLinkResponse converts a VideoLink to LinkResponse
func (l *VideoLink) ToLinkResponse() LinkResponse {
	resp := LinkResponse{
		ID:            l.ID,
		VideoID:       l.VideoID,
		DiscountCode:  l.DiscountCode,
		URL:            l.URL,
		CreatedAt:      l.CreatedAt.Format(time.RFC3339),
	}
	if l.CampaignID != nil {
		resp.CampaignID = *l.CampaignID
	}
	if l.UTMSource != nil {
		resp.UTMSource = *l.UTMSource
	}
	if l.UTMMedium != nil {
		resp.UTMMedium = *l.UTMMedium
	}
	if l.UTMCampaign != nil {
		resp.UTMCampaign = *l.UTMCampaign
	}
	return resp
}

// ToOrderResponse converts an Order to OrderResponse
func (o *Order) ToOrderResponse() OrderResponse {
	resp := OrderResponse{
		ID:             o.ID,
		ShopifyOrderID: o.ShopifyOrderID,
		StoreID:        o.StoreID,
		TotalPrice:     o.TotalPrice,
		Currency:       o.Currency,
		Status:         o.Status,
		CreatedAt:      o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      o.UpdatedAt.Format(time.RFC3339),
	}
	if o.CustomerEmail != nil {
		resp.CustomerEmail = *o.CustomerEmail
	}
	if o.DiscountCode != nil {
		resp.DiscountCode = *o.DiscountCode
	}
	if o.UTMSource != nil {
		resp.UTMSource = *o.UTMSource
	}
	if o.UTMMedium != nil {
		resp.UTMMedium = *o.UTMMedium
	}
	if o.UTMCampaign != nil {
		resp.UTMCampaign = *o.UTMCampaign
	}
	return resp
}

// ToAttributionResponse converts an Attribution to AttributionResponse
func (a *Attribution) ToAttributionResponse() AttributionResponse {
	resp := AttributionResponse{
		ID:                a.ID,
		OrderID:           a.OrderID,
		VideoID:           a.VideoID,
		AttributedAmount:  a.AttributedAmount,
		AttributionMethod: a.AttributionMethod,
		CreatedAt:         a.CreatedAt.Format(time.RFC3339),
	}
	if a.CampaignID != nil {
		resp.CampaignID = *a.CampaignID
	}
	return resp
}

// ParseOrderPayload parses webhook order payload from JSON
func ParseOrderPayload(data []byte) (*WebhookOrderPayload, error) {
	var payload WebhookOrderPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}