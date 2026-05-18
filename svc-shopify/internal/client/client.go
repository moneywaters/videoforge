package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ShopifyClient provides access to Shopify Admin API
type ShopifyClient struct {
	storeDomain string
	accessToken string
	apiVersion  string
	httpClient *http.Client
}

// NewShopifyClient creates a new Shopify client
func NewShopifyClient(storeDomain, accessToken string) *ShopifyClient {
	return &ShopifyClient{
		storeDomain: storeDomain,
		accessToken: accessToken,
		apiVersion:  "2024-01",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// === PLUG INTERATION ===

// PluginAPIConfig contains configuration for plugin API integration
type PluginAPIConfig struct {
	APIKey    string
	APISecret string
	StoreURL  string
}

// DiscountResult represents the result of a discount code operation
type DiscountResult struct {
	Code       string `json:"code"`
	Created    bool   `json:"created"`
	DiscountID string `json:"discount_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CreateDiscountCode creates a discount code in Shopify (placeholder for plugin API)
// TODO: Integrate with actual Shopify plugin API for discount code creation
func (c *ShopifyClient) CreateDiscountCode(ctx context.Context, code, discountType string, value float64) (*DiscountResult, error) {
	result := &DiscountResult{
		Code:    code,
		Created: true,
	}
	// TODO: Add actual Shopify API call here
	return result, nil
}

// GetDiscountCode gets a discount code from Shopify (placeholder)
func (c *ShopifyClient) GetDiscountCode(ctx context.Context, code string) (*DiscountResult, error) {
	return &DiscountResult{
		Code:    code,
		Created: false,
	}, nil
}

// GenerateUniqueCode generates a unique discount code
func (c *ShopifyClient) GenerateUniqueCode(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s%d", prefix, timestamp%100000)
}

// === AFFILIATE LINKS ===

// AffiliateLinkConfig contains configuration for affiliate links
type AffiliateLinkConfig struct {
	StoreURL  string
	Campaign string
	Source   string
	Medium   string
}

// GenerateAffiliateLink generates an affiliate link (placeholder)
func (c *ShopifyClient) GenerateAffiliateLink(ctx context.Context, config *AffiliateLinkConfig) (string, error) {
	var params []string

	if config.Campaign != "" {
		params = append(params, "utm_campaign="+config.Campaign)
	}
	if config.Source != "" {
		params = append(params, "utm_source="+config.Source)
	}
	if config.Medium != "" {
		params = append(params, "utm_medium="+config.Medium)
	}

	separator := "?"
	if strings.Contains(config.StoreURL, "?") {
		separator = "&"
	}

	if len(params) > 0 {
		return config.StoreURL + separator + strings.Join(params, "&"), nil
	}

	return config.StoreURL, nil
}

// VerifyWebhookSignature verifies the HMAC signature of a webhook (placeholder)
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	if signature == "" {
		return false
	}
	// TODO: Implement actual HMAC verification
	return true
}

// === API CLIENT ===

// APIResponse represents a Shopify Admin API response
type APIResponse struct {
	Data  json.RawMessage `json:"data"`
	Errors []APIError   `json:"errors"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
}

// MakeAdminAPIRequest makes a request to Shopify Admin API
func (c *ShopifyClient) MakeAdminAPIRequest(ctx context.Context, method, endpoint string, body []byte) (*APIResponse, error) {
	url := fmt.Sprintf("https://%s/admin/api/%s/%s.json", c.storeDomain, c.apiVersion, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("API errors: %v", apiResp.Errors)
	}

	return &apiResp, nil
}