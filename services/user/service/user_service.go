package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"preppi.com/services/user/repository"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
	ErrNoMentorProfile = errors.New("mentor profile not found")
	ErrInvalidUserID   = errors.New("invalid user id")
)

type UserService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateProfile(ctx context.Context, userID uint, name, avatarURL, phone, school, college, bio, role string, expertise, subTopics []string) error {
	p := &repository.Profile{
		UserID:    userID,
		Name:      name,
		AvatarURL: avatarURL,
		Phone:     phone,
		School:    school,
		College:   college,
		Bio:       bio,
		Role:      role,
	}
	if err := s.repo.CreateProfile(ctx, p); err != nil {
		return err
	}

	if role == "mentor" {
		m := &repository.MentorProfile{
			UserID:             userID,
			ExpertiseSubjects:  strings.Join(expertise, ","),
			SubTopics:          strings.Join(subTopics, ","),
			VerificationStatus: "pending",
		}
		if err := s.repo.UpsertMentorProfile(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*repository.Profile, error) {
	return s.repo.GetProfile(ctx, userID)
}

func (s *UserService) GetMentorProfile(ctx context.Context, userID uint) (*repository.MentorProfile, error) {
	return s.repo.GetMentorProfile(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, name, avatarURL, bio, school, college string) error {
	p, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return ErrProfileNotFound
	}
	if name != "" {
		p.Name = name
	}
	if avatarURL != "" {
		p.AvatarURL = avatarURL
	}
	if bio != "" {
		p.Bio = bio
	}
	if school != "" {
		p.School = school
	}
	if college != "" {
		p.College = college
	}
	return s.repo.UpdateProfile(ctx, p)
}

func (s *UserService) UpdateMentorProfile(ctx context.Context, userID uint, expertise, subTopics []string, bio string) error {
	m, err := s.repo.GetMentorProfile(ctx, userID)
	if err != nil {
		return ErrNoMentorProfile
	}
	m.ExpertiseSubjects = strings.Join(expertise, ",")
	m.SubTopics = strings.Join(subTopics, ",")
	if bio != "" {
		if p, perr := s.repo.GetProfile(ctx, userID); perr == nil {
			p.Bio = bio
			_ = s.repo.UpdateProfile(ctx, p)
		}
	}
	return s.repo.UpsertMentorProfile(ctx, m)
}

func (s *UserService) GetMentorsBySubject(ctx context.Context, subject string, limit, offset int) ([]repository.MentorProfile, error) {
	return s.repo.GetMentorsBySubject(ctx, subject, limit, offset)
}

func (s *UserService) SetOnline(ctx context.Context, userID uint, online bool) error {
	return s.repo.SetOnline(ctx, userID, online)
}

func (s *UserService) ApproveMentor(ctx context.Context, userID uint, approved bool) error {
	status := "approved"
	if !approved {
		status = "rejected"
	}
	return s.repo.ApproveMentor(ctx, userID, status)
}

func (s *UserService) GetPendingMentors(ctx context.Context, limit, offset int) ([]repository.MentorProfile, error) {
	return s.repo.GetPendingMentors(ctx, limit, offset)
}

func SplitCSV(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func ParseUserID(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, ErrInvalidUserID
	}
	return uint(v), nil
}
