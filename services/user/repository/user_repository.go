package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Profile struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"uniqueIndex;not null"`
	Name      string `gorm:"size:255;not null"`
	AvatarURL string `gorm:"size:500"`
	Phone     string `gorm:"size:20"`
	School    string `gorm:"size:255"`
	College   string `gorm:"size:255"`
	Bio       string `gorm:"type:text"`
	Role      string `gorm:"size:20;not null"`
	Online    bool   `gorm:"default:false"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt
}

type MentorProfile struct {
	ID                 uint   `gorm:"primaryKey"`
	UserID             uint   `gorm:"uniqueIndex;not null"`
	ExpertiseSubjects  string `gorm:"type:text"`
	SubTopics          string `gorm:"type:text"`
	VerificationStatus string `gorm:"size:20;default:pending"`
	Rating             float64
	QuestionsAnswered  int
	CreatedAt          int64 `gorm:"autoCreateTime"`
	UpdatedAt          int64 `gorm:"autoUpdateTime"`
}

type Repository interface {
	CreateProfile(ctx context.Context, p *Profile) error
	UpsertMentorProfile(ctx context.Context, m *MentorProfile) error
	GetProfile(ctx context.Context, userID uint) (*Profile, error)
	GetMentorProfile(ctx context.Context, userID uint) (*MentorProfile, error)
	UpdateProfile(ctx context.Context, p *Profile) error
	GetMentorsBySubject(ctx context.Context, subject string, limit, offset int) ([]MentorProfile, error)
	SetOnline(ctx context.Context, userID uint, online bool) error
	ApproveMentor(ctx context.Context, userID uint, status string) error
	GetPendingMentors(ctx context.Context, limit, offset int) ([]MentorProfile, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateProfile(ctx context.Context, p *Profile) error {
	return fmt.Errorf("user_repo_create_profile: %w", r.db.WithContext(ctx).Create(p).Error)
}

func (r *repository) UpsertMentorProfile(ctx context.Context, m *MentorProfile) error {
	return fmt.Errorf("user_repo_upsert_mentor: %w", r.db.WithContext(ctx).Save(m).Error)
}

func (r *repository) GetProfile(ctx context.Context, userID uint) (*Profile, error) {
	var p Profile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, fmt.Errorf("user_repo_get_profile: %w", err)
	}
	return &p, nil
}

func (r *repository) GetMentorProfile(ctx context.Context, userID uint) (*MentorProfile, error) {
	var m MentorProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error; err != nil {
		return nil, fmt.Errorf("user_repo_get_mentor: %w", err)
	}
	return &m, nil
}

func (r *repository) UpdateProfile(ctx context.Context, p *Profile) error {
	return fmt.Errorf("user_repo_update_profile: %w", r.db.WithContext(ctx).Save(p).Error)
}

func (r *repository) GetMentorsBySubject(ctx context.Context, subject string, limit, offset int) ([]MentorProfile, error) {
	var mentors []MentorProfile
	err := r.db.WithContext(ctx).
		Where("expertise_subjects LIKE ? AND verification_status = ?", "%"+subject+"%", "approved").
		Limit(limit).Offset(offset).Find(&mentors).Error
	if err != nil {
		return nil, fmt.Errorf("user_repo_get_mentors_by_subject: %w", err)
	}
	return mentors, nil
}

func (r *repository) SetOnline(ctx context.Context, userID uint, online bool) error {
	return fmt.Errorf("user_repo_set_online: %w",
		r.db.WithContext(ctx).Model(&Profile{}).Where("user_id = ?", userID).Update("online", online).Error)
}

func (r *repository) ApproveMentor(ctx context.Context, userID uint, status string) error {
	return fmt.Errorf("user_repo_approve_mentor: %w",
		r.db.WithContext(ctx).Model(&MentorProfile{}).Where("user_id = ?", userID).Update("verification_status", status).Error)
}

func (r *repository) GetPendingMentors(ctx context.Context, limit, offset int) ([]MentorProfile, error) {
	var mentors []MentorProfile
	err := r.db.WithContext(ctx).
		Where("verification_status = ?", "pending").
		Limit(limit).Offset(offset).Find(&mentors).Error
	if err != nil {
		return nil, fmt.Errorf("user_repo_get_pending_mentors: %w", err)
	}
	return mentors, nil
}
