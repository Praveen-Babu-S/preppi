package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Assignment struct {
	ID          uint   `gorm:"primaryKey"`
	QuestionID  uint   `gorm:"index;not null"`
	MentorID    uint   `gorm:"index;not null"`
	Status      string `gorm:"size:20;not null;default:pending"` // pending, skipped, in_progress, completed
	AssignedAt  int64  `gorm:"autoCreateTime"`
	RespondedAt int64  // 0 until mentor responds
	DeletedAt   gorm.DeletedAt
}

type Escalation struct {
	ID              uint   `gorm:"primaryKey"`
	QuestionID      uint   `gorm:"index;not null"`
	EscalationLevel int    `gorm:"not null;default:1"`
	Reason          string `gorm:"type:text;not null"`
	CreatedAt       int64  `gorm:"autoCreateTime"`
}

type Repository interface {
	CreateAssignment(ctx context.Context, a *Assignment) error
	GetAssignmentByQuestion(ctx context.Context, questionID uint) (*Assignment, error)
	GetAssignmentByID(ctx context.Context, id uint) (*Assignment, error)
	UpdateAssignmentStatus(ctx context.Context, id uint, status string, respondedAt time.Time) error
	GetPendingCount(ctx context.Context, mentorID uint) (int, error)
	GetPendingForMentor(ctx context.Context, mentorID uint, limit, offset int) ([]Assignment, error)
	CreateEscalation(ctx context.Context, e *Escalation) error
	GetLatestEscalation(ctx context.Context, questionID uint) (*Escalation, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateAssignment(ctx context.Context, a *Assignment) error {
	return fmt.Errorf("matching_repo_create_assignment: %w", r.db.WithContext(ctx).Create(a).Error)
}

func (r *repository) GetAssignmentByQuestion(ctx context.Context, questionID uint) (*Assignment, error) {
	var a Assignment
	if err := r.db.WithContext(ctx).Where("question_id = ?", questionID).Order("assigned_at DESC").First(&a).Error; err != nil {
		return nil, fmt.Errorf("matching_repo_find_assignment: %w", err)
	}
	return &a, nil
}

func (r *repository) GetAssignmentByID(ctx context.Context, id uint) (*Assignment, error) {
	var a Assignment
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		return nil, fmt.Errorf("matching_repo_find_assignment_by_id: %w", err)
	}
	return &a, nil
}

func (r *repository) UpdateAssignmentStatus(ctx context.Context, id uint, status string, respondedAt time.Time) error {
	updates := map[string]any{"status": status}
	if !respondedAt.IsZero() {
		updates["responded_at"] = respondedAt.Unix()
	}
	return fmt.Errorf("matching_repo_update_status: %w",
		r.db.WithContext(ctx).Model(&Assignment{}).Where("id = ?", id).Updates(updates).Error)
}

func (r *repository) GetPendingCount(ctx context.Context, mentorID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Assignment{}).
		Where("mentor_id = ? AND status = 'pending'", mentorID).
		Count(&count).Error
	return int(count), fmt.Errorf("matching_repo_pending_count: %w", err)
}

func (r *repository) GetPendingForMentor(ctx context.Context, mentorID uint, limit, offset int) ([]Assignment, error) {
	var assignments []Assignment
	err := r.db.WithContext(ctx).
		Where("mentor_id = ? AND status = 'pending'", mentorID).
		Order("assigned_at ASC").
		Limit(limit).Offset(offset).
		Find(&assignments).Error
	return assignments, fmt.Errorf("matching_repo_pending_for_mentor: %w", err)
}

func (r *repository) CreateEscalation(ctx context.Context, e *Escalation) error {
	return fmt.Errorf("matching_repo_create_escalation: %w", r.db.WithContext(ctx).Create(e).Error)
}

func (r *repository) GetLatestEscalation(ctx context.Context, questionID uint) (*Escalation, error) {
	var e Escalation
	err := r.db.WithContext(ctx).
		Where("question_id = ?", questionID).
		Order("created_at DESC").
		First(&e).Error
	if err != nil {
		return &Escalation{EscalationLevel: 0}, fmt.Errorf("matching_repo_get_escalation: %w", err)
	}
	return &e, nil
}
