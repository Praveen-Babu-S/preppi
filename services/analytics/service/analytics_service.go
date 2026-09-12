package service

import (
	"context"
	"fmt"
	"time"

	"preppi.com/services/analytics/repository"
)

type AnalyticsService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetLeaderboard(ctx context.Context, period string, limit int) ([]repository.MentorStat, error) {
	since := periodToSince(period)
	stats, err := s.repo.GetLeaderboard(ctx, since, limit)
	if err != nil {
		return nil, fmt.Errorf("analytics_service_get_leaderboard: %w", err)
	}
	return stats, nil
}

func (s *AnalyticsService) GetStudentStats(ctx context.Context, studentID uint, period string) (*repository.StudentStat, error) {
	since := periodToSince(period)
	stat, err := s.repo.GetStudentStats(ctx, studentID, since)
	if err != nil {
		return nil, fmt.Errorf("analytics_service_get_student_stats: %w", err)
	}
	return stat, nil
}

func (s *AnalyticsService) GetMentorStats(ctx context.Context, mentorID uint, period string) (*repository.MentorStat, error) {
	since := periodToSince(period)
	stat, err := s.repo.GetMentorStats(ctx, mentorID, since)
	if err != nil {
		return nil, fmt.Errorf("analytics_service_get_mentor_stats: %w", err)
	}
	return stat, nil
}

func (s *AnalyticsService) GetPlatformMetrics(ctx context.Context, period string) (*repository.PlatformMetrics, error) {
	since := periodToSince(period)
	m, err := s.repo.GetPlatformMetrics(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("analytics_service_get_platform_metrics: %w", err)
	}
	return m, nil
}

func periodToSince(period string) time.Time {
	now := time.Now()
	switch period {
	case "day":
		return now.AddDate(0, 0, -1)
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "year":
		return now.AddDate(-1, 0, 0)
	default:
		return now.AddDate(0, -1, 0)
	}
}
