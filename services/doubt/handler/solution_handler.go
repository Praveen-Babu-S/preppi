package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "preppi.com/proto/doubt/v1"
	"preppi.com/services/doubt/repository"
	"preppi.com/services/doubt/service"
)

type SolutionHandler struct {
	pb.UnimplementedSolutionServiceServer
	svc *service.DoubtService
}

func NewSolutionHandler(svc *service.DoubtService) *SolutionHandler {
	return &SolutionHandler{svc: svc}
}

func (h *SolutionHandler) CreateSolution(ctx context.Context, req *pb.CreateSolutionRequest) (*pb.CreateSolutionResponse, error) {
	questionID, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}
	mentorID, err := parseID(req.GetMentorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mentor_id")
	}
	sol, err := h.svc.CreateSolution(ctx, questionID, mentorID, req.GetDescription(), req.GetImageUrls())
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, "description is required")
		}
		return nil, status.Error(codes.Internal, "failed to create solution")
	}
	return &pb.CreateSolutionResponse{Solution: toSolutionPB(sol)}, nil
}

func (h *SolutionHandler) GetSolutionById(ctx context.Context, req *pb.GetSolutionByIdRequest) (*pb.GetSolutionByIdResponse, error) {
	id, err := parseID(req.GetSolutionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid solution_id")
	}
	sol, err := h.svc.GetSolution(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "solution not found")
	}
	return &pb.GetSolutionByIdResponse{Solution: toSolutionPB(sol)}, nil
}

func (h *SolutionHandler) ListSolutionsByQuestion(ctx context.Context, req *pb.ListSolutionsByQuestionRequest) (*pb.ListSolutionsByQuestionResponse, error) {
	questionID, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	solutions, err := h.svc.ListSolutionsByQuestion(ctx, questionID, limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list solutions")
	}
	resp := &pb.ListSolutionsByQuestionResponse{}
	for i := range solutions {
		resp.Solutions = append(resp.Solutions, toSolutionPB(&solutions[i]))
	}
	return resp, nil
}

func (h *SolutionHandler) VoteSolution(ctx context.Context, req *pb.VoteSolutionRequest) (*pb.VoteSolutionResponse, error) {
	id, err := parseID(req.GetSolutionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid solution_id")
	}
	isUpvote := req.GetVote() == pb.VoteType_VOTE_TYPE_UP
	upvotes, downvotes, err := h.svc.Vote(ctx, id, isUpvote)
	if err != nil {
		if errors.Is(err, service.ErrSolutionNotFound) {
			return nil, status.Error(codes.NotFound, "solution not found")
		}
		return nil, status.Error(codes.Internal, "failed to vote")
	}
	return &pb.VoteSolutionResponse{Upvotes: int32(upvotes), Downvotes: int32(downvotes)}, nil
}

func (h *SolutionHandler) AcceptSolution(ctx context.Context, req *pb.AcceptSolutionRequest) (*pb.AcceptSolutionResponse, error) {
	id, err := parseID(req.GetSolutionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid solution_id")
	}
	accepted, err := h.svc.Accept(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrSolutionNotFound) {
			return nil, status.Error(codes.NotFound, "solution not found")
		}
		return nil, status.Error(codes.Internal, "failed to accept solution")
	}
	return &pb.AcceptSolutionResponse{Accepted: accepted}, nil
}

func (h *SolutionHandler) CreateFollowUp(ctx context.Context, req *pb.CreateFollowUpRequest) (*pb.CreateFollowUpResponse, error) {
	solutionID, err := parseID(req.GetSolutionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid solution_id")
	}
	userID, err := parseID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	fu, err := h.svc.CreateFollowUp(ctx, solutionID, userID, req.GetMessage(), req.GetImageUrls())
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, "message is required")
		}
		return nil, status.Error(codes.Internal, "failed to create follow-up")
	}
	return &pb.CreateFollowUpResponse{FollowUp: followUpToPB(fu)}, nil
}

func (h *SolutionHandler) ListFollowUps(ctx context.Context, req *pb.ListFollowUpsRequest) (*pb.ListFollowUpsResponse, error) {
	solutionID, err := parseID(req.GetSolutionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid solution_id")
	}
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	followUps, err := h.svc.ListFollowUps(ctx, solutionID, limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list follow-ups")
	}
	resp := &pb.ListFollowUpsResponse{}
	for i := range followUps {
		resp.FollowUps = append(resp.FollowUps, followUpToPB(&followUps[i]))
	}
	return resp, nil
}

func toSolutionPB(s *repository.Solution) *pb.Solution {
	if s == nil {
		return &pb.Solution{}
	}
	return &pb.Solution{
		Id:          uintToStr(s.ID),
		QuestionId:  uintToStr(s.QuestionID),
		MentorId:    uintToStr(s.MentorID),
		Description: s.Description,
		ImageUrls:   service.SplitImageURLs(s.ImageURLs),
		Upvotes:     int32(s.Upvotes),
		Downvotes:   int32(s.Downvotes),
		IsAccepted:  s.IsAccepted,
		CreatedAt:   timestamppb.New(unixTime(s.CreatedAt)),
		UpdatedAt:   timestamppb.New(unixTime(s.UpdatedAt)),
	}
}

func followUpToPB(f *repository.FollowUp) *pb.FollowUp {
	if f == nil {
		return &pb.FollowUp{}
	}
	return &pb.FollowUp{
		Id:         uintToStr(f.ID),
		SolutionId: uintToStr(f.SolutionID),
		UserId:     uintToStr(f.UserID),
		Message:    f.Message,
		ImageUrls:  service.SplitImageURLs(f.ImageURLs),
		CreatedAt:  timestamppb.New(unixTime(f.CreatedAt)),
	}
}
