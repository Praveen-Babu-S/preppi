package repository

import (
	"context"

	"github.com/rs/zerolog/log"
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
	err := r.db.WithContext(ctx).Create(p).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", p.UserID).Msg("unable to create profile")
	}
	return err
}

func (r *repository) UpsertMentorProfile(ctx context.Context, m *MentorProfile) error {
	err := r.db.WithContext(ctx).Save(m).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", m.UserID).Msg("unable to upsert mentor profile")
	}
	return err
}

func (r *repository) GetProfile(ctx context.Context, userID uint) (*Profile, error) {
	var p Profile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Uint("user_id", userID).Msg("profile not found")
		} else {
			log.Error().Err(err).Uint("user_id", userID).Msg("unable to find profile")
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetMentorProfile(ctx context.Context, userID uint) (*MentorProfile, error) {
	var m MentorProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Uint("user_id", userID).Msg("mentor profile not found")
		} else {
			log.Error().Err(err).Uint("user_id", userID).Msg("unable to find mentor profile")
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) UpdateProfile(ctx context.Context, p *Profile) error {
	err := r.db.WithContext(ctx).Save(p).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", p.UserID).Msg("unable to update profile")
	}
	return err
}

func (r *repository) GetMentorsBySubject(ctx context.Context, subject string, limit, offset int) ([]MentorProfile, error) {
	var mentors []MentorProfile
	err := r.db.WithContext(ctx).
		Where("expertise_subjects LIKE ? AND verification_status = ?", "%"+subject+"%", "approved").
		Limit(limit).Offset(offset).Find(&mentors).Error
	if err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("unable to find mentors by subject")
		return nil, err
	}
	return mentors, nil
}

func (r *repository) SetOnline(ctx context.Context, userID uint, online bool) error {
	err := r.db.WithContext(ctx).Model(&Profile{}).Where("user_id = ?", userID).Update("online", online).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Bool("online", online).Msg("unable to update online status")
	}
	return err
}

func (r *repository) ApproveMentor(ctx context.Context, userID uint, status string) error {
	err := r.db.WithContext(ctx).Model(&MentorProfile{}).Where("user_id = ?", userID).Update("verification_status", status).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Str("status", status).Msg("unable to approve mentor")
	}
	return err
}

func (r *repository) GetPendingMentors(ctx context.Context, limit, offset int) ([]MentorProfile, error) {
	var mentors []MentorProfile
	err := r.db.WithContext(ctx).
		Where("verification_status = ?", "pending").
		Limit(limit).Offset(offset).Find(&mentors).Error
	if err != nil {
		log.Error().Err(err).Msg("unable to list pending mentors")
		return nil, err
	}
	return mentors, nil
}
