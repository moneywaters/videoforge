package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// RuulClient handles Ruul.io Business API calls
type RuulClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewRuulClient creates a new Ruul client
func NewRuulClient(apiKey, baseURL string) *RuulClient {
	return &RuulClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// CreateBulkPayout creates a bulk payout request to multiple freelancers
// TODO: Implement actual Ruul.io API call
func (c *RuulClient) CreateBulkPayout(req *RuulBulkPayoutRequest) (*RuulBulkPayoutResponse, error) {
	// TODO: Implement Ruul.io bulk payout API
	// POST /payouts/bulk
	return nil, fmt.Errorf("Ruul.io CreateBulkPayout not implemented - STUB")
}

// CreatePayoutRequest creates a single payout request
// TODO: Implement actual Ruul.io API call
func (c *RuulClient) CreatePayoutRequest(req *RuulPayoutRequest) (*RuulPayoutResponse, error) {
	// TODO: Implement Ruul.io payout API
	// POST /payouts
	return nil, fmt.Errorf("Ruul.io CreatePayoutRequest not implemented - STUB")
}

// GetPayoutStatus retrieves payout status
// TODO: Implement actual Ruul.io API call
func (c *RuulClient) GetPayoutStatus(payoutID string) (*RuulPayoutResponse, error) {
	// TODO: Implement Ruul.io status API
	// GET /payouts/{id}
	return nil, fmt.Errorf("Ruul.io GetPayoutStatus not implemented - STUB")
}

// VerifyWebhook verifies webhook signature
func (c *RuulClient) VerifyWebhook(payload []byte, signature string) bool {
	// TODO: Implement webhook signature verification
	return true
}

// RuulBulkPayoutRequest represents bulk payout request
type RuulBulkPayoutRequest struct {
	BatchName  string              `json:"batch_name"`
	Payouts    []RuulPayoutItem    `json:"payouts"`
	Currency  string              `json:"currency"`
	Metadata  map[string]string    `json:"metadata,omitempty"`
}

// RuulPayoutItem represents a single payout in bulk request
type RuulPayoutItem struct {
	UserID   uuid.UUID `json:"user_id"`
	Amount   float64   `json:"amount"`
	Currency string   `json:"currency"`
}

// RuulBulkPayoutResponse represents bulk payout response
type RuulBulkPayoutResponse struct {
	BatchID       string                 `json:"batch_id"`
	Status       string                 `json:"status"`
	Payouts      []RuulPayoutResponse    `json:"payouts"`
	TotalAmount  float64                `json:"total_amount"`
	Currency     string                 `json:"currency"`
}

// RuulPayoutRequest represents single payout request
type RuulPayoutRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	Amount   float64   `json:"amount"`
	Currency string   `json:"currency"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RuulPayoutResponse represents payout response
type RuulPayoutResponse struct {
	ID        string  `json:"id"`
	Status   string  `json:"status"`
	UserID   uuid.UUID `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// post sends a POST request to Ruul API
func (c *RuulClient) post(endpoint string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Ruul.io API error: status %d", resp.StatusCode)
	}

	return nil, nil
}

// get sends a GET request to Ruul API
func (c *RuulClient) get(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ruul.io API error: status %d", resp.StatusCode)
	}

	return nil, nil
}