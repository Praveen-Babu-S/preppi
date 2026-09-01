package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Question

type Question struct {
	ID          uint `gorm:"primaryKey"`
	StudentID   uint `gorm:"index:idx_questions_student_status,priority:2;not null"`
	AssigneeID  uint
	Subject     string `gorm:"size:100;not null"`
	Topic       string `gorm:"size:100"`
	Description string `gorm:"type:text;not null"`
	ImageURLs   string `gorm:"type:text"`
	Urgency     string `gorm:"size:20;default:normal;not null"`
	Status      string `gorm:"size:20;default:open;index:idx_questions_student_status,priority:1;not null"`
	CreatedAt   int64  `gorm:"autoCreateTime"`
	UpdatedAt   int64  `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt
}

type Filter struct {
	StudentID uint
	Status    string
	Subject   string
	Query     string
	Limit     int
	Offset    int
}

// Solution

type Solution struct {
	ID          uint   `gorm:"primaryKey"`
	QuestionID  uint   `gorm:"index:idx_solutions_question;not null"`
	MentorID    uint   `gorm:"index:idx_solutions_mentor;not null"`
	Description string `gorm:"type:text;not null"`
	ImageURLs   string `gorm:"type:text"`
	Upvotes     int    `gorm:"default:0"`
	Downvotes   int    `gorm:"default:0"`
	IsAccepted  bool   `gorm:"default:false;index:idx_solutions_accepted"`
	CreatedAt   int64  `gorm:"autoCreateTime"`
	UpdatedAt   int64  `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt
}

type FollowUp struct {
	ID         uint   `gorm:"primaryKey"`
	SolutionID uint   `gorm:"index:idx_follow_ups_solution;not null"`
	UserID     uint   `gorm:"not null"`
	Message    string `gorm:"type:text;not null"`
	ImageURLs  string `gorm:"type:text"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
}

// Chat

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

// Notification

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

// Repository is the single interface for all doubt-domain database access.

type Repository interface {
	// Question
	CreateQuestion(ctx context.Context, q *Question) error
	FindQuestionByID(ctx context.Context, id uint) (*Question, error)
	ListQuestions(ctx context.Context, f Filter) ([]Question, error)
	UpdateQuestionFields(ctx context.Context, id uint, fields map[string]any) error
	SearchQuestions(ctx context.Context, query, subject string, limit, offset int) ([]Question, error)

	// Solution
	CreateSolution(ctx context.Context, s *Solution) error
	FindSolutionByID(ctx context.Context, id uint) (*Solution, error)
	ListSolutionsByQuestion(ctx context.Context, questionID uint, limit, offset int) ([]Solution, error)
	UpdateSolutionVotes(ctx context.Context, id uint, upvotes, downvotes int) error
	AcceptSolution(ctx context.Context, id uint) error
	UnacceptOtherSolutions(ctx context.Context, questionID uint) error
	CreateFollowUp(ctx context.Context, f *FollowUp) error
	ListFollowUps(ctx context.Context, solutionID uint, limit, offset int) ([]FollowUp, error)

	// Chat
	CreateRoom(ctx context.Context, r *ChatRoom) error
	FindRoomByID(ctx context.Context, id uint) (*ChatRoom, error)
	FindRoomByQuestionID(ctx context.Context, questionID uint) (*ChatRoom, error)
	CreateMessage(ctx context.Context, m *Message) error
	GetHistory(ctx context.Context, roomID uint, limit, offset int) ([]Message, error)
	MarkRead(ctx context.Context, roomID, userID uint) (int64, error)

	// Notification
	CreateNotification(ctx context.Context, n *NotificationRecord) error
	ListNotifications(ctx context.Context, userID uint, limit, offset int) ([]NotificationRecord, error)
	MarkNotificationRead(ctx context.Context, userID uint, notificationID uint) error
	GetPreferences(ctx context.Context, userID uint) (*NotificationPreference, error)
	UpsertPreferences(ctx context.Context, p *NotificationPreference) error
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ── Question ────────────────────────────────

func (r *repository) CreateQuestion(ctx context.Context, q *Question) error {
	return fmt.Errorf("doubt_repo_create_question: %w", r.db.WithContext(ctx).Create(q).Error)
}

func (r *repository) FindQuestionByID(ctx context.Context, id uint) (*Question, error) {
	var q Question
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil {
		return nil, fmt.Errorf("doubt_repo_find_question: %w", err)
	}
	return &q, nil
}

func (r *repository) ListQuestions(ctx context.Context, f Filter) ([]Question, error) {
	q := r.db.WithContext(ctx)
	if f.StudentID != 0 {
		q = q.Where("student_id = ?", f.StudentID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}

	var questions []Question
	if err := q.Order("created_at DESC").Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("doubt_repo_list_questions: %w", err)
	}
	return questions, nil
}

func (r *repository) UpdateQuestionFields(ctx context.Context, id uint, fields map[string]any) error {
	return fmt.Errorf("doubt_repo_update_question: %w",
		r.db.WithContext(ctx).Model(&Question{}).Where("id = ?", id).Updates(fields).Error)
}

func (r *repository) SearchQuestions(ctx context.Context, query, subject string, limit, offset int) ([]Question, error) {
	q := r.db.WithContext(ctx)
	qs := []string{}
	args := []any{}
	if query != "" {
		qs = append(qs, "description ILIKE ?")
		args = append(args, "%"+query+"%")
	}
	if subject != "" {
		qs = append(qs, "subject = ?")
		args = append(args, subject)
	}
	if len(qs) > 0 {
		q = q.Where(strings.Join(qs, " AND "), args...)
	}

	var questions []Question
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("doubt_repo_search_questions: %w", err)
	}
	return questions, nil
}

// ── Solution ────────────────────────────────

func (r *repository) CreateSolution(ctx context.Context, s *Solution) error {
	return fmt.Errorf("doubt_repo_create_solution: %w", r.db.WithContext(ctx).Create(s).Error)
}

func (r *repository) FindSolutionByID(ctx context.Context, id uint) (*Solution, error) {
	var s Solution
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		return nil, fmt.Errorf("doubt_repo_find_solution: %w", err)
	}
	return &s, nil
}

func (r *repository) ListSolutionsByQuestion(ctx context.Context, questionID uint, limit, offset int) ([]Solution, error) {
	var solutions []Solution
	err := r.db.WithContext(ctx).
		Where("question_id = ?", questionID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&solutions).Error
	if err != nil {
		return nil, fmt.Errorf("doubt_repo_list_solutions: %w", err)
	}
	return solutions, nil
}

func (r *repository) UpdateSolutionVotes(ctx context.Context, id uint, upvotes, downvotes int) error {
	return fmt.Errorf("doubt_repo_update_votes: %w",
		r.db.WithContext(ctx).Model(&Solution{}).Where("id = ?", id).
			Updates(map[string]any{"upvotes": upvotes, "downvotes": downvotes}).Error)
}

func (r *repository) AcceptSolution(ctx context.Context, id uint) error {
	return fmt.Errorf("doubt_repo_accept_solution: %w",
		r.db.WithContext(ctx).Model(&Solution{}).Where("id = ?", id).
			Update("is_accepted", true).Error)
}

func (r *repository) UnacceptOtherSolutions(ctx context.Context, questionID uint) error {
	return fmt.Errorf("doubt_repo_unaccept_others: %w",
		r.db.WithContext(ctx).Model(&Solution{}).
			Where("question_id = ? AND is_accepted = ?", questionID, true).
			Update("is_accepted", false).Error)
}

func (r *repository) CreateFollowUp(ctx context.Context, f *FollowUp) error {
	return fmt.Errorf("doubt_repo_create_follow_up: %w", r.db.WithContext(ctx).Create(f).Error)
}

func (r *repository) ListFollowUps(ctx context.Context, solutionID uint, limit, offset int) ([]FollowUp, error) {
	var followUps []FollowUp
	err := r.db.WithContext(ctx).
		Where("solution_id = ?", solutionID).
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&followUps).Error
	if err != nil {
		return nil, fmt.Errorf("doubt_repo_list_follow_ups: %w", err)
	}
	return followUps, nil
}

// ── Chat ────────────────────────────────────

func (r *repository) CreateRoom(ctx context.Context, room *ChatRoom) error {
	return fmt.Errorf("doubt_repo_create_room: %w", r.db.WithContext(ctx).Create(room).Error)
}

func (r *repository) FindRoomByID(ctx context.Context, id uint) (*ChatRoom, error) {
	var room ChatRoom
	if err := r.db.WithContext(ctx).First(&room, id).Error; err != nil {
		return nil, fmt.Errorf("doubt_repo_find_room: %w", err)
	}
	return &room, nil
}

func (r *repository) FindRoomByQuestionID(ctx context.Context, questionID uint) (*ChatRoom, error) {
	var room ChatRoom
	if err := r.db.WithContext(ctx).Where("question_id = ?", questionID).First(&room).Error; err != nil {
		return nil, fmt.Errorf("doubt_repo_find_room_by_question: %w", err)
	}
	return &room, nil
}

func (r *repository) CreateMessage(ctx context.Context, m *Message) error {
	return fmt.Errorf("doubt_repo_create_message: %w", r.db.WithContext(ctx).Create(m).Error)
}

func (r *repository) GetHistory(ctx context.Context, roomID uint, limit, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("doubt_repo_get_history: %w", err)
	}
	return messages, nil
}

func (r *repository) MarkRead(ctx context.Context, roomID, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Model(&Message{}).
		Where("room_id = ? AND sender_id != ? AND read = ?", roomID, userID, false).
		Update("read", true)
	if result.Error != nil {
		return 0, fmt.Errorf("doubt_repo_mark_read: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ── Notification ────────────────────────────

func (r *repository) CreateNotification(ctx context.Context, n *NotificationRecord) error {
	return fmt.Errorf("doubt_repo_create_notification: %w", r.db.WithContext(ctx).Create(n).Error)
}

func (r *repository) ListNotifications(ctx context.Context, userID uint, limit, offset int) ([]NotificationRecord, error) {
	var notifications []NotificationRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("doubt_repo_list_notifications: %w", err)
	}
	return notifications, nil
}

func (r *repository) MarkNotificationRead(ctx context.Context, userID uint, notificationID uint) error {
	return fmt.Errorf("doubt_repo_mark_notification_read: %w",
		r.db.WithContext(ctx).Model(&NotificationRecord{}).
			Where("id = ? AND user_id = ?", notificationID, userID).
			Update("read", true).Error)
}

func (r *repository) GetPreferences(ctx context.Context, userID uint) (*NotificationPreference, error) {
	var p NotificationPreference
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, fmt.Errorf("doubt_repo_get_prefs: %w", err)
	}
	return &p, nil
}

func (r *repository) UpsertPreferences(ctx context.Context, p *NotificationPreference) error {
	return fmt.Errorf("doubt_repo_upsert_prefs: %w", r.db.WithContext(ctx).Save(p).Error)
}
