package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	ID            uint   `gorm:"primaryKey"`
	Email         string `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash  string `gorm:"size:255;not null"`
	Role          string `gorm:"size:20;not null"`
	Subject       string `gorm:"size:100"`
	EmailVerified bool   `gorm:"default:false"`
	CreatedAt     int64  `gorm:"autoCreateTime"`
	UpdatedAt     int64  `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt
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
	return fmt.Errorf("auth_repository_create: %w", r.db.WithContext(ctx).Create(u).Error)
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, fmt.Errorf("auth_repository_find_by_email: %w", err)
	}
	return &u, nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, fmt.Errorf("auth_repository_find_by_id: %w", err)
	}
	return &u, nil
}

func (r *repository) Update(ctx context.Context, u *User) error {
	return fmt.Errorf("auth_repository_update: %w", r.db.WithContext(ctx).Save(u).Error)
}

func (r *repository) MarkEmailVerified(ctx context.Context, id uint) error {
	return fmt.Errorf("auth_repository_mark_email_verified: %w",
		r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("email_verified", true).Error)
}
