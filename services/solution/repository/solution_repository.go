package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

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

type Repository interface {
	CreateSolution(ctx context.Context, s *Solution) error
	FindSolutionByID(ctx context.Context, id uint) (*Solution, error)
	ListSolutionsByQuestion(ctx context.Context, questionID uint, limit, offset int) ([]Solution, error)
	UpdateSolutionVotes(ctx context.Context, id uint, upvotes, downvotes int) error
	AcceptSolution(ctx context.Context, id uint) error
	UnacceptOtherSolutions(ctx context.Context, questionID uint) error
	CreateFollowUp(ctx context.Context, f *FollowUp) error
	ListFollowUps(ctx context.Context, solutionID uint, limit, offset int) ([]FollowUp, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateSolution(ctx context.Context, s *Solution) error {
	return fmt.Errorf("solution_repo_create: %w", r.db.WithContext(ctx).Create(s).Error)
}

func (r *repository) FindSolutionByID(ctx context.Context, id uint) (*Solution, error) {
	var s Solution
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		return nil, fmt.Errorf("solution_repo_find_by_id: %w", err)
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
		return nil, fmt.Errorf("solution_repo_list_by_question: %w", err)
	}
	return solutions, nil
}

func (r *repository) UpdateSolutionVotes(ctx context.Context, id uint, upvotes, downvotes int) error {
	return fmt.Errorf("solution_repo_update_votes: %w",
		r.db.WithContext(ctx).Model(&Solution{}).Where("id = ?", id).
			Updates(map[string]any{"upvotes": upvotes, "downvotes": downvotes}).Error)
}

func (r *repository) AcceptSolution(ctx context.Context, id uint) error {
	return fmt.Errorf("solution_repo_accept: %w",
		r.db.WithContext(ctx).Model(&Solution{}).Where("id = ?", id).
			Update("is_accepted", true).Error)
}

func (r *repository) UnacceptOtherSolutions(ctx context.Context, questionID uint) error {
	return fmt.Errorf("solution_repo_unaccept_others: %w",
		r.db.WithContext(ctx).Model(&Solution{}).
			Where("question_id = ? AND is_accepted = ?", questionID, true).
			Update("is_accepted", false).Error)
}

func (r *repository) CreateFollowUp(ctx context.Context, f *FollowUp) error {
	return fmt.Errorf("solution_repo_create_follow_up: %w", r.db.WithContext(ctx).Create(f).Error)
}

func (r *repository) ListFollowUps(ctx context.Context, solutionID uint, limit, offset int) ([]FollowUp, error) {
	var followUps []FollowUp
	err := r.db.WithContext(ctx).
		Where("solution_id = ?", solutionID).
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&followUps).Error
	if err != nil {
		return nil, fmt.Errorf("solution_repo_list_follow_ups: %w", err)
	}
	return followUps, nil
}
