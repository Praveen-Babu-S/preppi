package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"preppi.com/services/doubt/repository"
)

var (
	ErrValidation         = errors.New("validation error")
	ErrSolutionNotFound   = errors.New("solution not found")
	ErrRoomNotFound       = errors.New("chat room not found")
	ErrNotQuestionStudent = errors.New("only the question student can accept a solution")
)

type DoubtService struct {
	repo        repository.Repository
	mu          sync.RWMutex
	subscribers map[uint][]chan *repository.Message
}

func New(repo repository.Repository) *DoubtService {
	return &DoubtService{
		repo:        repo,
		subscribers: make(map[uint][]chan *repository.Message),
	}
}

// ── Question ────────────────────────────────

func (s *DoubtService) CreateQuestion(ctx context.Context, studentID uint, subject, topic, description string, imageURLs []string, urgency string) (uint, string, error) {
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
	if err := s.repo.CreateQuestion(ctx, q); err != nil {
		return 0, "", err
	}
	return q.ID, q.Status, nil
}

func (s *DoubtService) GetQuestion(ctx context.Context, id uint) (*repository.Question, error) {
	return s.repo.FindQuestionByID(ctx, id)
}

func (s *DoubtService) ListQuestionsByStudent(ctx context.Context, studentID uint, limit, offset int) ([]repository.Question, error) {
	return s.repo.ListQuestions(ctx, repository.Filter{StudentID: studentID, Limit: limit, Offset: offset})
}

func (s *DoubtService) SearchQuestions(ctx context.Context, query, subject string, limit, offset int) ([]repository.Question, error) {
	return s.repo.SearchQuestions(ctx, query, subject, limit, offset)
}

func (s *DoubtService) UpdateQuestionStatus(ctx context.Context, id uint, status string) error {
	if !validStatus(status) {
		return fmt.Errorf("invalid status %q", status)
	}
	return s.repo.UpdateQuestionFields(ctx, id, map[string]any{"status": status})
}

func (s *DoubtService) UpdateQuestion(ctx context.Context, id uint, fields map[string]any) error {
	if len(fields) == 0 {
		return ErrValidation
	}
	if _, ok := fields["status"]; ok {
		if v, isStr := fields["status"].(string); isStr && !validStatus(v) {
			return fmt.Errorf("invalid status %q", v)
		}
	}
	if _, ok := fields["urgency"]; ok {
		if v, isStr := fields["urgency"].(string); isStr {
			if v != "low" && v != "normal" && v != "urgent" {
				return fmt.Errorf("invalid urgency %q", v)
			}
		}
	}
	return s.repo.UpdateQuestionFields(ctx, id, fields)
}

func (s *DoubtService) Assign(ctx context.Context, id uint, mentorID uint) error {
	return s.repo.UpdateQuestionFields(ctx, id, map[string]any{
		"assignee_id": mentorID,
		"status":      "assigned",
	})
}

// ── Solution ────────────────────────────────

func (s *DoubtService) CreateSolution(ctx context.Context, questionID, mentorID uint, description string, imageURLs []string) (*repository.Solution, error) {
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

func (s *DoubtService) GetSolution(ctx context.Context, id uint) (*repository.Solution, error) {
	return s.repo.FindSolutionByID(ctx, id)
}

func (s *DoubtService) ListSolutionsByQuestion(ctx context.Context, questionID uint, limit, offset int) ([]repository.Solution, error) {
	return s.repo.ListSolutionsByQuestion(ctx, questionID, limit, offset)
}

func (s *DoubtService) Vote(ctx context.Context, id uint, isUpvote bool) (int, int, error) {
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

func (s *DoubtService) Accept(ctx context.Context, id uint) (bool, error) {
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

func (s *DoubtService) CreateFollowUp(ctx context.Context, solutionID, userID uint, message string, imageURLs []string) (*repository.FollowUp, error) {
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

func (s *DoubtService) ListFollowUps(ctx context.Context, solutionID uint, limit, offset int) ([]repository.FollowUp, error) {
	return s.repo.ListFollowUps(ctx, solutionID, limit, offset)
}

// ── Chat ────────────────────────────────────

func (s *DoubtService) CreateRoom(ctx context.Context, questionID, studentID, mentorID uint) (*repository.ChatRoom, error) {
	existing, err := s.repo.FindRoomByQuestionID(ctx, questionID)
	if err == nil && existing != nil {
		return existing, nil
	}
	room := &repository.ChatRoom{
		QuestionID: questionID,
		StudentID:  studentID,
		MentorID:   mentorID,
		Status:     "active",
	}
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *DoubtService) GetRoom(ctx context.Context, id uint) (*repository.ChatRoom, error) {
	return s.repo.FindRoomByID(ctx, id)
}

func (s *DoubtService) GetHistory(ctx context.Context, roomID uint, limit, offset int) ([]repository.Message, error) {
	return s.repo.GetHistory(ctx, roomID, limit, offset)
}

func (s *DoubtService) SendMessage(ctx context.Context, roomID, senderID uint, content, msgType, imageURL string) (*repository.Message, error) {
	if content == "" && imageURL == "" {
		return nil, ErrValidation
	}
	if msgType == "" {
		msgType = "text"
	}
	msg := &repository.Message{
		RoomID:   roomID,
		SenderID: senderID,
		Content:  content,
		Type:     msgType,
		ImageURL: imageURL,
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	s.mu.RLock()
	channels := s.subscribers[roomID]
	s.mu.RUnlock()
	for _, ch := range channels {
		select {
		case ch <- msg:
		default:
		}
	}

	return msg, nil
}

func (s *DoubtService) Subscribe(roomID uint) <-chan *repository.Message {
	ch := make(chan *repository.Message, 100)
	s.mu.Lock()
	s.subscribers[roomID] = append(s.subscribers[roomID], ch)
	s.mu.Unlock()
	return ch
}

func (s *DoubtService) Unsubscribe(roomID uint, ch <-chan *repository.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := s.subscribers[roomID]
	for i, sub := range subs {
		if fmt.Sprintf("%p", sub) == fmt.Sprintf("%p", ch) {
			s.subscribers[roomID] = append(subs[:i], subs[i+1:]...)
			close(sub)
			break
		}
	}
}

func (s *DoubtService) MarkChatRead(ctx context.Context, roomID, userID uint) (int64, error) {
	return s.repo.MarkRead(ctx, roomID, userID)
}

// ── Notification ────────────────────────────

func (s *DoubtService) SendNotification(ctx context.Context, userID uint, nType, title, body string, channels []string) (uint, error) {
	if title == "" || body == "" {
		return 0, ErrValidation
	}
	n := &repository.NotificationRecord{
		UserID:   userID,
		Type:     nType,
		Title:    title,
		Body:     body,
		Channels: strings.Join(channels, ","),
		Read:     false,
	}
	if err := s.repo.CreateNotification(ctx, n); err != nil {
		return 0, err
	}
	return n.ID, nil
}

func (s *DoubtService) ListNotifications(ctx context.Context, userID uint, limit, offset int) ([]repository.NotificationRecord, error) {
	return s.repo.ListNotifications(ctx, userID, limit, offset)
}

func (s *DoubtService) MarkNotificationRead(ctx context.Context, userID, notificationID uint) error {
	return s.repo.MarkNotificationRead(ctx, userID, notificationID)
}

func (s *DoubtService) GetPreferences(ctx context.Context, userID uint) (*repository.NotificationPreference, error) {
	p, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return &repository.NotificationPreference{
			UserID:       userID,
			InAppEnabled: true,
			PushEnabled:  true,
		}, nil
	}
	return p, nil
}

func (s *DoubtService) UpdatePreferences(ctx context.Context, userID uint, inApp, push, email, sms, digest bool) (*repository.NotificationPreference, error) {
	p, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		p = &repository.NotificationPreference{UserID: userID}
	}
	p.InAppEnabled = inApp
	p.PushEnabled = push
	p.EmailEnabled = email
	p.SMSEnabled = sms
	p.DigestMode = digest
	if err := s.repo.UpsertPreferences(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ── helpers ─────────────────────────────────

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
