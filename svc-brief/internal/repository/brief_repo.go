package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"svc-brief/internal/model"

	"github.com/videoforge/backend/pkg/errors"
)

// Error constants
var (
	ErrBriefNotFound = errors.New("brief not found", 0)
)

// BriefRepo implements brief data access
type BriefRepo struct {
	db *pgxpool.Pool
}

// BriefRepoInterface defines the brief repository interface
type BriefRepoInterface interface {
	CreateBrief(ctx context.Context, brief *model.Brief) error
	GetBriefByID(ctx context.Context, id uuid.UUID) (*model.Brief, error)
	UpdateBrief(ctx context.Context, brief *model.Brief) error
	DeleteBrief(ctx context.Context, id uuid.UUID) error
	ListBriefs(ctx context.Context, clientID *uuid.UUID, status *string, tags []string, page, limit int) (*model.ListBriefsResponse, error)
	PublishBrief(ctx context.Context, id uuid.UUID) error
	CloseBrief(ctx context.Context, id uuid.UUID) error

	GetBriefTags(ctx context.Context, briefID uuid.UUID) ([]string, error)
	SetBriefTags(ctx context.Context, briefID uuid.UUID, tags []string) error

	CreateBriefQuestion(ctx context.Context, q *model.BriefQuestion) error
	GetBriefQuestions(ctx context.Context, briefID uuid.UUID) ([]model.BriefQuestion, error)

	MarkBriefViewed(ctx context.Context, briefID, editorID uuid.UUID) error
	GetBriefViewers(ctx context.Context, briefID uuid.UUID) ([]uuid.UUID, error)
	HasViewedBrief(ctx context.Context, briefID, editorID uuid.UUID) (bool, error)

	GetMatchingBriefs(ctx context.Context, tags []string, page, limit int) (*model.ListBriefsResponse, error)
}

// NewBriefRepo creates a new BriefRepo
func NewBriefRepo(db *pgxpool.Pool) *BriefRepo {
	return &BriefRepo{db: db}
}

// CreateBrief creates a new brief
func (r *BriefRepo) CreateBrief(ctx context.Context, brief *model.Brief) error {
	if brief.ID == uuid.Nil {
		brief.ID = uuid.Must(uuid.NewV7())
	}

	query := `
		INSERT INTO brief.briefs (
			id, client_id, title, description, goals, target_audience,
			tone, style_preferences, cta, status, bounty_budget, bounty_deposited,
			submissions_limit, is_blind, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err := r.db.Exec(ctx, query,
		brief.ID,
		brief.ClientID,
		brief.Title,
		brief.Description,
		brief.Goals,
		brief.TargetAudience,
		brief.Tone,
		brief.StylePreferences,
		brief.CTA,
		brief.Status,
		brief.BountyBudget,
		brief.BountyDeposited,
		brief.SubmissionsLimit,
		brief.IsBlind,
		brief.CreatedAt,
		brief.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create brief: %w", err)
	}

	// Set tags if provided
	if len(brief.Tags) > 0 {
		if err := r.SetBriefTags(ctx, brief.ID, brief.Tags); err != nil {
			return err
		}
	}

	return nil
}

// GetBriefByID gets a brief by ID
func (r *BriefRepo) GetBriefByID(ctx context.Context, id uuid.UUID) (*model.Brief, error) {
	query := `
		SELECT id, client_id, title, description, goals, target_audience,
			tone, style_preferences, cta, status, bounty_budget, bounty_deposited,
			submissions_limit, is_blind, created_at, updated_at
		FROM brief.briefs
		WHERE id = $1
	`

	var brief model.Brief
	err := r.db.QueryRow(ctx, query, id).Scan(
		&brief.ID,
		&brief.ClientID,
		&brief.Title,
		&brief.Description,
		&brief.Goals,
		&brief.TargetAudience,
		&brief.Tone,
		&brief.StylePreferences,
		&brief.CTA,
		&brief.Status,
		&brief.BountyBudget,
		&brief.BountyDeposited,
		&brief.SubmissionsLimit,
		&brief.IsBlind,
		&brief.CreatedAt,
		&brief.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrBriefNotFound
		}
		return nil, fmt.Errorf("failed to get brief: %w", err)
	}

	// Get tags
	tags, err := r.GetBriefTags(ctx, id)
	if err != nil {
		return nil, err
	}
	brief.Tags = tags

	return &brief, nil
}

// UpdateBrief updates a brief
func (r *BriefRepo) UpdateBrief(ctx context.Context, brief *model.Brief) error {
	query := `
		UPDATE brief.briefs SET
			title = $2,
			description = $3,
			goals = $4,
			target_audience = $5,
			tone = $6,
			style_preferences = $7,
			cta = $8,
			bounty_budget = $9,
			submissions_limit = $10,
			is_blind = $11,
			updated_at = $12
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		brief.ID,
		brief.Title,
		brief.Description,
		brief.Goals,
		brief.TargetAudience,
		brief.Tone,
		brief.StylePreferences,
		brief.CTA,
		brief.BountyBudget,
		brief.SubmissionsLimit,
		brief.IsBlind,
		brief.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update brief: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrBriefNotFound
	}

	// Update tags if provided
	if len(brief.Tags) > 0 {
		if err := r.SetBriefTags(ctx, brief.ID, brief.Tags); err != nil {
			return err
		}
	}

	return nil
}

// DeleteBrief deletes a brief
func (r *BriefRepo) DeleteBrief(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM brief.briefs WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete brief: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrBriefNotFound
	}

	return nil
}

// buildListBriefsQuery builds the query and args for listing briefs
func (r *BriefRepo) buildListBriefsQuery(clientID *uuid.UUID, status *string, tags []string, page, limit int) (string, []interface{}, int) {
	baseQuery := `
		SELECT b.id, b.client_id, b.title, b.description, b.goals, b.target_audience,
			b.tone, b.style_preferences, b.cta, b.status, b.bounty_budget, b.bounty_deposited,
			b.submissions_limit, b.is_blind, b.created_at, b.updated_at
		FROM brief.briefs b
	`
	countQuery := `SELECT COUNT(DISTINCT b.id) FROM brief.briefs b`

	var conditions []string
	var args []interface{}
	argNum := 1

	if clientID != nil {
		conditions = append(conditions, fmt.Sprintf("b.client_id = $%d", argNum))
		args = append(args, *clientID)
		argNum++
	}

	if status != nil {
		conditions = append(conditions, fmt.Sprintf("b.status = $%d", argNum))
		args = append(args, *status)
		argNum++
	}

	if len(tags) > 0 {
		baseQuery += ` JOIN brief.brief_tags bt ON b.id = bt.brief_id`
		countQuery += ` JOIN brief.brief_tags bt ON b.id = bt.brief_id`
		tagArgs := make([]interface{}, len(tags))
		placeholders := make([]string, len(tags))
		for i, tag := range tags {
			tagArgs[i] = tag
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			argNum++
		}
		args = append(args, tagArgs...)
		conditions = append(conditions, fmt.Sprintf("bt.tag IN ("+joinStrings(placeholders)+")"))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + joinStrings(conditions)
	}

	offset := (page - 1) * limit
	query := baseQuery + whereClause + fmt.Sprintf(" ORDER BY b.created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)
	countQuery += whereClause

	return countQuery, query, 0
}

	if status != nil {
		conditions = append(conditions, fmt.Sprintf("b.status = $%d", argNum))
		args = append(args, *status)
		argNum++
	}

	if len(tags) > 0 {
		baseQuery += ` JOIN brief.brief_tags bt ON b.id = bt.brief_id`
		countQuery += ` JOIN brief.brief_tags bt ON b.id = bt.brief_id`
		placeholders := make([]string, len(tags))
		for i, tag := range tags {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, tag)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("bt.tag IN (%s)", joinStrings(placeholders)))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + joinStrings(conditions)
	}

	offset := (page - 1) * limit
	query := baseQuery + whereClause + fmt.Sprintf(" ORDER BY b.created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	countQuery += whereClause

	return countQuery + "; " + query, args, offset
}

// joinStrings joins strings with commas
func joinStrings(s []string) string {
	result := ""
	for i, str := range s {
		if i > 0 {
			result += ", "
		}
		result += str
	}
	return result
}

// ListBriefs lists briefs with filtering and pagination
func (r *BriefRepo) ListBriefs(ctx context.Context, clientID *uuid.UUID, status *string, tags []string, page, limit int) (*model.ListBriefsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Build the main query
	countQuery, query, _ := r.buildListBriefsQuery(clientID, status, tags, page, limit)

	// Execute main query - last 2 args are limit and offset
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list briefs: %w", err)
	}
	defer rows.Close()

	var briefs []model.Brief
	for rows.Next() {
		var brief model.Brief
		err := rows.Scan(
			&brief.ID,
			&brief.ClientID,
			&brief.Title,
			&brief.Description,
			&brief.Goals,
			&brief.TargetAudience,
			&brief.Tone,
			&brief.StylePreferences,
			&brief.CTA,
			&brief.Status,
			&brief.BountyBudget,
			&brief.BountyDeposited,
			&brief.SubmissionsLimit,
			&brief.IsBlind,
			&brief.CreatedAt,
			&brief.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan brief: %w", err)
		}
		briefs = append(briefs, brief)
	}

	// Get tags for each brief
	for i := range briefs {
		tags, err := r.GetBriefTags(ctx, briefs[i].ID)
		if err != nil {
			return nil, err
		}
		briefs[i].Tags = tags
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	totalPages := (total + limit - 1) / limit

	responses := make([]model.BriefResponse, len(briefs))
	for i, b := range briefs {
		responses[i] = b.ToResponse()
	}

	return &model.ListBriefsResponse{
		Briefs:     responses,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// PublishBrief publishes a brief
func (r *BriefRepo) PublishBrief(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE brief.briefs SET
			status = 'published',
			updated_at = NOW()
		WHERE id = $1 AND status = 'draft'
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to publish brief: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrBriefNotFound
	}

	return nil
}

// CloseBrief closes a brief
func (r *BriefRepo) CloseBrief(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE brief.briefs SET
			status = 'closed',
			updated_at = NOW()
		WHERE id = $1 AND status = 'published'
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to close brief: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrBriefNotFound
	}

	return nil
}

// GetBriefTags gets tags for a brief
func (r *BriefRepo) GetBriefTags(ctx context.Context, briefID uuid.UUID) ([]string, error) {
	query := `SELECT tag FROM brief.brief_tags WHERE brief_id = $1`

	rows, err := r.db.Query(ctx, query, briefID)
	if err != nil {
		return nil, fmt.Errorf("failed to get brief tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// SetBriefTags sets tags for a brief
func (r *BriefRepo) SetBriefTags(ctx context.Context, briefID uuid.UUID, tags []string) error {
	// Delete existing tags
	if _, err := r.db.Exec(ctx, `DELETE FROM brief.brief_tags WHERE brief_id = $1`, briefID); err != nil {
		return fmt.Errorf("failed to delete brief tags: %w", err)
	}

	// Insert new tags
	for _, tag := range tags {
		if _, err := r.db.Exec(ctx, `
			INSERT INTO brief.brief_tags (id, brief_id, tag, created_at)
			VALUES ($1, $2, $3, NOW())
		`, uuid.Must(uuid.NewV7()), briefID, tag); err != nil {
			return fmt.Errorf("failed to insert brief tag: %w", err)
		}
	}

	return nil
}

// CreateBriefQuestion creates a brief question
func (r *BriefRepo) CreateBriefQuestion(ctx context.Context, q *model.BriefQuestion) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.Must(uuid.NewV7())
	}

	query := `
		INSERT INTO brief.brief_questions (id, brief_id, question, answer, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query, q.ID, q.BriefID, q.Question, q.Answer, q.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create brief question: %w", err)
	}

	return nil
}

// GetBriefQuestions gets questions for a brief
func (r *BriefRepo) GetBriefQuestions(ctx context.Context, briefID uuid.UUID) ([]model.BriefQuestion, error) {
	query := `SELECT id, brief_id, question, answer, created_at FROM brief.brief_questions WHERE brief_id = $1`

	rows, err := r.db.Query(ctx, query, briefID)
	if err != nil {
		return nil, fmt.Errorf("failed to get brief questions: %w", err)
	}
	defer rows.Close()

	var questions []model.BriefQuestion
	for rows.Next() {
		var q model.BriefQuestion
		if err := rows.Scan(&q.ID, &q.BriefID, &q.Question, &q.Answer, &q.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}
		questions = append(questions, q)
	}

	return questions, nil
}

// MarkBriefViewed marks a brief as viewed by an editor
func (r *BriefRepo) MarkBriefViewed(ctx context.Context, briefID, editorID uuid.UUID) error {
	query := `
		INSERT INTO brief.brief_editor_views (brief_id, editor_id, viewed_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (brief_id, editor_id) DO UPDATE SET viewed_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, briefID, editorID)
	if err != nil {
		return fmt.Errorf("failed to mark brief viewed: %w", err)
	}

	return nil
}

// GetBriefViewers gets editors who viewed a brief
func (r *BriefRepo) GetBriefViewers(ctx context.Context, briefID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT editor_id FROM brief.brief_editor_views WHERE brief_id = $1`

	rows, err := r.db.Query(ctx, query, briefID)
	if err != nil {
		return nil, fmt.Errorf("failed to get brief viewers: %w", err)
	}
	defer rows.Close()

	var viewers []uuid.UUID
	for rows.Next() {
		var editorID uuid.UUID
		if err := rows.Scan(&editorID); err != nil {
			return nil, fmt.Errorf("failed to scan viewer: %w", err)
		}
		viewers = append(viewers, editorID)
	}

	return viewers, nil
}

// HasViewedBrief checks if an editor has viewed a brief
func (r *BriefRepo) HasViewedBrief(ctx context.Context, briefID, editorID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM brief.brief_editor_views WHERE brief_id = $1 AND editor_id = $2`

	var exists int
	err := r.db.QueryRow(ctx, query, briefID, editorID).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check brief viewed: %w", err)
	}

	return true, nil
}

// GetMatchingBriefs gets briefs matching given tags
func (r *BriefRepo) GetMatchingBriefs(ctx context.Context, tags []string, page, limit int) (*model.ListBriefsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Only return published briefs
	status := model.BriefStatusPublished
	return r.ListBriefs(ctx, nil, &status, tags, page, limit)
}