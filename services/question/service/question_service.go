package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"preppi.com/services/question/repository"
)

var ErrValidation = errors.New("validation error")

type QuestionService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *QuestionService {
	return &QuestionService{repo: repo}
}

func (s *QuestionService) Create(ctx context.Context, studentID uint, subject, topic, description string, imageURLs []string, urgency string) (uint, string, error) {
	if subject == "" || description == "" {
		return 0, "", ErrValidation
	}
	q := &repository.Question{
		StudentID:   studentID,
		Subject:     subject,
		Topic:       topic,
		Description: description,
		ImageURLs:   strings.Join(imageURLs, ","),
		Urgency:     normalizeUrgency(urgency),
		Status:      "open",
	}
	if err := s.repo.Create(ctx, q); err != nil {
		return 0, "", err
	}
	return q.ID, q.Status, nil
}

func (s *QuestionService) GetByID(ctx context.Context, id uint) (*repository.Question, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *QuestionService) ListByStudent(ctx context.Context, studentID uint, limit, offset int) ([]repository.Question, error) {
	return s.repo.List(ctx, repository.Filter{StudentID: studentID, Limit: limit, Offset: offset})
}

func (s *QuestionService) Search(ctx context.Context, query, subject string, limit, offset int) ([]repository.Question, error) {
	return s.repo.Search(ctx, query, subject, limit, offset)
}

func (s *QuestionService) UpdateStatus(ctx context.Context, id uint, status string) error {
	if !validStatus(status) {
		return fmt.Errorf("invalid status %q", status)
	}
	return s.repo.UpdateFields(ctx, id, map[string]any{"status": status})
}

func (s *QuestionService) Assign(ctx context.Context, id uint, mentorID uint) error {
	return s.repo.UpdateFields(ctx, id, map[string]any{
		"assignee_id": mentorID,
		"status":      "assigned",
	})
}

func normalizeUrgency(u string) string {
	switch u {
	case "low", "urgent":
		return u
	default:
		return "normal"
	}
}

func validStatus(s string) bool {
	switch s {
	case "open", "assigned", "in_progress", "answered", "escalated":
		return true
	default:
		return false
	}
}

func SplitImageURLs(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
