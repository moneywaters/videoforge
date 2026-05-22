package model

// Targeting defines campaign targeting options
type Targeting struct {
	AgeMin      int      `json:"age_min"`
	AgeMax      int      `json:"age_max"`
	Gender      string   `json:"gender"` // all, male, female
	Locations   []string `json:"locations"`
	Interests   []string `json:"interests"`
	Keywords    []string `json:"keywords"`
}

// CampaignStats represents campaign performance statistics
type CampaignStats struct {
	CampaignID    string  `json:"campaign_id"`
	Impressions  int64   `json:"impressions"`
	Clicks       int64   `json:"clicks"`
	Views        int64   `json:"views"`
	Conversions  int64   `json:"conversions"`
	Spend        float64 `json:"spend"`
	CTR          float64 `json:"ctr"` // click-through rate
	CPC          float64 `json:"cpc"` // cost per click
	CPV          float64 `json:"cpv"` // cost per view
}