package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

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

type Repository interface {
	Create(ctx context.Context, q *Question) error
	FindByID(ctx context.Context, id uint) (*Question, error)
	List(ctx context.Context, f Filter) ([]Question, error)
	UpdateFields(ctx context.Context, id uint, fields map[string]any) error
	Search(ctx context.Context, query, subject string, limit, offset int) ([]Question, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, q *Question) error {
	return fmt.Errorf("question_repo_create: %w", r.db.WithContext(ctx).Create(q).Error)
}

func (r *repository) FindByID(ctx context.Context, id uint) (*Question, error) {
	var q Question
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil {
		return nil, fmt.Errorf("question_repo_find_by_id: %w", err)
	}
	return &q, nil
}

func (r *repository) List(ctx context.Context, f Filter) ([]Question, error) {
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
		return nil, fmt.Errorf("question_repo_list: %w", err)
	}
	return questions, nil
}

func (r *repository) UpdateFields(ctx context.Context, id uint, fields map[string]any) error {
	return fmt.Errorf("question_repo_update_fields: %w", r.db.WithContext(ctx).Model(&Question{}).Where("id = ?", id).Updates(fields).Error)
}

func (r *repository) Search(ctx context.Context, query, subject string, limit, offset int) ([]Question, error) {
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
		return nil, fmt.Errorf("question_repo_search: %w", err)
	}
	return questions, nil
}
