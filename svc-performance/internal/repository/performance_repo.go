package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/svc-performance/internal/model"
)

type PerformanceRepository struct {
	db *pgx.Conn
}

func NewPerformanceRepository(db *pgx.Conn) *PerformanceRepository {
	return &PerformanceRepository{db: db}
}

// VideoSales operations

func (r *PerformanceRepository) GetVideoSales(ctx context.Context, videoID string) (*model.VideoSales, error) {
	query := `
		SELECT id, video_id, campaign_id, total_orders, total_revenue, currency, first_sale_at, last_sale_at, updated_at
		FROM video_sales
		WHERE video_id = $1
	`

	var sales model.VideoSales
	err := r.db.QueryRow(ctx, query, videoID).Scan(
		&sales.ID, &sales.VideoID, &sales.CampaignID, &sales.TotalOrders, &sales.TotalRevenue,
		&sales.Currency, &sales.FirstSaleAt, &sales.LastSaleAt, &sales.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("video sales not found")
		}
		return nil, err
	}
	return &sales, nil
}

func (r *PerformanceRepository) UpsertVideoSales(ctx context.Context, videoID, campaignID string, amount float64, currency string) error {
	query := `
		INSERT INTO video_sales (video_id, campaign_id, total_orders, total_revenue, currency, first_sale_at, last_sale_at)
		VALUES ($1, $2, 1, $3, $4, NOW(), NOW())
		ON CONFLICT (video_id) DO UPDATE SET
			total_orders = video_sales.total_orders + 1,
			total_revenue = video_sales.total_revenue + $3,
			last_sale_at = NOW(),
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, videoID, campaignID, amount, currency)
	if err != nil {
		return err
	}
	return nil
}

// EditorSales operations

func (r *PerformanceRepository) GetEditorSales(ctx context.Context, editorID string) (*model.EditorSales, error) {
	query := `
		SELECT id, editor_id, total_videos, total_orders, total_revenue, currency, updated_at
		FROM editor_sales
		WHERE editor_id = $1
	`

	var sales model.EditorSales
	err := r.db.QueryRow(ctx, query, editorID).Scan(
		&sales.ID, &sales.EditorID, &sales.TotalVideos, &sales.TotalOrders, &sales.TotalRevenue,
		&sales.Currency, &sales.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("editor sales not found")
		}
		return nil, err
	}
	return &sales, nil
}

func (r *PerformanceRepository) UpsertEditorSales(ctx context.Context, editorID string, amount float64, currency string) error {
	query := `
		INSERT INTO editor_sales (editor_id, total_videos, total_orders, total_revenue, currency)
		VALUES ($1, 1, 1, $2, $3)
		ON CONFLICT (editor_id) DO UPDATE SET
			total_orders = editor_sales.total_orders + 1,
			total_revenue = editor_sales.total_revenue + $2,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, editorID, amount, currency)
	if err != nil {
		return err
	}
	return nil
}

// SpecialistSales operations

func (r *PerformanceRepository) GetSpecialistSales(ctx context.Context, specialistID string) (*model.SpecialistSales, error) {
	query := `
		SELECT id, specialist_id, total_campaigns, total_orders, total_revenue, currency, updated_at
		FROM specialist_sales
		WHERE specialist_id = $1
	`

	var sales model.SpecialistSales
	err := r.db.QueryRow(ctx, query, specialistID).Scan(
		&sales.ID, &sales.SpecialistID, &sales.TotalCampaigns, &sales.TotalOrders, &sales.TotalRevenue,
		&sales.Currency, &sales.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("specialist sales not found")
		}
		return nil, err
	}
	return &sales, nil
}

func (r *PerformanceRepository) UpsertSpecialistSales(ctx context.Context, specialistID string, campaignID string, amount float64, currency string) error {
	query := `
		INSERT INTO specialist_sales (specialist_id, total_campaigns, total_orders, total_revenue, currency)
		VALUES ($1, 1, 1, $2, $3)
		ON CONFLICT (specialist_id) DO UPDATE SET
			total_campaigns = specialist_sales.total_campaigns + 1,
			total_orders = specialist_sales.total_orders + 1,
			total_revenue = specialist_sales.total_revenue + $2,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, specialistID, amount, currency)
	if err != nil {
		return err
	}
	return nil
}

// CampaignSales operations

func (r *PerformanceRepository) GetCampaignSales(ctx context.Context, campaignID string) (*model.CampaignSales, error) {
	query := `
		SELECT id, campaign_id, total_orders, total_revenue, currency, start_date, end_date, updated_at
		FROM campaign_sales
		WHERE campaign_id = $1
	`

	var sales model.CampaignSales
	err := r.db.QueryRow(ctx, query, campaignID).Scan(
		&sales.ID, &sales.CampaignID, &sales.TotalOrders, &sales.TotalRevenue,
		&sales.Currency, &sales.StartDate, &sales.EndDate, &sales.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("campaign sales not found")
		}
		return nil, err
	}
	return &sales, nil
}

func (r *PerformanceRepository) UpsertCampaignSales(ctx context.Context, campaignID string, amount float64, currency string, startDate, endDate time.Time) error {
	query := `
		INSERT INTO campaign_sales (campaign_id, total_orders, total_revenue, currency, start_date, end_date)
		VALUES ($1, 1, $2, $3, $4, $5)
		ON CONFLICT (campaign_id) DO UPDATE SET
			total_orders = campaign_sales.total_orders + 1,
			total_revenue = campaign_sales.total_revenue + $2,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, campaignID, amount, currency, startDate, endDate)
	if err != nil {
		return err
	}
	return nil
}

// Leaderboard operations

func (r *PerformanceRepository) GetLeaderboard(ctx context.Context, briefID string, entityType string) ([]model.LeaderboardEntry, error) {
	query := `
		SELECT id, brief_id, entity_type, entity_id, rank, total_revenue, total_orders, updated_at
		FROM leaderboards
		WHERE brief_id = $1 AND entity_type = $2
		ORDER BY rank ASC
	`

	rows, err := r.db.Query(ctx, query, briefID, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.LeaderboardEntry
	for rows.Next() {
		var entry model.LeaderboardEntry
		err := rows.Scan(
			&entry.ID, &entry.BriefID, &entry.EntityType, &entry.EntityID,
			&entry.Rank, &entry.TotalRevenue, &entry.TotalOrders, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *PerformanceRepository) GetLeaderboardRankings(ctx context.Context, briefID string, entityType string, limit, offset int) ([]model.LeaderboardEntry, error) {
	query := `
		SELECT id, brief_id, entity_type, entity_id, rank, total_revenue, total_orders, updated_at
		FROM leaderboards
		WHERE brief_id = $1 AND entity_type = $2
		ORDER BY rank ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, briefID, entityType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.LeaderboardEntry
	for rows.Next() {
		var entry model.LeaderboardEntry
		err := rows.Scan(
			&entry.ID, &entry.BriefID, &entry.EntityType, &entry.EntityID,
			&entry.Rank, &entry.TotalRevenue, &entry.TotalOrders, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *PerformanceRepository) CalculateAndStoreLeaderboard(ctx context.Context, briefID string, entityType string) error {
	// Calculate leaderboard based on entity type
	var rankingQuery string
	switch entityType {
	case "editor":
		rankingQuery = `
			WITH ranked AS (
				SELECT 
					es.editor_id as entity_id,
					SUM(es.total_revenue) as total_revenue,
					SUM(es.total_orders) as total_orders,
					ROW_NUMBER() OVER (ORDER BY SUM(es.total_revenue) DESC) as rank
				FROM editor_sales es
				JOIN videos v ON v.editor_id = es.editor_id
				JOIN briefs b ON b.id = v.brief_id
				WHERE b.id = $1
				GROUP BY es.editor_id
			)
			INSERT INTO leaderboards (brief_id, entity_type, entity_id, rank, total_revenue, total_orders)
			SELECT $1, 'editor', entity_id, rank, total_revenue, total_orders
			FROM ranked
			ON CONFLICT (brief_id, entity_type, entity_id) DO UPDATE SET
				rank = EXCLUDED.rank,
				total_revenue = EXCLUDED.total_revenue,
				total_orders = EXCLUDED.total_orders,
				updated_at = NOW()
		`
	case "video":
		rankingQuery = `
			WITH ranked AS (
				SELECT 
					vs.video_id as entity_id,
					SUM(vs.total_revenue) as total_revenue,
					SUM(vs.total_orders) as total_orders,
					ROW_NUMBER() OVER (ORDER BY SUM(vs.total_revenue) DESC) as rank
				FROM video_sales vs
				JOIN campaigns c ON c.id = vs.campaign_id
				JOIN briefs b ON b.id = c.brief_id
				WHERE b.id = $1
				GROUP BY vs.video_id
			)
			INSERT INTO leaderboards (brief_id, entity_type, entity_id, rank, total_revenue, total_orders)
			SELECT $1, 'video', entity_id, rank, total_revenue, total_orders
			FROM ranked
			ON CONFLICT (brief_id, entity_type, entity_id) DO UPDATE SET
				rank = EXCLUDED.rank,
				total_revenue = EXCLUDED.total_revenue,
				total_orders = EXCLUDED.total_orders,
				updated_at = NOW()
		`
	default:
		return errors.BadRequest("unsupported entity type for leaderboard")
	}

	_, err := r.db.Exec(ctx, rankingQuery, briefID)
	if err != nil {
		return err
	}
	return nil
}

// DailyMetrics operations

func (r *PerformanceRepository) InsertDailyMetric(ctx context.Context, date time.Time, videoID, campaignID string, orders int, revenue float64) error {
	query := `
		INSERT INTO daily_metrics (date, video_id, campaign_id, orders, revenue)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (date, video_id) DO UPDATE SET
			orders = daily_metrics.orders + $4,
			revenue = daily_metrics.revenue + $5
	`

	_, err := r.db.Exec(ctx, query, date, videoID, campaignID, orders, revenue)
	if err != nil {
		return err
	}
	return nil
}

func (r *PerformanceRepository) GetDailyMetrics(ctx context.Context, entityType, entityID string, startDate, endDate time.Time, granularity string) ([]map[string]interface{}, error) {
	var query string
	var rows pgx.Rows
	var err error

	dateGroupBy := "date"
	switch granularity {
	case "weekly":
		dateGroupBy = "DATE_TRUNC('week', date)"
	case "monthly":
		dateGroupBy = "DATE_TRUNC('month', date)"
	}

	switch entityType {
	case "video":
		query = fmt.Sprintf(`
			SELECT %s as period, SUM(orders) as orders, SUM(revenue) as revenue
			FROM daily_metrics
			WHERE video_id = $1 AND date >= $2 AND date <= $3
			GROUP BY %s
			ORDER BY period
		`, dateGroupBy, dateGroupBy)
		rows, err = r.db.Query(ctx, query, entityID, startDate, endDate)
	case "campaign":
		query = fmt.Sprintf(`
			SELECT %s as period, SUM(orders) as orders, SUM(revenue) as revenue
			FROM daily_metrics
			WHERE campaign_id = $1 AND date >= $2 AND date <= $3
			GROUP BY %s
			ORDER BY period
		`, dateGroupBy, dateGroupBy)
		rows, err = r.db.Query(ctx, query, entityID, startDate, endDate)
	default:
		return nil, errors.BadRequest("unsupported entity type for analytics")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var period time.Time
		var orders int
		var revenue pgtype.Numeric
		if err := rows.Scan(&period, &orders, &revenue); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"period":  period,
			"orders": orders,
			"revenue": revenue,
		})
	}
	return results, nil
}

// Anomaly operations (placeholder)

func (r *PerformanceRepository) GetAnomalies(ctx context.Context) ([]model.Anomaly, error) {
	// Placeholder - return empty array for now
	return []model.Anomaly{}, nil
}