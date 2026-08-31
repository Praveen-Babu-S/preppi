package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"preppi.com/services/matching/repository"
)

var (
	ErrNoCandidates      = errors.New("no candidates available")
	ErrAlreadyAssigned   = errors.New("question already assigned")
	ErrMaxEscalation     = errors.New("max escalation level reached")
	ErrAssignmentNotFound = errors.New("assignment not found")
)

const MaxEscalationLevel = 3

type MatchingService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *MatchingService {
	return &MatchingService{repo: repo}
}

// AssignMentor picks the least loaded mentor from candidates and creates an assignment.
func (s *MatchingService) AssignMentor(ctx context.Context, questionID uint, candidateMentorIDs []uint) (uint, uint, error) {
	if len(candidateMentorIDs) == 0 {
		return 0, 0, ErrNoCandidates
	}

	// Check if already assigned
	if existing, _ := s.repo.GetAssignmentByQuestion(ctx, questionID); existing != nil {
		return 0, 0, ErrAlreadyAssigned
	}

	// Find least loaded mentor from candidates
	mentorID, err := s.findLeastLoaded(ctx, candidateMentorIDs)
	if err != nil {
		return 0, 0, err
	}

	assignment := &repository.Assignment{
		QuestionID: questionID,
		MentorID:   mentorID,
		Status:     "pending",
	}
	if err := s.repo.CreateAssignment(ctx, assignment); err != nil {
		return 0, 0, err
	}
	return assignment.ID, mentorID, nil
}

// SkipQuestion marks the assignment as skipped and returns whether reassignment is needed.
func (s *MatchingService) SkipQuestion(ctx context.Context, assignmentID uint, reason string) (bool, error) {
	a, err := s.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return false, ErrAssignmentNotFound
	}

	if err := s.repo.UpdateAssignmentStatus(ctx, assignmentID, "skipped", time.Now()); err != nil {
		return false, err
	}

	// Reassign: try to find another candidate for this question
	if _, err := s.repo.GetAssignmentByQuestion(ctx, a.QuestionID); err == nil {
		return true, nil
	}
	return false, nil
}

// GetNextQuestion returns the oldest pending assignment for a mentor.
func (s *MatchingService) GetNextQuestion(ctx context.Context, mentorID uint) (uint, uint, error) {
	assignments, err := s.repo.GetPendingForMentor(ctx, mentorID, 1, 0)
	if err != nil {
		return 0, 0, err
	}
	if len(assignments) == 0 {
		return 0, 0, ErrNoCandidates
	}
	return assignments[0].ID, assignments[0].QuestionID, nil
}

// CompleteAssignment marks an assignment as completed when a mentor provides a solution.
func (s *MatchingService) CompleteAssignment(ctx context.Context, assignmentID uint) error {
	return s.repo.UpdateAssignmentStatus(ctx, assignmentID, "completed", time.Now())
}

// EscalateQuestion bumps the escalation level for a question.
func (s *MatchingService) EscalateQuestion(ctx context.Context, questionID uint, reason string) (int, error) {
	latest, _ := s.repo.GetLatestEscalation(ctx, questionID)
	newLevel := latest.EscalationLevel + 1
	if newLevel > MaxEscalationLevel {
		return 0, ErrMaxEscalation
	}

	esc := &repository.Escalation{
		QuestionID:      questionID,
		EscalationLevel: newLevel,
		Reason:          reason,
	}
	if err := s.repo.CreateEscalation(ctx, esc); err != nil {
		return 0, err
	}
	return newLevel, nil
}

// findLeastLoaded returns the mentor with the fewest pending assignments.
func (s *MatchingService) findLeastLoaded(ctx context.Context, candidates []uint) (uint, error) {
	bestID := candidates[0]
	bestCount := int(^uint(0) >> 1) // max int

	for _, id := range candidates {
		count, err := s.repo.GetPendingCount(ctx, id)
		if err != nil {
			continue
		}
		if count < bestCount {
			bestCount = count
			bestID = id
		}
	}
	return bestID, nil
}
