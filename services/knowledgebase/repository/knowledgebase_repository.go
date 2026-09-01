package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Article struct {
	ID        uint   `gorm:"primaryKey"`
	Subject   string `gorm:"size:100;not null"`
	Topic     string `gorm:"size:100;not null"`
	Title     string `gorm:"size:255;not null"`
	Summary   string `gorm:"type:text"`
	FullText  string `gorm:"type:text"`
	Upvotes   int    `gorm:"default:0"`
	Tags      string `gorm:"type:text"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt
}

type Topic struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"size:100;not null"`
	Subject      string `gorm:"size:100;not null"`
	ArticleCount int    `gorm:"default:0"`
	CreatedAt    int64  `gorm:"autoCreateTime"`
}

type Repository interface {
	SearchArticles(ctx context.Context, query, subject string, limit, offset int) ([]Article, error)
	GetArticle(ctx context.Context, id uint) (*Article, error)
	GetRelatedTopics(ctx context.Context, topic, subject string) ([]Topic, error)
	SuggestKeywords(ctx context.Context, query string) ([]string, error)
	IncrementUpvotes(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) SearchArticles(ctx context.Context, query, subject string, limit, offset int) ([]Article, error) {
	q := r.db.WithContext(ctx)
	conditions := []string{}
	args := []any{}
	if query != "" {
		conditions = append(conditions, "(title ILIKE ? OR summary ILIKE ? OR tags ILIKE ?)")
		qArg := "%" + query + "%"
		args = append(args, qArg, qArg, qArg)
	}
	if subject != "" {
		conditions = append(conditions, "subject = ?")
		args = append(args, subject)
	}
	if len(conditions) > 0 {
		q = q.Where(strings.Join(conditions, " AND "), args...)
	}
	var articles []Article
	err := q.Order("upvotes DESC, created_at DESC").Limit(limit).Offset(offset).Find(&articles).Error
	if err != nil {
		return nil, fmt.Errorf("kb_repo_search: %w", err)
	}
	return articles, nil
}

func (r *repository) GetArticle(ctx context.Context, id uint) (*Article, error) {
	var a Article
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		return nil, fmt.Errorf("kb_repo_get_article: %w", err)
	}
	return &a, nil
}

func (r *repository) GetRelatedTopics(ctx context.Context, topic, subject string) ([]Topic, error) {
	var topics []Topic
	q := r.db.WithContext(ctx)
	if subject != "" {
		q = q.Where("subject = ? AND name != ?", subject, topic)
	} else {
		q = q.Where("name != ?", topic)
	}
	err := q.Order("article_count DESC").Limit(10).Find(&topics).Error
	if err != nil {
		return nil, fmt.Errorf("kb_repo_get_related_topics: %w", err)
	}
	return topics, nil
}

func (r *repository) SuggestKeywords(ctx context.Context, query string) ([]string, error) {
	var titles []string
	err := r.db.WithContext(ctx).Model(&Article{}).
		Where("title ILIKE ?", "%"+query+"%").
		Limit(10).Pluck("title", &titles).Error
	if err != nil {
		return nil, fmt.Errorf("kb_repo_suggest_keywords: %w", err)
	}
	return titles, nil
}

func (r *repository) IncrementUpvotes(ctx context.Context, id uint) error {
	return fmt.Errorf("kb_repo_increment_upvotes: %w",
		r.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).
			Update("upvotes", gorm.Expr("upvotes + 1")).Error)
}
