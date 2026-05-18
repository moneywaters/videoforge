package model

import (
	"time"

	"github.com/google/uuid"
)

// Brief status constants
const (
	BriefStatusDraft     = "draft"
	BriefStatusPublished = "published"
	BriefStatusClosed   = "closed"
)

// Brief represents a video brief
type Brief struct {
	ID               uuid.UUID  `json:"id"`
	ClientID         uuid.UUID  `json:"client_id"`
	Title            string    `json:"title"`
	Description     string    `json:"description"`
	Goals            string    `json:"goals"`
	TargetAudience   string    `json:"target_audience"`
	Tone             string    `json:"tone"`
	StylePreferences string    `json:"style_preferences"`
	CTA              string    `json:"cta"`
	Status           string    `json:"status"`
	BountyBudget     float64   `json:"bounty_budget"`
	BountyDeposited  bool      `json:"bounty_deposited"`
	SubmissionsLimit int       `json:"submissions_limit"`
	IsBlind          bool      `json:"is_blind"`
	Tags             []string  `json:"tags,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// BriefTag represents a brief tag
type BriefTag struct {
	ID        uuid.UUID `json:"id"`
	BriefID   uuid.UUID `json:"brief_id"`
	Tag      string    `json:"tag"`
	CreatedAt time.Time `json:"created_at"`
}

// BriefQuestion represents an AI interview Q&A
type BriefQuestion struct {
	ID        uuid.UUID `json:"id"`
	BriefID   uuid.UUID `json:"brief_id"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
}

// BriefEditorView tracks editor views
type BriefEditorView struct {
	BriefID  uuid.UUID `json:"brief_id"`
	EditorID uuid.UUID `json:"editor_id"`
	ViewedAt time.Time `json:"viewed_at"`
}

// CreateBriefRequest represents a request to create a brief
type CreateBriefRequest struct {
	Title            string  `json:"title"`
	Description     string  `json:"description"`
	Goals           string  `json:"goals"`
	TargetAudience string  `json:"target_audience"`
	Tone            string  `json:"tone"`
	StylePreferences string `json:"style_preferences"`
	CTA             string  `json:"cta"`
	BountyBudget    float64 `json:"bounty_budget"`
	SubmissionsLimit int    `json:"submissions_limit"`
	IsBlind         bool    `json:"is_blind"`
	Tags            []string `json:"tags,omitempty"`
}

// UpdateBriefRequest represents a request to update a brief
type UpdateBriefRequest struct {
	Title            string  `json:"title,omitempty"`
	Description     string  `json:"description,omitempty"`
	Goals           string  `json:"goals,omitempty"`
	TargetAudience string  `json:"target_audience,omitempty"`
	Tone            string  `json:"tone,omitempty"`
	StylePreferences string `json:"style_preferences,omitempty"`
	CTA             string  `json:"cta,omitempty"`
	BountyBudget    *float64 `json:"bounty_budget,omitempty"`
	SubmissionsLimit *int    `json:"submissions_limit,omitempty"`
	IsBlind         *bool   `json:"is_blind,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

// BriefResponse represents a brief response
type BriefResponse struct {
	ID               uuid.UUID `json:"id"`
	ClientID         uuid.UUID `json:"client_id"`
	Title            string    `json:"title"`
	Description     string    `json:"description"`
	Goals            string    `json:"goals"`
	TargetAudience   string    `json:"target_audience"`
	Tone             string    `json:"tone"`
	StylePreferences string    `json:"style_preferences"`
	CTA              string    `json:"cta"`
	Status           string    `json:"status"`
	BountyBudget     float64   `json:"bounty_budget"`
	BountyDeposited  bool      `json:"bounty_deposited"`
	SubmissionsLimit int       `json:"submissions_limit"`
	IsBlind          bool      `json:"is_blind"`
	Tags             []string  `json:"tags"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToResponse converts a Brief to BriefResponse
func (b *Brief) ToResponse() BriefResponse {
	return BriefResponse{
		ID:               b.ID,
		ClientID:         b.ClientID,
		Title:            b.Title,
		Description:      b.Description,
		Goals:            b.Goals,
		TargetAudience:   b.TargetAudience,
		Tone:             b.Tone,
		StylePreferences: b.StylePreferences,
		CTA:              b.CTA,
		Status:           b.Status,
		BountyBudget:     b.BountyBudget,
		BountyDeposited:  b.BountyDeposited,
		SubmissionsLimit: b.SubmissionsLimit,
		IsBlind:          b.IsBlind,
		Tags:             b.Tags,
		CreatedAt:        b.CreatedAt,
		UpdatedAt:        b.UpdatedAt,
	}
}

// ListBriefsResponse represents a paginated list of briefs
type ListBriefsResponse struct {
	Briefs   []BriefResponse `json:"briefs"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	Total   int             `json:"total"`
	TotalPages int          `json:"total_pages"`
}

// InterviewRequest represents an AI interview request
type InterviewRequest struct {
	Message string `json:"message"`
}

// StructuredBrief represents a structured brief from AI interview
type StructuredBrief struct {
	Goals            string `json:"goals"`
	TargetAudience string `json:"target_audience"`
	Tone           string `json:"tone"`
	StylePreferences string `json:"style_preferences"`
	CTA            string `json:"cta"`
}

// InterviewResponse represents an AI interview response
type InterviewResponse struct {
	Message         string          `json:"message"`
	StructuredBrief StructuredBrief `json:"structured_brief"`
}

// PublishBriefRequest represents a request to publish a brief
type PublishBriefRequest struct {
	BountyDeposited bool `json:"bounty_deposited"`
}

// ViewBriefRequest represents a request to mark a brief as viewed
type ViewBriefRequest struct {
	EditorID string `json:"editor_id"`
}

// MatchingBriefsRequest represents a request to find matching briefs
type MatchingBriefsRequest struct {
	EditorTags []string `json:"editor_tags"`
}