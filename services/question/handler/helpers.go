package handler

import (
	"strconv"
	"time"

	pb "preppi.com/proto/question/v1"
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

func optionalUint(id uint) string {
	if id == 0 {
		return ""
	}
	return uintToStr(id)
}

func statusToPB(s string) pb.QuestionStatus {
	switch s {
	case "open":
		return pb.QuestionStatus_QUESTION_STATUS_OPEN
	case "assigned":
		return pb.QuestionStatus_QUESTION_STATUS_ASSIGNED
	case "in_progress":
		return pb.QuestionStatus_QUESTION_STATUS_IN_PROGRESS
	case "answered":
		return pb.QuestionStatus_QUESTION_STATUS_ANSWERED
	case "escalated":
		return pb.QuestionStatus_QUESTION_STATUS_ESCALATED
	default:
		return pb.QuestionStatus_QUESTION_STATUS_UNSPECIFIED
	}
}

func urgencyToPB(u string) pb.Urgency {
	switch u {
	case "low":
		return pb.Urgency_URGENCY_LOW
	case "urgent":
		return pb.Urgency_URGENCY_URGENT
	case "normal":
		return pb.Urgency_URGENCY_NORMAL
	default:
		return pb.Urgency_URGENCY_UNSPECIFIED
	}
}

func urgencyFromPB(u pb.Urgency) string {
	switch u {
	case pb.Urgency_URGENCY_LOW:
		return "low"
	case pb.Urgency_URGENCY_URGENT:
		return "urgent"
	default:
		return "normal"
	}
}
