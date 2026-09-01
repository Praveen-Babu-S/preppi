package service

import (
	"context"
	"errors"
	"strings"

	"preppi.com/services/solution/repository"
)

var (
	ErrValidation         = errors.New("validation error")
	ErrSolutionNotFound   = errors.New("solution not found")
	ErrNotQuestionStudent = errors.New("only the question student can accept a solution")
)

type SolutionService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *SolutionService {
	return &SolutionService{repo: repo}
}

func (s *SolutionService) Create(ctx context.Context, questionID, mentorID uint, description string, imageURLs []string) (*repository.Solution, error) {
	if description == "" {
		return nil, ErrValidation
	}
	sol := &repository.Solution{
		QuestionID:  questionID,
		MentorID:    mentorID,
		Description: description,
		ImageURLs:   strings.Join(imageURLs, ","),
	}
	if err := s.repo.CreateSolution(ctx, sol); err != nil {
		return nil, err
	}
	return sol, nil
}

func (s *SolutionService) GetByID(ctx context.Context, id uint) (*repository.Solution, error) {
	return s.repo.FindSolutionByID(ctx, id)
}

func (s *SolutionService) ListByQuestion(ctx context.Context, questionID uint, limit, offset int) ([]repository.Solution, error) {
	return s.repo.ListSolutionsByQuestion(ctx, questionID, limit, offset)
}

func (s *SolutionService) Vote(ctx context.Context, id uint, isUpvote bool) (int, int, error) {
	sol, err := s.repo.FindSolutionByID(ctx, id)
	if err != nil {
		return 0, 0, ErrSolutionNotFound
	}
	if isUpvote {
		sol.Upvotes++
	} else {
		sol.Downvotes++
	}
	if err := s.repo.UpdateSolutionVotes(ctx, id, sol.Upvotes, sol.Downvotes); err != nil {
		return 0, 0, err
	}
	return sol.Upvotes, sol.Downvotes, nil
}

func (s *SolutionService) Accept(ctx context.Context, id uint) (bool, error) {
	sol, err := s.repo.FindSolutionByID(ctx, id)
	if err != nil {
		return false, ErrSolutionNotFound
	}
	if err := s.repo.UnacceptOtherSolutions(ctx, sol.QuestionID); err != nil {
		return false, err
	}
	if err := s.repo.AcceptSolution(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

func (s *SolutionService) CreateFollowUp(ctx context.Context, solutionID, userID uint, message string, imageURLs []string) (*repository.FollowUp, error) {
	if message == "" {
		return nil, ErrValidation
	}
	fu := &repository.FollowUp{
		SolutionID: solutionID,
		UserID:     userID,
		Message:    message,
		ImageURLs:  strings.Join(imageURLs, ","),
	}
	if err := s.repo.CreateFollowUp(ctx, fu); err != nil {
		return nil, err
	}
	return fu, nil
}

func (s *SolutionService) ListFollowUps(ctx context.Context, solutionID uint, limit, offset int) ([]repository.FollowUp, error) {
	return s.repo.ListFollowUps(ctx, solutionID, limit, offset)
}

func SplitImageURLs(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
