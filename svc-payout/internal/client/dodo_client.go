package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// DodoClient handles DodoPayments API calls
type DodoClient struct {
	apiKey  string
	secret string
	baseURL string
	httpClient *http.Client
}

// NewDodoClient creates a new DodoPayments client
func NewDodoClient(apiKey, secret, baseURL string) *DodoClient {
	return &DodoClient{
		apiKey:   apiKey,
		secret:  secret,
		baseURL: baseURL,
		httpClient: &http.Client{},
	}
}

// CreateCheckout creates a checkout session for payment
// TODO: Implement actual DodoPayments API call
func (c *DodoClient) CreateCheckout(req interface{}) (interface{}, error) {
	// TODO: Implement DodoPayments checkout API
	// POST /checkout/sessions
	return nil, fmt.Errorf("DodoPayments CreateCheckout not implemented - STUB")
}

// GetCheckoutStatus retrieves checkout status
// TODO: Implement actual DodoPayments API call
func (c *DodoClient) GetCheckoutStatus(checkoutID string) (interface{}, error) {
	// TODO: Implement DodoPayments status API
	// GET /checkout/sessions/{id}
	return nil, fmt.Errorf("DodoPayments GetCheckoutStatus not implemented - STUB")
}

// CreatePayout creates a payout to the business account
// TODO: Implement actual DodoPayments API call
func (c *DodoClient) CreatePayout(req interface{}) (interface{}, error) {
	// TODO: Implement DodoPayments payout API
	// Note: DodoPayments pays out to the business (VideoForge), not individual freelancers
	return nil, fmt.Errorf("DodoPayments CreatePayout not implemented - STUB")
}

// VerifyWebhook verifies webhook signature
func (c *DodoClient) VerifyWebhook(payload []byte, signature string) bool {
	// TODO: Implement webhook signature verification
	return true
}

// dodoCheckoutRequest represents DodoPayments checkout request
type dodoCheckoutRequest struct {
	Amount      float64            `json:"amount"`
	Currency   string             `json:"currency"`
	ExternalID string             `json:"external_id"`
	Metadata   map[string]string   `json:"metadata"`
}

// dodoCheckoutResponse represents DodoPayments checkout response
type dodoCheckoutResponse struct {
	ID        string `json:"id"`
	Status   string `json:"status"`
	Amount   float64 `json:"amount"`
	Currency string `json:"currency"`
}

// MarshalJSON marshals request to JSON
func (r *dodoCheckoutRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (c *DodoClient) post(endpoint string, body interface{}) ([]byte, error) {
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
		return nil, fmt.Errorf("DodoPayments API error: status %d", resp.StatusCode)
	}

	return nil, nil
}