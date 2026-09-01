package service

import (
	"context"
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
	return s.repo.GetLeaderboard(ctx, since, limit)
}

func (s *AnalyticsService) GetStudentStats(ctx context.Context, studentID uint, period string) (*repository.StudentStat, error) {
	since := periodToSince(period)
	return s.repo.GetStudentStats(ctx, studentID, since)
}

func (s *AnalyticsService) GetMentorStats(ctx context.Context, mentorID uint, period string) (*repository.MentorStat, error) {
	since := periodToSince(period)
	return s.repo.GetMentorStats(ctx, mentorID, since)
}

func (s *AnalyticsService) GetPlatformMetrics(ctx context.Context, period string) (*repository.PlatformMetrics, error) {
	since := periodToSince(period)
	return s.repo.GetPlatformMetrics(ctx, since)
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
