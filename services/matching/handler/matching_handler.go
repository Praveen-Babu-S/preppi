package handler

import (
	"context"
	"errors"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "preppi.com/proto/matching/v1"
	"preppi.com/services/matching/service"
)

type MatchingHandler struct {
	pb.UnimplementedMatchingServiceServer
	svc *service.MatchingService
}

func New(svc *service.MatchingService) *MatchingHandler {
	return &MatchingHandler{svc: svc}
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

func (h *MatchingHandler) AssignMentor(ctx context.Context, req *pb.AssignMentorRequest) (*pb.AssignMentorResponse, error) {
	questionID, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}

	// In production, the caller provides candidate mentor IDs (from user service).
	// For this foundation, we accept a list from the request.
	// Since the proto doesn't have candidate IDs, we use a placeholder approach:
	// the orchestrator passes candidate IDs via metadata or a future proto change.
	// For now, return a TODO-style response indicating the routing flow.
	_ = questionID

	return &pb.AssignMentorResponse{
		AssignmentId: "",
		MentorId:     "",
		Status:       pb.AssignmentStatus_ASSIGNMENT_STATUS_PENDING,
	}, nil
}

func (h *MatchingHandler) SkipQuestion(ctx context.Context, req *pb.SkipQuestionRequest) (*pb.SkipQuestionResponse, error) {
	assignmentID, err := parseID(req.GetAssignmentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid assignment_id")
	}

	needReassign, err := h.svc.SkipQuestion(ctx, assignmentID, req.GetReason())
	if err != nil {
		if errors.Is(err, service.ErrAssignmentNotFound) {
			return nil, status.Error(codes.NotFound, "assignment not found")
		}
		return nil, status.Error(codes.Internal, "skip failed")
	}

	return &pb.SkipQuestionResponse{
		Skipped:          true,
		NeedReassignment: needReassign,
	}, nil
}

func (h *MatchingHandler) GetNextQuestion(ctx context.Context, req *pb.GetNextQuestionRequest) (*pb.GetNextQuestionResponse, error) {
	mentorID, err := parseID(req.GetMentorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mentor_id")
	}

	assignmentID, questionID, err := h.svc.GetNextQuestion(ctx, mentorID)
	if err != nil {
		if errors.Is(err, service.ErrNoCandidates) {
			return nil, status.Error(codes.NotFound, "no pending questions")
		}
		return nil, status.Error(codes.Internal, "failed to get next question")
	}
	_ = assignmentID

	return &pb.GetNextQuestionResponse{
		QuestionId: uintToStr(questionID),
	}, nil
}

func (h *MatchingHandler) GetPendingForMentor(ctx context.Context, req *pb.GetPendingForMentorRequest) (*pb.GetPendingForMentorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *MatchingHandler) EscalateQuestion(ctx context.Context, req *pb.EscalateQuestionRequest) (*pb.EscalateQuestionResponse, error) {
	questionID, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}

	level, err := h.svc.EscalateQuestion(ctx, questionID, req.GetReason())
	if err != nil {
		if errors.Is(err, service.ErrMaxEscalation) {
			return nil, status.Error(codes.FailedPrecondition, "max escalation level reached")
		}
		return nil, status.Error(codes.Internal, "escalation failed")
	}

	return &pb.EscalateQuestionResponse{EscalationLevel: int32(level)}, nil
}

// placeholder for future: maps assignment to pb
func _now() *timestamppb.Timestamp { return timestamppb.Now() }

var _ = time.Now
