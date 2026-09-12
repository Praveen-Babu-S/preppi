package repository

import (
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email         string `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash  string `gorm:"size:255;not null"`
	Role          string `gorm:"size:20;not null"`
	Subject       string `gorm:"size:100"`
	EmailVerified bool   `gorm:"default:false"`
}

type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, u *User) error
	MarkEmailVerified(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, u *User) error {
	err := r.db.WithContext(ctx).Create(u).Error
	if err != nil {
		log.Error().Err(err).Msg("failed to create user")
	}
	return err
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Str("email", email).Msg("user not found")
		} else {
			log.Error().Err(err).Str("email", email).Msg("unable to find user with email")
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		log.Error().Err(err).Uint("id", id).Msg("unable to find user with id")
		return nil, err
	}
	return &u, nil
}

func (r *repository) Update(ctx context.Context, u *User) error {
	if err := r.db.WithContext(ctx).Save(u).Error; err != nil {
		log.Error().Err(err).Uint("id", u.ID).Msg("unable to save user")
		return err
	}
	return nil
}

func (r *repository) MarkEmailVerified(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("email_verified", true).Error; err != nil {
		log.Error().Err(err).Uint("id", id).Msg("unable to mark user verified")
		return err
	}
	return nil
}
