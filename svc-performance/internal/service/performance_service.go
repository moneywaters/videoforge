package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/videoforge/backend/pkg/logger"
	"svc-performance/internal/model"
	"svc-performance/internal/repository"
)

type PerformanceService struct {
	repo *repository.PerformanceRepository
}

func NewPerformanceService(repo *repository.PerformanceRepository) *PerformanceService {
	return &PerformanceService{repo: repo}
}

// Video Sales

func (s *PerformanceService) GetVideoSales(ctx context.Context, videoID string) (*model.VideoSales, error) {
	return s.repo.GetVideoSales(ctx, videoID)
}

// Editor Sales

func (s *PerformanceService) GetEditorSales(ctx context.Context, editorID string) (*model.EditorSales, error) {
	return s.repo.GetEditorSales(ctx, editorID)
}

// Specialist Sales

func (s *PerformanceService) GetSpecialistSales(ctx context.Context, specialistID string) (*model.SpecialistSales, error) {
	return s.repo.GetSpecialistSales(ctx, specialistID)
}

// Campaign Sales

func (s *PerformanceService) GetCampaignSales(ctx context.Context, campaignID string) (*model.CampaignSales, error) {
	return s.repo.GetCampaignSales(ctx, campaignID)
}

// Leaderboard

func (s *PerformanceService) GetLeaderboard(ctx context.Context, briefID string, entityType string) ([]model.LeaderboardEntry, error) {
	// Optionally recalculate leaderboard before returning
	if err := s.repo.CalculateAndStoreLeaderboard(ctx, briefID, entityType); err != nil {
		logger.Warn("Failed to recalculate leaderboard", "error", err, "brief_id", briefID)
		// Continue with existing data
	}
	return s.repo.GetLeaderboard(ctx, briefID, entityType)
}

func (s *PerformanceService) GetRankings(ctx context.Context, briefID string, entityType string, limit, offset int) ([]model.LeaderboardEntry, error) {
	// Optionally recalculate leaderboard before returning
	if err := s.repo.CalculateAndStoreLeaderboard(ctx, briefID, entityType); err != nil {
		logger.Warn("Failed to recalculate leaderboard", "error", err, "brief_id", briefID)
	}
	return s.repo.GetLeaderboardRankings(ctx, briefID, entityType, limit, offset)
}

// Analytics

func (s *PerformanceService) GetAnalytics(ctx context.Context, query model.AnalyticsQuery) ([]map[string]interface{}, error) {
	startDate, err := time.Parse("2006-01-02", query.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", query.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}
	return s.repo.GetDailyMetrics(ctx, query.EntityType, query.EntityID, startDate, endDate, query.Granularity)
}

// Anomalies

func (s *PerformanceService) GetAnomalies(ctx context.Context) ([]model.Anomaly, error) {
	// Placeholder - return empty array for now
	return []model.Anomaly{}, nil
}

// ProcessSaleAttributedEvent handles sale.attributed NATS events
func (s *PerformanceService) ProcessSaleAttributedEvent(ctx context.Context, data []byte) error {
	var event model.SaleAttributedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		logger.Error("Failed to unmarshal sale event", "error", err)
		return err
	}

	logger.Info("Processing sale attributed event", "sale_id", event.SaleID, "video_id", event.VideoID, "amount", event.Amount)

	// Update video_sales
	if err := s.repo.UpsertVideoSales(ctx, event.VideoID, event.CampaignID, event.Amount, event.Currency); err != nil {
		logger.Error("Failed to update video sales", "error", err)
		return err
	}

	// Update editor_sales
	if event.EditorID != "" {
		if err := s.repo.UpsertEditorSales(ctx, event.EditorID, event.Amount, event.Currency); err != nil {
			logger.Error("Failed to update editor sales", "error", err)
			return err
		}
	}

	// Update specialist_sales
	if event.SpecialistID != "" {
		if err := s.repo.UpsertSpecialistSales(ctx, event.SpecialistID, event.CampaignID, event.Amount, event.Currency); err != nil {
			logger.Error("Failed to update specialist sales", "error", err)
			return err
		}
	}

	// Update campaign_sales
	if err := s.repo.UpsertCampaignSales(ctx, event.CampaignID, event.Amount, event.Currency, time.Time{}, time.Time{}); err != nil {
		logger.Error("Failed to update campaign sales", "error", err)
		return err
	}

	// Insert daily metric
	date := event.Timestamp.UTC().Truncate(24 * time.Hour)
	if err := s.repo.InsertDailyMetric(ctx, date, event.VideoID, event.CampaignID, 1, event.Amount); err != nil {
		logger.Error("Failed to insert daily metric", "error", err)
		return err
	}

	// Optionally recalculate leaderboard for the brief
	if event.BriefID != "" {
		if err := s.repo.CalculateAndStoreLeaderboard(ctx, event.BriefID, "video"); err != nil {
			logger.Warn("Failed to recalculate video leaderboard", "error", err)
		}
		if err := s.repo.CalculateAndStoreLeaderboard(ctx, event.BriefID, "editor"); err != nil {
			logger.Warn("Failed to recalculate editor leaderboard", "error", err)
		}
	}

	logger.Info("Successfully processed sale attributed event", "sale_id", event.SaleID)
	return nil
}