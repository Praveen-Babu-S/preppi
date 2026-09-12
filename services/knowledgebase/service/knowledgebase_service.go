package service

import (
	"context"
	"errors"
	"fmt"

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
	articles, err := s.repo.SearchArticles(ctx, query, subject, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("kb_service_search: %w", err)
	}
	return articles, nil
}

func (s *KnowledgeBaseService) GetArticle(ctx context.Context, id uint) (*repository.Article, error) {
	a, err := s.repo.GetArticle(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("kb_service_get_article: %w", err)
	}
	return a, nil
}

func (s *KnowledgeBaseService) GetRelatedTopics(ctx context.Context, topic, subject string) ([]repository.Topic, error) {
	topics, err := s.repo.GetRelatedTopics(ctx, topic, subject)
	if err != nil {
		return nil, fmt.Errorf("kb_service_get_related_topics: %w", err)
	}
	return topics, nil
}

func (s *KnowledgeBaseService) SuggestKeywords(ctx context.Context, query string) ([]string, error) {
	keywords, err := s.repo.SuggestKeywords(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("kb_service_suggest_keywords: %w", err)
	}
	return keywords, nil
}
