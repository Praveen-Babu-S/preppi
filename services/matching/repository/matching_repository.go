package repository

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
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
	err := r.db.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Error().Err(err).Uint("question_id", a.QuestionID).Uint("mentor_id", a.MentorID).Msg("unable to create assignment")
	}
	return err
}

func (r *repository) GetAssignmentByQuestion(ctx context.Context, questionID uint) (*Assignment, error) {
	var a Assignment
	if err := r.db.WithContext(ctx).Where("question_id = ?", questionID).Order("assigned_at DESC").First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Uint("question_id", questionID).Msg("assignment not found for question")
		} else {
			log.Error().Err(err).Uint("question_id", questionID).Msg("unable to find assignment")
		}
		return nil, err
	}
	return &a, nil
}

func (r *repository) GetAssignmentByID(ctx context.Context, id uint) (*Assignment, error) {
	var a Assignment
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Uint("id", id).Msg("assignment not found")
		} else {
			log.Error().Err(err).Uint("id", id).Msg("unable to find assignment by id")
		}
		return nil, err
	}
	return &a, nil
}

func (r *repository) UpdateAssignmentStatus(ctx context.Context, id uint, status string, respondedAt time.Time) error {
	updates := map[string]any{"status": status}
	if !respondedAt.IsZero() {
		updates["responded_at"] = respondedAt.Unix()
	}
	err := r.db.WithContext(ctx).Model(&Assignment{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		log.Error().Err(err).Uint("id", id).Str("status", status).Msg("unable to update assignment status")
	}
	return err
}

func (r *repository) GetPendingCount(ctx context.Context, mentorID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Assignment{}).
		Where("mentor_id = ? AND status = 'pending'", mentorID).
		Count(&count).Error
	if err != nil {
		log.Error().Err(err).Uint("mentor_id", mentorID).Msg("unable to count pending assignments")
		return 0, err
	}
	return int(count), nil
}

func (r *repository) GetPendingForMentor(ctx context.Context, mentorID uint, limit, offset int) ([]Assignment, error) {
	var assignments []Assignment
	err := r.db.WithContext(ctx).
		Where("mentor_id = ? AND status = 'pending'", mentorID).
		Order("assigned_at ASC").
		Limit(limit).Offset(offset).
		Find(&assignments).Error
	if err != nil {
		log.Error().Err(err).Uint("mentor_id", mentorID).Msg("unable to list pending assignments")
		return nil, err
	}
	return assignments, nil
}

func (r *repository) CreateEscalation(ctx context.Context, e *Escalation) error {
	err := r.db.WithContext(ctx).Create(e).Error
	if err != nil {
		log.Error().Err(err).Uint("question_id", e.QuestionID).Int("level", e.EscalationLevel).Msg("unable to create escalation")
	}
	return err
}

func (r *repository) GetLatestEscalation(ctx context.Context, questionID uint) (*Escalation, error) {
	var e Escalation
	err := r.db.WithContext(ctx).
		Where("question_id = ?", questionID).
		Order("created_at DESC").
		First(&e).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Uint("question_id", questionID).Msg("no escalation found")
			return nil, nil
		}
		log.Error().Err(err).Uint("question_id", questionID).Msg("unable to get latest escalation")
		return nil, err
	}
	return &e, nil
}
