package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type NotificationRecord struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index:idx_notifications_user;not null"`
	Type      string `gorm:"size:50;not null"`
	Title     string `gorm:"size:255;not null"`
	Body      string `gorm:"type:text;not null"`
	Data      string `gorm:"type:jsonb"`
	Channels  string `gorm:"type:text"`
	Read      bool   `gorm:"default:false;index:idx_notifications_user_read"`
	CreatedAt int64  `gorm:"autoCreateTime"`
}

type NotificationPreference struct {
	ID           uint  `gorm:"primaryKey"`
	UserID       uint  `gorm:"uniqueIndex:idx_notification_prefs_user;not null"`
	InAppEnabled bool  `gorm:"default:true"`
	PushEnabled  bool  `gorm:"default:true"`
	EmailEnabled bool  `gorm:"default:false"`
	SMSEnabled   bool  `gorm:"default:false"`
	DigestMode   bool  `gorm:"default:false"`
	CreatedAt    int64 `gorm:"autoCreateTime"`
	UpdatedAt    int64 `gorm:"autoUpdateTime"`
}

type Repository interface {
	CreateNotification(ctx context.Context, n *NotificationRecord) error
	ListNotifications(ctx context.Context, userID uint, limit, offset int) ([]NotificationRecord, error)
	MarkRead(ctx context.Context, userID uint, notificationID uint) error
	GetPreferences(ctx context.Context, userID uint) (*NotificationPreference, error)
	UpsertPreferences(ctx context.Context, p *NotificationPreference) error
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateNotification(ctx context.Context, n *NotificationRecord) error {
	return fmt.Errorf("notification_repo_create: %w", r.db.WithContext(ctx).Create(n).Error)
}

func (r *repository) ListNotifications(ctx context.Context, userID uint, limit, offset int) ([]NotificationRecord, error) {
	var notifications []NotificationRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("notification_repo_list: %w", err)
	}
	return notifications, nil
}

func (r *repository) MarkRead(ctx context.Context, userID uint, notificationID uint) error {
	return fmt.Errorf("notification_repo_mark_read: %w",
		r.db.WithContext(ctx).Model(&NotificationRecord{}).
			Where("id = ? AND user_id = ?", notificationID, userID).
			Update("read", true).Error)
}

func (r *repository) GetPreferences(ctx context.Context, userID uint) (*NotificationPreference, error) {
	var p NotificationPreference
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, fmt.Errorf("notification_repo_get_prefs: %w", err)
	}
	return &p, nil
}

func (r *repository) UpsertPreferences(ctx context.Context, p *NotificationPreference) error {
	return fmt.Errorf("notification_repo_upsert_prefs: %w", r.db.WithContext(ctx).Save(p).Error)
}
