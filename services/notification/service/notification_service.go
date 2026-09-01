package service

import (
	"context"
	"errors"
	"strings"

	"preppi.com/services/notification/repository"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrValidation           = errors.New("validation error")
)

type NotificationService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Send(ctx context.Context, userID uint, nType, title, body string, channels []string) (uint, error) {
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

func (s *NotificationService) List(ctx context.Context, userID uint, limit, offset int) ([]repository.NotificationRecord, error) {
	return s.repo.ListNotifications(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, notificationID uint) error {
	return s.repo.MarkRead(ctx, userID, notificationID)
}

func (s *NotificationService) GetPreferences(ctx context.Context, userID uint) (*repository.NotificationPreference, error) {
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

func (s *NotificationService) UpdatePreferences(ctx context.Context, userID uint, inApp, push, email, sms, digest bool) (*repository.NotificationPreference, error) {
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
