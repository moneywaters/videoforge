package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/svc-video/internal/model"
)

// VideoRepository defines the interface for video data access
type VideoRepository interface {
	Create(ctx context.Context, v *model.Video) error
	GetByID(ctx context.Context, id string) (*model.Video, error)
	Update(ctx context.Context, v *model.Video) error
	Delete(ctx context.Context, id string) error

	List(ctx context.Context, briefID, editorID, status string, page, pageSize int) ([]*model.Video, int, error)
	ListByBriefForEditor(ctx context.Context, briefID, editorID string, page, pageSize int) ([]*model.Video, int, error)

	CreateRevision(ctx context.Context, r *model.VideoRevision) error
	GetRevisionByID(ctx context.Context, id string) (*model.VideoRevision, error)
	ListRevisions(ctx context.Context, videoID string) ([]*model.VideoRevision, error)
	GetLatestRevisionNumber(ctx context.Context, videoID string) (int, error)

	CreateApproval(ctx context.Context, a *model.VideoApproval) error

	CreateFeedback(ctx context.Context, f *model.VideoFeedback) error
	ListFeedback(ctx context.Context, videoID string) ([]*model.VideoFeedback, error)

	// Brief operations (stub - in production call svc-brief)
	GetBrief(ctx context.Context, briefID string) (*model.Brief, error)
}

// videoRepo implements VideoRepository
type videoRepo struct {
	db *pgxpool.Pool
}

// NewVideoRepository creates a new video repository
func NewVideoRepository(db *pgxpool.Pool) VideoRepository {
	return &videoRepo{db: db}
}

// Create creates a new video
func (r *videoRepo) Create(ctx context.Context, v *model.Video) error {
	query := `
		INSERT INTO video.videos (
			id, brief_id, editor_id, title, description, storj_key,
			thumbnail_storj_key, file_size,
			status, current_revision_id, duration, resolution, thumbnail_url,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8,
			$9, $10, $11, $12, $13,
			$14, $15
		)`

	now := getCurrentTime()
	_, err := r.db.Exec(ctx, query,
		v.ID,
		v.BriefID,
		v.EditorID,
		v.Title,
		v.Description,
		v.StorjKey,
		v.ThumbnailStorjKey,
		v.FileSize,
		v.Status,
		v.CurrentRevisionID,
		v.Duration,
		v.Resolution,
		v.ThumbnailURL,
		now,
		now,
	)
	return err
}

// GetByID retrieves a video by ID
func (r *videoRepo) GetByID(ctx context.Context, id string) (*model.Video, error) {
	query := `
		SELECT id, brief_id, editor_id, title, description, storj_key,
		       thumbnail_storj_key, file_size,
		       status, current_revision_id, duration, resolution, thumbnail_url,
		       submitted_at, created_at, updated_at
		FROM video.videos
		WHERE id = $1`

	var v model.Video
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID,
		&v.BriefID,
		&v.EditorID,
		&v.Title,
		&v.Description,
		&v.StorjKey,
		&v.ThumbnailStorjKey,
		&v.FileSize,
		&v.Status,
		&v.CurrentRevisionID,
		&v.Duration,
		&v.Resolution,
		&v.ThumbnailURL,
		&v.SubmittedAt,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Update updates an existing video
func (r *videoRepo) Update(ctx context.Context, v *model.Video) error {
	query := `
		UPDATE video.videos SET
			brief_id = $2,
			editor_id = $3,
			title = $4,
			description = $5,
			storj_key = $6,
			thumbnail_storj_key = $7,
			file_size = $8,
			status = $9,
			current_revision_id = $10,
			duration = $11,
			resolution = $12,
			thumbnail_url = $13,
			submitted_at = $14,
			updated_at = $15
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		v.ID,
		v.BriefID,
		v.EditorID,
		v.Title,
		v.Description,
		v.StorjKey,
		v.ThumbnailStorjKey,
		v.FileSize,
		v.Status,
		v.CurrentRevisionID,
		v.Duration,
		v.Resolution,
		v.ThumbnailURL,
		v.SubmittedAt,
		getCurrentTime(),
	)
	return err
}

// Delete deletes a video by ID
func (r *videoRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM video.videos WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// List retrieves videos with optional filters and blind submission enforcement
// For editors: only returns their own videos for a brief
// For clients: returns all videos for their briefs
// For admins: returns all videos
func (r *videoRepo) List(ctx context.Context, briefID, editorID, status string, page, pageSize int) ([]*model.Video, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Build query with conditions
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if briefID != "" {
		conditions = append(conditions, fmt.Sprintf("brief_id = $%d", argIdx))
		args = append(args, briefID)
		argIdx++
	}

	// Enforce blind submissions for editors
	if editorID != "" {
		conditions = append(conditions, fmt.Sprintf("editor_id = $%d", argIdx))
		args = append(args, editorID)
		argIdx++
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM video.videos %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Main query
	mainQuery := fmt.Sprintf(`
		SELECT id, brief_id, editor_id, title, description, storj_key,
		       thumbnail_storj_key, file_size,
		       status, current_revision_id, duration, resolution, thumbnail_url,
		       submitted_at, created_at, updated_at
		FROM video.videos
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, mainQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var videos []*model.Video
	for rows.Next() {
		var v model.Video
		if err := rows.Scan(
			&v.ID,
			&v.BriefID,
			&v.EditorID,
			&v.Title,
			&v.Description,
			&v.StorjKey,
			&v.ThumbnailStorjKey,
			&v.FileSize,
			&v.Status,
			&v.CurrentRevisionID,
			&v.Duration,
			&v.Resolution,
			&v.ThumbnailURL,
			&v.SubmittedAt,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		videos = append(videos, &v)
	}

	return videos, total, nil
}

// ListByBriefForEditor retrieves videos for a brief, enforcing blind submissions
// Returns only videos from the specified editor
func (r *videoRepo) ListByBriefForEditor(ctx context.Context, briefID, editorID string, page, pageSize int) ([]*model.Video, int, error) {
	return r.List(ctx, briefID, editorID, "", page, pageSize)
}

// CreateRevision creates a new video revision
func (r *videoRepo) CreateRevision(ctx context.Context, rev *model.VideoRevision) error {
	query := `
		INSERT INTO video.video_revisions (
			id, video_id, revision_number, storj_key, changelog, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`

	_, err := r.db.Exec(ctx, query,
		rev.ID,
		rev.VideoID,
		rev.RevisionNumber,
		rev.StorjKey,
		rev.Changelog,
		getCurrentTime(),
	)
	return err
}

// GetRevisionByID retrieves a revision by ID
func (r *videoRepo) GetRevisionByID(ctx context.Context, id string) (*model.VideoRevision, error) {
	query := `
		SELECT id, video_id, revision_number, storj_key, changelog, created_at
		FROM video.video_revisions
		WHERE id = $1`

	var rev model.VideoRevision
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rev.ID,
		&rev.VideoID,
		&rev.RevisionNumber,
		&rev.StorjKey,
		&rev.Changelog,
		&rev.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rev, nil
}

// ListRevisions retrieves all revisions for a video
func (r *videoRepo) ListRevisions(ctx context.Context, videoID string) ([]*model.VideoRevision, error) {
	query := `
		SELECT id, video_id, revision_number, storj_key, changelog, created_at
		FROM video.video_revisions
		WHERE video_id = $1
		ORDER BY revision_number DESC`

	rows, err := r.db.Query(ctx, query, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revisions []*model.VideoRevision
	for rows.Next() {
		var rev model.VideoRevision
		if err := rows.Scan(
			&rev.ID,
			&rev.VideoID,
			&rev.RevisionNumber,
			&rev.StorjKey,
			&rev.Changelog,
			&rev.CreatedAt,
		); err != nil {
			return nil, err
		}
		revisions = append(revisions, &rev)
	}
	return revisions, nil
}

// GetLatestRevisionNumber gets the latest revision number for a video
func (r *videoRepo) GetLatestRevisionNumber(ctx context.Context, videoID string) (int, error) {
	query := `
		SELECT COALESCE(MAX(revision_number), 0)
		FROM video.video_revisions
		WHERE video_id = $1`

	var num int
	err := r.db.QueryRow(ctx, query, videoID).Scan(&num)
	return num, err
}

// CreateApproval creates an approval record
func (r *videoRepo) CreateApproval(ctx context.Context, a *model.VideoApproval) error {
	query := `
		INSERT INTO video.video_approvals (
			id, video_id, status, approved_by, approved_at, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`

	_, err := r.db.Exec(ctx, query,
		a.ID,
		a.VideoID,
		a.Status,
		a.ApprovedBy,
		a.ApprovedAt,
		a.Notes,
	)
	return err
}

// CreateFeedback creates feedback for a video
func (r *videoRepo) CreateFeedback(ctx context.Context, f *model.VideoFeedback) error {
	query := `
		INSERT INTO video.video_feedback (
			id, video_id, revision_id, feedback, created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`

	_, err := r.db.Exec(ctx, query,
		f.ID,
		f.VideoID,
		f.RevisionID,
		f.Feedback,
		f.CreatedBy,
		getCurrentTime(),
	)
	return err
}

// ListFeedback retrieves all feedback for a video
func (r *videoRepo) ListFeedback(ctx context.Context, videoID string) ([]*model.VideoFeedback, error) {
	query := `
		SELECT id, video_id, revision_id, feedback, created_by, created_at
		FROM video.video_feedback
		WHERE video_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbackList []*model.VideoFeedback
	for rows.Next() {
		var f model.VideoFeedback
		if err := rows.Scan(
			&f.ID,
			&f.VideoID,
			&f.RevisionID,
			&f.Feedback,
			&f.CreatedBy,
			&f.CreatedAt,
		); err != nil {
			return nil, err
		}
		feedbackList = append(feedbackList, &f)
	}
	return feedbackList, nil
}

// GetBrief retrieves a brief by ID
// This is a stub - in production, this would call svc-brief
func (r *videoRepo) GetBrief(ctx context.Context, briefID string) (*model.Brief, error) {
	// Stub: in production this would call the brief service
	// For now, return a placeholder that will fail validation
	return &model.Brief{
		ID:          briefID,
		ClientID:    "",
		Status:     "draft",
		IsPublished: false,
	}, nil
}

// Helper to generate UUID
func newID() string {
	return uuid.New().String()
}

// getCurrentTime returns current time
func getCurrentTime() time.Time {
	return time.Now()
}