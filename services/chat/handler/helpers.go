package handler

import (
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "preppi.com/proto/chat/v1"
	"preppi.com/services/chat/repository"
)

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

func unixTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

func timestamppbNew(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func roomToPB(r *repository.ChatRoom) *pb.GetRoomResponse {
	if r == nil {
		return &pb.GetRoomResponse{}
	}
	return &pb.GetRoomResponse{
		RoomId:     uintToStr(r.ID),
		QuestionId: uintToStr(r.QuestionID),
		StudentId:  uintToStr(r.StudentID),
		MentorId:   uintToStr(r.MentorID),
		Status:     r.Status,
		CreatedAt:  timestamppbNew(unixTime(r.CreatedAt)),
	}
}

func msgToStreamPB(m *repository.Message) *pb.StreamMessagesResponse {
	if m == nil {
		return &pb.StreamMessagesResponse{}
	}
	return &pb.StreamMessagesResponse{
		Id:        uintToStr(m.ID),
		RoomId:    uintToStr(m.RoomID),
		SenderId:  uintToStr(m.SenderID),
		Content:   m.Content,
		Type:      msgTypeToPB(m.Type),
		ImageUrl:  m.ImageURL,
		Read:      m.Read,
		CreatedAt: timestamppbNew(unixTime(m.CreatedAt)),
	}
}

func msgTypeToPB(s string) pb.MessageType {
	switch s {
	case "image":
		return pb.MessageType_MESSAGE_TYPE_IMAGE
	case "system":
		return pb.MessageType_MESSAGE_TYPE_SYSTEM
	default:
		return pb.MessageType_MESSAGE_TYPE_TEXT
	}
}
