package handler

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "preppi.com/proto/knowledgebase/v1"
	"preppi.com/services/knowledgebase/repository"
	"preppi.com/services/knowledgebase/service"
)

type KnowledgeBaseHandler struct {
	pb.UnimplementedKnowledgeBaseServiceServer
	svc *service.KnowledgeBaseService
}

func New(svc *service.KnowledgeBaseService) *KnowledgeBaseHandler {
	return &KnowledgeBaseHandler{svc: svc}
}

func (h *KnowledgeBaseHandler) SearchKB(ctx context.Context, req *pb.SearchKBRequest) (*pb.SearchKBResponse, error) {
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	articles, err := h.svc.Search(ctx, req.GetQuery(), req.GetSubject(), limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "search failed")
	}
	resp := &pb.SearchKBResponse{}
	for i := range articles {
		resp.Articles = append(resp.Articles, articleToPB(&articles[i]))
	}
	return resp, nil
}

func (h *KnowledgeBaseHandler) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	id, err := parseID(req.GetArticleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid article_id")
	}
	a, err := h.svc.GetArticle(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "article not found")
	}
	return &pb.GetArticleResponse{Article: articleToPB(a), FullText: a.FullText}, nil
}

func (h *KnowledgeBaseHandler) GetRelatedTopics(ctx context.Context, req *pb.GetRelatedTopicsRequest) (*pb.GetRelatedTopicsResponse, error) {
	topics, err := h.svc.GetRelatedTopics(ctx, req.GetTopic(), req.GetSubject())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get related topics")
	}
	resp := &pb.GetRelatedTopicsResponse{}
	for i := range topics {
		resp.Topics = append(resp.Topics, topicToPB(&topics[i]))
	}
	return resp, nil
}

func (h *KnowledgeBaseHandler) SuggestKeywords(ctx context.Context, req *pb.SuggestKeywordsRequest) (*pb.SuggestKeywordsResponse, error) {
	keywords, err := h.svc.SuggestKeywords(ctx, req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to suggest keywords")
	}
	return &pb.SuggestKeywordsResponse{Keywords: keywords}, nil
}

func articleToPB(a *repository.Article) *pb.Article {
	if a == nil {
		return &pb.Article{}
	}
	return &pb.Article{
		Id:        uintToStr(a.ID),
		Subject:   a.Subject,
		Topic:     a.Topic,
		Title:     a.Title,
		Summary:   a.Summary,
		Upvotes:   int32(a.Upvotes),
		CreatedAt: timestamppb.New(time.Unix(a.CreatedAt, 0)),
	}
}

func topicToPB(t *repository.Topic) *pb.Topic {
	if t == nil {
		return &pb.Topic{}
	}
	return &pb.Topic{
		Id:           uintToStr(t.ID),
		Name:         t.Name,
		Subject:      t.Subject,
		ArticleCount: int32(t.ArticleCount),
	}
}

func parseID(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func uintToStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
