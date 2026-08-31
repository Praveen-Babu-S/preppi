package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	pb "preppi.com/proto/user/v1"
	"preppi.com/services/user/repository"
	"preppi.com/services/user/service"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	svc *service.UserService
}

func New(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	userID, err := service.ParseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := h.svc.CreateProfile(
		ctx, userID, req.GetName(), req.GetAvatarUrl(), req.GetPhone(),
		req.GetSchool(), req.GetCollege(), req.GetBio(), roleToString(req.GetRole()),
		req.GetExpertiseSubjects(), req.GetSubTopics(),
	); err != nil {
		return nil, status.Error(codes.Internal, "failed to create profile")
	}

	prof, _ := h.svc.GetProfile(ctx, userID)
	return &pb.CreateProfileResponse{Profile: toProfile(prof)}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	userID, err := service.ParseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	p, err := h.svc.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to get profile")
	}
	return &pb.GetProfileResponse{Profile: toProfile(p)}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	userID, err := service.ParseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := h.svc.UpdateProfile(ctx, userID, req.GetName(), req.GetAvatarUrl(), req.GetBio(), req.GetSchool(), req.GetCollege()); err != nil {
		return nil, status.Error(codes.Internal, "failed to update profile")
	}
	p, _ := h.svc.GetProfile(ctx, userID)
	return &pb.UpdateProfileResponse{Profile: toProfile(p)}, nil
}

func (h *UserHandler) UpdateMentorProfile(ctx context.Context, req *pb.UpdateMentorProfileRequest) (*pb.UpdateMentorProfileResponse, error) {
	userID, err := service.ParseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := h.svc.UpdateMentorProfile(ctx, userID, req.GetExpertiseSubjects(), req.GetSubTopics(), req.GetBio()); err != nil {
		return nil, status.Error(codes.Internal, "failed to update mentor profile")
	}

	m, err := h.svc.GetMentorProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get mentor profile")
	}
	return &pb.UpdateMentorProfileResponse{Mentor: toMentor(m)}, nil
}

func (h *UserHandler) GetMentorsBySubject(ctx context.Context, req *pb.GetMentorsBySubjectRequest) (*pb.GetMentorsBySubjectResponse, error) {
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	mentors, err := h.svc.GetMentorsBySubject(ctx, req.GetSubject(), limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get mentors")
	}
	resp := &pb.GetMentorsBySubjectResponse{}
	for i := range mentors {
		resp.Mentors = append(resp.Mentors, toMentor(&mentors[i]))
	}
	return resp, nil
}

func (h *UserHandler) SetOnlineStatus(ctx context.Context, req *pb.SetOnlineStatusRequest) (*pb.SetOnlineStatusResponse, error) {
	userID, err := service.ParseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	online := req.GetStatus() == pb.OnlineStatus_ONLINE_STATUS_ONLINE
	if err := h.svc.SetOnline(ctx, userID, online); err != nil {
		return nil, status.Error(codes.Internal, "failed to update online status")
	}
	return &pb.SetOnlineStatusResponse{Success: true}, nil
}

func (h *UserHandler) ApproveMentor(ctx context.Context, req *pb.ApproveMentorRequest) (*pb.ApproveMentorResponse, error) {
	userID, err := service.ParseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	if err := h.svc.ApproveMentor(ctx, userID, req.GetApproved()); err != nil {
		return nil, status.Error(codes.Internal, "failed to approve mentor")
	}
	statusStr := pb.VerificationStatus_VERIFICATION_STATUS_APPROVED
	if !req.GetApproved() {
		statusStr = pb.VerificationStatus_VERIFICATION_STATUS_REJECTED
	}
	return &pb.ApproveMentorResponse{Status: statusStr}, nil
}

func roleToString(r pb.Role) string {
	switch r {
	case pb.Role_ROLE_STUDENT:
		return "student"
	case pb.Role_ROLE_MENTOR:
		return "mentor"
	case pb.Role_ROLE_ADMIN:
		return "admin"
	default:
		return "student"
	}
}

func toProfile(p *repository.Profile) *pb.Profile {
	if p == nil {
		return &pb.Profile{}
	}
	online := pb.OnlineStatus_ONLINE_STATUS_OFFLINE
	if p.Online {
		online = pb.OnlineStatus_ONLINE_STATUS_ONLINE
	}
	return &pb.Profile{
		UserId:       uintToStr(p.UserID),
		Name:         p.Name,
		AvatarUrl:    p.AvatarURL,
		Phone:        p.Phone,
		School:       p.School,
		College:      p.College,
		Bio:          p.Bio,
		Role:         roleFromString(p.Role),
		OnlineStatus: online,
		CreatedAt:    timestamppb.New(unixTime(p.CreatedAt)),
	}
}

func toMentor(m *repository.MentorProfile) *pb.MentorProfile {
	if m == nil {
		return &pb.MentorProfile{}
	}
	verification := pb.VerificationStatus_VERIFICATION_STATUS_PENDING
	switch m.VerificationStatus {
	case "approved":
		verification = pb.VerificationStatus_VERIFICATION_STATUS_APPROVED
	case "rejected":
		verification = pb.VerificationStatus_VERIFICATION_STATUS_REJECTED
	}
	return &pb.MentorProfile{
		UserId:             uintToStr(m.UserID),
		ExpertiseSubjects:  service.SplitCSV(m.ExpertiseSubjects),
		SubTopics:          service.SplitCSV(m.SubTopics),
		VerificationStatus: verification,
		Rating:             m.Rating,
		QuestionsAnswered:  int32(m.QuestionsAnswered),
	}
}
