package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"preppi.com/services/chat/repository"
)

var (
	ErrRoomNotFound = errors.New("chat room not found")
	ErrValidation   = errors.New("validation error")
)

type ChatService struct {
	repo        repository.Repository
	mu          sync.RWMutex
	subscribers map[uint][]chan *repository.Message
}

func New(repo repository.Repository) *ChatService {
	return &ChatService{
		repo:        repo,
		subscribers: make(map[uint][]chan *repository.Message),
	}
}

func (s *ChatService) CreateRoom(ctx context.Context, questionID, studentID, mentorID uint) (*repository.ChatRoom, error) {
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

func (s *ChatService) GetRoom(ctx context.Context, id uint) (*repository.ChatRoom, error) {
	return s.repo.FindRoomByID(ctx, id)
}

func (s *ChatService) GetHistory(ctx context.Context, roomID uint, limit, offset int) ([]repository.Message, error) {
	return s.repo.GetHistory(ctx, roomID, limit, offset)
}

func (s *ChatService) SendMessage(ctx context.Context, roomID, senderID uint, content, msgType, imageURL string) (*repository.Message, error) {
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

func (s *ChatService) Subscribe(roomID uint) <-chan *repository.Message {
	ch := make(chan *repository.Message, 100)
	s.mu.Lock()
	s.subscribers[roomID] = append(s.subscribers[roomID], ch)
	s.mu.Unlock()
	return ch
}

func (s *ChatService) Unsubscribe(roomID uint, ch <-chan *repository.Message) {
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

func (s *ChatService) MarkRead(ctx context.Context, roomID, userID uint) (int64, error) {
	return s.repo.MarkRead(ctx, roomID, userID)
}

func SplitImageURLs(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
