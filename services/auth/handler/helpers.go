package handler

import (
	"strconv"

	pb "preppi.com/proto/auth/v1"
)

func fmtUint(id uint64) string {
	return strconv.FormatUint(id, 10)
}

func roleToPB(role string) pb.Role {
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
