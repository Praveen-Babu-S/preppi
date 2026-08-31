package handler

import (
	"strconv"
	"time"

	pb "preppi.com/proto/user/v1"
)

func uintToStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func unixTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

func roleFromString(role string) pb.Role {
	switch role {
	case "student":
		return pb.Role_ROLE_STUDENT
	case "mentor":
		return pb.Role_ROLE_MENTOR
	case "admin":
		return pb.Role_ROLE_ADMIN
	default:
		return pb.Role_ROLE_UNSPECIFIED
	}
}
