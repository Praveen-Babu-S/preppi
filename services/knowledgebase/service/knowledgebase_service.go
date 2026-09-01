package service

import (
	"context"
	"errors"

	"preppi.com/services/knowledgebase/repository"
)

var (
	ErrArticleNotFound = errors.New("article not found")
	ErrValidation      = errors.New("validation error")
)

type KnowledgeBaseService struct {
	repo repository.Repository
}

func New(repo repository.Repository) *KnowledgeBaseService {
	return &KnowledgeBaseService{repo: repo}
}

func (s *KnowledgeBaseService) Search(ctx context.Context, query, subject string, limit, offset int) ([]repository.Article, error) {
	return s.repo.SearchArticles(ctx, query, subject, limit, offset)
}

func (s *KnowledgeBaseService) GetArticle(ctx context.Context, id uint) (*repository.Article, error) {
	return s.repo.GetArticle(ctx, id)
}

func (s *KnowledgeBaseService) GetRelatedTopics(ctx context.Context, topic, subject string) ([]repository.Topic, error) {
	return s.repo.GetRelatedTopics(ctx, topic, subject)
}

func (s *KnowledgeBaseService) SuggestKeywords(ctx context.Context, query string) ([]string, error) {
	return s.repo.SuggestKeywords(ctx, query)
}
