package handler

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "preppi.com/proto/analytics/v1"
	"preppi.com/services/analytics/service"
)

type AnalyticsHandler struct {
	pb.UnimplementedAnalyticsServiceServer
	svc *service.AnalyticsService
}

func New(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 10
	}
	entries, err := h.svc.GetLeaderboard(ctx, req.GetPeriod(), limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get leaderboard")
	}
	resp := &pb.GetLeaderboardResponse{}
	for i, e := range entries {
		resp.Entries = append(resp.Entries, &pb.LeaderboardEntry{
			MentorId:          uintToStr(e.MentorID),
			Rank:              int32(i + 1),
			QuestionsAnswered: int32(e.QuestionsAnswered),
			Upvotes:           int32(e.TotalUpvotes),
		})
	}
	return resp, nil
}

func (h *AnalyticsHandler) GetStudentStats(ctx context.Context, req *pb.GetStudentStatsRequest) (*pb.GetStudentStatsResponse, error) {
	studentID, err := parseID(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}
	stat, err := h.svc.GetStudentStats(ctx, studentID, req.GetPeriod())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get student stats")
	}
	resp := &pb.GetStudentStatsResponse{
		QuestionsAsked:    int32(stat.QuestionsAsked),
		QuestionsAnswered: int32(stat.QuestionsAnswered),
	}
	if stat.QuestionsAsked > 0 {
		resp.ResolutionRate = float64(stat.QuestionsAnswered) / float64(stat.QuestionsAsked)
	}
	return resp, nil
}

func (h *AnalyticsHandler) GetMentorStats(ctx context.Context, req *pb.GetMentorStatsRequest) (*pb.GetMentorStatsResponse, error) {
	mentorID, err := parseID(req.GetMentorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mentor_id")
	}
	stat, err := h.svc.GetMentorStats(ctx, mentorID, req.GetPeriod())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get mentor stats")
	}
	return &pb.GetMentorStatsResponse{
		QuestionsAnswered: int32(stat.QuestionsAnswered),
		AcceptedSolutions: int32(stat.AcceptedSolutions),
		TotalUpvotes:      int32(stat.TotalUpvotes),
	}, nil
}

func (h *AnalyticsHandler) GetPlatformMetrics(ctx context.Context, req *pb.GetPlatformMetricsRequest) (*pb.GetPlatformMetricsResponse, error) {
	m, err := h.svc.GetPlatformMetrics(ctx, req.GetPeriod())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get platform metrics")
	}
	return &pb.GetPlatformMetricsResponse{
		ActiveStudents:       int32(m.ActiveStudents),
		ActiveMentors:        int32(m.ActiveMentors),
		TotalQuestions:       int32(m.TotalQuestions),
		PendingQuestions:     int32(m.PendingQuestions),
		AvgResponseTimeHours: m.AvgResponseHours,
		ResolutionRate:       m.ResolutionRate,
	}, nil
}

func parseID(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func uintToStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
