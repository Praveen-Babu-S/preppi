package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "preppi.com/proto/auth/v1"
	"preppi.com/services/auth/service"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func New(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email and password are required")
	}

	role := req.GetRole().String()
	switch role {
	case "ROLE_STUDENT":
		role = "student"
	case "ROLE_MENTOR":
		role = "mentor"
	case "ROLE_ADMIN":
		role = "admin"
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	id, email, err := h.svc.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword(), role, req.GetSubject())
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &pb.RegisterResponse{UserId: uint64ToStr(id), Email: email}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	access, refresh, id, role, err := h.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, service.ErrInvalidCreds) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "login failed")
	}

	now := timestamppb.Now()
	return &pb.LoginResponse{
		AccessToken:      access,
		RefreshToken:     refresh,
		UserId:           uint64ToStr(id),
		Role:             roleToPB(role),
		AccessExpiresAt:  now,
		RefreshExpiresAt: now,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	access, err := h.svc.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return nil, err
		}
		return nil, status.Error(codes.Internal, "refresh failed")
	}
	return &pb.RefreshTokenResponse{AccessToken: access, AccessExpiresAt: timestamppb.Now()}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{Success: true}, nil
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	return &pb.VerifyEmailResponse{Verified: false}, nil
}

func uint64ToStr(id uint) string {
	return fmtUint(uint64(id))
}
