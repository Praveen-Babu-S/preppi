package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	pb "preppi.com/proto/question/v1"
	"preppi.com/services/question/repository"
	"preppi.com/services/question/service"
)

type QuestionHandler struct {
	pb.UnimplementedQuestionServiceServer
	svc *service.QuestionService
}

func New(svc *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{svc: svc}
}

func (h *QuestionHandler) CreateQuestion(ctx context.Context, req *pb.CreateQuestionRequest) (*pb.CreateQuestionResponse, error) {
	studentID, err := parseID(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}

	urgency := urgencyFromPB(req.GetUrgency())
	id, qStatus, err := h.svc.Create(ctx, studentID, req.GetSubject(), req.GetTopic(), req.GetDescription(), req.GetImageUrls(), urgency)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, "subject and description are required")
		}
		return nil, status.Error(codes.Internal, "failed to create question")
	}

	now := timestamppb.Now()
	return &pb.CreateQuestionResponse{
		QuestionId: idToStr(id),
		Status:     statusToPB(qStatus),
		CreatedAt:  now,
	}, nil
}

func (h *QuestionHandler) GetQuestionById(ctx context.Context, req *pb.GetQuestionByIdRequest) (*pb.GetQuestionByIdResponse, error) {
	id, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}
	q, err := h.svc.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "question not found")
		}
		return nil, status.Error(codes.Internal, "failed to get question")
	}
	return &pb.GetQuestionByIdResponse{Question: toPB(q)}, nil
}

func (h *QuestionHandler) ListQuestionsByStudent(ctx context.Context, req *pb.ListQuestionsByStudentRequest) (*pb.ListQuestionsByStudentResponse, error) {
	studentID, err := parseID(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	questions, err := h.svc.ListByStudent(ctx, studentID, limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list questions")
	}
	resp := &pb.ListQuestionsByStudentResponse{}
	for i := range questions {
		resp.Questions = append(resp.Questions, toPB(&questions[i]))
	}
	return resp, nil
}

func (h *QuestionHandler) UpdateQuestionStatus(ctx context.Context, req *pb.UpdateQuestionStatusRequest) (*pb.UpdateQuestionStatusResponse, error) {
	id, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}
	statusStr := statusToString(req.GetStatus())
	if err := h.svc.UpdateStatus(ctx, id, statusStr); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	q, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get question")
	}
	return &pb.UpdateQuestionStatusResponse{Question: toPB(q)}, nil
}

func (h *QuestionHandler) SearchQuestions(ctx context.Context, req *pb.SearchQuestionsRequest) (*pb.SearchQuestionsResponse, error) {
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	questions, err := h.svc.Search(ctx, req.GetQuery(), req.GetSubject(), limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "search failed")
	}
	resp := &pb.SearchQuestionsResponse{}
	for i := range questions {
		resp.Questions = append(resp.Questions, toPB(&questions[i]))
	}
	return resp, nil
}

func (h *QuestionHandler) UpdateQuestion(ctx context.Context, req *pb.UpdateQuestionRequest) (*pb.UpdateQuestionResponse, error) {
	id, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}

	fields := map[string]any{}
	if req.GetSubject() != "" {
		fields["subject"] = req.GetSubject()
	}
	if req.GetTopic() != "" {
		fields["topic"] = req.GetTopic()
	}
	if req.GetDescription() != "" {
		fields["description"] = req.GetDescription()
	}
	if req.GetUrgency() != pb.Urgency_URGENCY_UNSPECIFIED {
		fields["urgency"] = urgencyFromPB(req.GetUrgency())
	}

	if err := h.svc.Update(ctx, id, fields); err != nil {
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, "no fields to update")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	q, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "question not found")
	}
	return &pb.UpdateQuestionResponse{Question: toPB(q)}, nil
}

func idToStr(id uint) string {
	return uintToStr(id)
}

func toPB(q *repository.Question) *pb.Question {
	if q == nil {
		return &pb.Question{}
	}
	return &pb.Question{
		Id:          uintToStr(q.ID),
		StudentId:   uintToStr(q.StudentID),
		AssigneeId:  optionalUint(q.AssigneeID),
		Subject:     q.Subject,
		Topic:       q.Topic,
		Description: q.Description,
		ImageUrls:   service.SplitImageURLs(q.ImageURLs),
		Urgency:     urgencyToPB(q.Urgency),
		Status:      statusToPB(q.Status),
		CreatedAt:   timestamppb.New(unixTime(q.CreatedAt)),
		UpdatedAt:   timestamppb.New(unixTime(q.UpdatedAt)),
	}
}

func statusToString(s pb.QuestionStatus) string {
	switch s {
	case pb.QuestionStatus_QUESTION_STATUS_OPEN:
		return "open"
	case pb.QuestionStatus_QUESTION_STATUS_ASSIGNED:
		return "assigned"
	case pb.QuestionStatus_QUESTION_STATUS_IN_PROGRESS:
		return "in_progress"
	case pb.QuestionStatus_QUESTION_STATUS_ANSWERED:
		return "answered"
	case pb.QuestionStatus_QUESTION_STATUS_ESCALATED:
		return "escalated"
	default:
		return "open"
	}
}
