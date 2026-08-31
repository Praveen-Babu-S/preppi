package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "preppi.com/proto/user/v1"
)

type UserHandler struct {
	client pb.UserServiceClient
}

func NewUserHandler(conn *grpc.ClientConn) *UserHandler {
	return &UserHandler{client: pb.NewUserServiceClient(conn)}
}

type createProfileReq struct {
	UserId    string   `json:"user_id" binding:"required"`
	Name      string   `json:"name" binding:"required"`
	AvatarURL string   `json:"avatar_url"`
	Phone     string   `json:"phone"`
	School    string   `json:"school"`
	College   string   `json:"college"`
	Bio       string   `json:"bio"`
	Role      string   `json:"role" binding:"required,oneof=student mentor"`
	Expertise []string `json:"expertise"`
	SubTopics []string `json:"sub_topics"`
}

func (h *UserHandler) CreateProfile(c *gin.Context) {
	var req createProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.CreateProfile(c, &pb.CreateProfileRequest{
		UserId:            req.UserId,
		Name:              req.Name,
		AvatarUrl:         req.AvatarURL,
		Phone:             req.Phone,
		School:            req.School,
		College:           req.College,
		Bio:               req.Bio,
		Role:              parseRole(req.Role),
		ExpertiseSubjects: req.Expertise,
		SubTopics:         req.SubTopics,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"profile": resp.GetProfile()})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		userID = c.GetString("user_id")
	}

	resp, err := h.client.GetProfile(c, &pb.GetProfileRequest{UserId: userID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": resp.GetProfile()})
}

func (h *UserHandler) GetMentorsBySubject(c *gin.Context) {
	subject := c.Query("subject")
	if subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject is required"})
		return
	}

	resp, err := h.client.GetMentorsBySubject(c, &pb.GetMentorsBySubjectRequest{Subject: subject})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get mentors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mentors": resp.GetMentors()})
}

func parseRole(role string) pb.Role {
	switch role {
	case "student":
		return pb.Role_ROLE_STUDENT
	case "mentor":
		return pb.Role_ROLE_MENTOR
	default:
		return pb.Role_ROLE_UNSPECIFIED
	}
}
