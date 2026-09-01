package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type ChatRoom struct {
	ID         uint   `gorm:"primaryKey"`
	QuestionID uint   `gorm:"uniqueIndex:idx_chat_rooms_question;not null"`
	StudentID  uint   `gorm:"index:idx_chat_rooms_student;not null"`
	MentorID   uint   `gorm:"index:idx_chat_rooms_mentor;not null"`
	Status     string `gorm:"size:20;default:active;not null"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
	UpdatedAt  int64  `gorm:"autoUpdateTime"`
}

type Message struct {
	ID        uint   `gorm:"primaryKey"`
	RoomID    uint   `gorm:"index:idx_messages_room;not null"`
	SenderID  uint   `gorm:"index:idx_messages_sender;not null"`
	Content   string `gorm:"type:text;not null"`
	Type      string `gorm:"size:20;default:text;not null"`
	ImageURL  string `gorm:"size:500"`
	Read      bool   `gorm:"default:false"`
	CreatedAt int64  `gorm:"autoCreateTime"`
}

type Repository interface {
	CreateRoom(ctx context.Context, r *ChatRoom) error
	FindRoomByID(ctx context.Context, id uint) (*ChatRoom, error)
	FindRoomByQuestionID(ctx context.Context, questionID uint) (*ChatRoom, error)
	CreateMessage(ctx context.Context, m *Message) error
	GetHistory(ctx context.Context, roomID uint, limit, offset int) ([]Message, error)
	MarkRead(ctx context.Context, roomID, userID uint) (int64, error)
	UnreadCount(ctx context.Context, roomID, userID uint) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateRoom(ctx context.Context, room *ChatRoom) error {
	return fmt.Errorf("chat_repo_create_room: %w", r.db.WithContext(ctx).Create(room).Error)
}

func (r *repository) FindRoomByID(ctx context.Context, id uint) (*ChatRoom, error) {
	var room ChatRoom
	if err := r.db.WithContext(ctx).First(&room, id).Error; err != nil {
		return nil, fmt.Errorf("chat_repo_find_room_by_id: %w", err)
	}
	return &room, nil
}

func (r *repository) FindRoomByQuestionID(ctx context.Context, questionID uint) (*ChatRoom, error) {
	var room ChatRoom
	if err := r.db.WithContext(ctx).Where("question_id = ?", questionID).First(&room).Error; err != nil {
		return nil, fmt.Errorf("chat_repo_find_room_by_question: %w", err)
	}
	return &room, nil
}

func (r *repository) CreateMessage(ctx context.Context, m *Message) error {
	return fmt.Errorf("chat_repo_create_message: %w", r.db.WithContext(ctx).Create(m).Error)
}

func (r *repository) GetHistory(ctx context.Context, roomID uint, limit, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("chat_repo_get_history: %w", err)
	}
	return messages, nil
}

func (r *repository) MarkRead(ctx context.Context, roomID, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Model(&Message{}).
		Where("room_id = ? AND sender_id != ? AND read = ?", roomID, userID, false).
		Update("read", true)
	if result.Error != nil {
		return 0, fmt.Errorf("chat_repo_mark_read: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *repository) UnreadCount(ctx context.Context, roomID, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Message{}).
		Where("room_id = ? AND sender_id != ? AND read = ?", roomID, userID, false).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("chat_repo_unread_count: %w", err)
	}
	return count, nil
}
