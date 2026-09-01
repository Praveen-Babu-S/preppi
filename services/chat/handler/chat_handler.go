package handler

import (
	"context"
	"errors"
	"io"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "preppi.com/proto/chat/v1"
	"preppi.com/services/chat/service"
)

type ChatHandler struct {
	pb.UnimplementedChatServiceServer
	svc *service.ChatService
}

func New(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	questionID, err := parseID(req.GetQuestionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid question_id")
	}
	studentID, err := parseID(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid student_id")
	}
	mentorID, err := parseID(req.GetMentorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mentor_id")
	}
	room, err := h.svc.CreateRoom(ctx, questionID, studentID, mentorID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create room")
	}
	return &pb.CreateRoomResponse{
		RoomId: uintToStr(room.ID),
		Status: room.Status,
	}, nil
}

func (h *ChatHandler) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.GetRoomResponse, error) {
	id, err := parseID(req.GetRoomId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid room_id")
	}
	room, err := h.svc.GetRoom(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "room not found")
	}
	return roomToPB(room), nil
}

func (h *ChatHandler) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	roomID, err := parseID(req.GetRoomId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid room_id")
	}
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 50
	}
	messages, err := h.svc.GetHistory(ctx, roomID, limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get history")
	}
	resp := &pb.GetHistoryResponse{}
	for i := range messages {
		resp.Messages = append(resp.Messages, msgToStreamPB(&messages[i]))
	}
	return resp, nil
}

func (h *ChatHandler) SendMessage(stream grpc.BidiStreamingServer[pb.SendMessageRequest, pb.SendMessageResponse]) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		roomID, err := parseID(req.GetRoomId())
		if err != nil {
			return status.Error(codes.InvalidArgument, "invalid room_id")
		}
		senderID, err := parseID(req.GetSenderId())
		if err != nil {
			return status.Error(codes.InvalidArgument, "invalid sender_id")
		}
		msgType := msgTypeString(req.GetType())
		msg, err := h.svc.SendMessage(stream.Context(), roomID, senderID, req.GetContent(), msgType, req.GetImageUrl())
		if err != nil {
			if errors.Is(err, service.ErrValidation) {
				return status.Error(codes.InvalidArgument, "content or image_url required")
			}
			return status.Error(codes.Internal, "failed to send message")
		}
		resp := &pb.SendMessageResponse{
			MessageId: strconv.FormatUint(uint64(msg.ID), 10),
			SentAt:    timestamppb.New(unixTime(msg.CreatedAt)),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (h *ChatHandler) StreamMessages(req *pb.StreamMessagesRequest, stream grpc.ServerStreamingServer[pb.StreamMessagesResponse]) error {
	roomID, err := parseID(req.GetRoomId())
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid room_id")
	}

	ch := h.svc.Subscribe(roomID)
	defer h.svc.Unsubscribe(roomID, ch)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(msgToStreamPB(msg)); err != nil {
				return err
			}
		}
	}
}

func (h *ChatHandler) MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error) {
	roomID, err := parseID(req.GetRoomId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid room_id")
	}
	userID, err := parseID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	count, err := h.svc.MarkRead(ctx, roomID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark read")
	}
	return &pb.MarkReadResponse{UnreadCount: int32(count)}, nil
}

func msgTypeString(t pb.MessageType) string {
	switch t {
	case pb.MessageType_MESSAGE_TYPE_IMAGE:
		return "image"
	case pb.MessageType_MESSAGE_TYPE_SYSTEM:
		return "system"
	default:
		return "text"
	}
}
