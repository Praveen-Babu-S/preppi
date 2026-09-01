package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "preppi.com/proto/doubt/v1"
	"preppi.com/services/doubt/repository"
	"preppi.com/services/doubt/service"
)

type NotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	svc *service.DoubtService
}

func NewNotificationHandler(svc *service.DoubtService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	userID, err := parseID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	var channels []string
	for _, c := range req.GetChannels() {
		channels = append(channels, channelToString(c))
	}
	id, err := h.svc.SendNotification(ctx, userID, typeToString(req.GetType()), req.GetTitle(), req.GetBody(), channels)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, "title and body are required")
		}
		return nil, status.Error(codes.Internal, "failed to send notification")
	}
	return &pb.SendNotificationResponse{NotificationId: uintToStr(id)}, nil
}

func (h *NotificationHandler) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	userID, err := parseID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	limit := int(req.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	notifications, err := h.svc.ListNotifications(ctx, userID, limit, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list notifications")
	}
	resp := &pb.ListNotificationsResponse{}
	for i := range notifications {
		resp.Notifications = append(resp.Notifications, notificationToPB(&notifications[i]))
	}
	return resp, nil
}

func (h *NotificationHandler) NotificationMarkRead(ctx context.Context, req *pb.NotificationMarkReadRequest) (*pb.NotificationMarkReadResponse, error) {
	userID, err := parseID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	notificationID, err := parseID(req.GetNotificationId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id")
	}
	if err := h.svc.MarkNotificationRead(ctx, userID, notificationID); err != nil {
		return nil, status.Error(codes.Internal, "failed to mark read")
	}
	return &pb.NotificationMarkReadResponse{Success: true}, nil
}

func (h *NotificationHandler) GetPreferences(ctx context.Context, req *pb.GetPreferencesRequest) (*pb.GetPreferencesResponse, error) {
	userID, err := parseID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	p, err := h.svc.GetPreferences(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get preferences")
	}
	return &pb.GetPreferencesResponse{Preferences: prefsToPB(p)}, nil
}

func (h *NotificationHandler) UpdatePreferences(ctx context.Context, req *pb.UpdatePreferencesRequest) (*pb.UpdatePreferencesResponse, error) {
	userID, err := parseID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	prefs := req.GetPreferences()
	p, err := h.svc.UpdatePreferences(ctx, userID, prefs.GetInAppEnabled(), prefs.GetPushEnabled(), prefs.GetEmailEnabled(), prefs.GetSmsEnabled(), prefs.GetDigestMode())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update preferences")
	}
	return &pb.UpdatePreferencesResponse{Preferences: prefsToPB(p)}, nil
}

func notificationToPB(n *repository.NotificationRecord) *pb.Notification {
	if n == nil {
		return &pb.Notification{}
	}
	return &pb.Notification{
		Id:        uintToStr(n.ID),
		UserId:    uintToStr(n.UserID),
		Type:      typeToPB(n.Type),
		Title:     n.Title,
		Body:      n.Body,
		Read:      n.Read,
		CreatedAt: timestamppb.New(unixTime(n.CreatedAt)),
	}
}

func prefsToPB(p *repository.NotificationPreference) *pb.NotificationPreferences {
	if p == nil {
		return &pb.NotificationPreferences{}
	}
	return &pb.NotificationPreferences{
		InAppEnabled: p.InAppEnabled,
		PushEnabled:  p.PushEnabled,
		EmailEnabled: p.EmailEnabled,
		SmsEnabled:   p.SMSEnabled,
		DigestMode:   p.DigestMode,
	}
}

func typeToString(t pb.NotificationType) string {
	switch t {
	case pb.NotificationType_NOTIFICATION_TYPE_QUESTION_ASSIGNED:
		return "question_assigned"
	case pb.NotificationType_NOTIFICATION_TYPE_SOLUTION_POSTED:
		return "solution_posted"
	case pb.NotificationType_NOTIFICATION_TYPE_QUESTION_ESCALATED:
		return "question_escalated"
	case pb.NotificationType_NOTIFICATION_TYPE_COMMENT_POSTED:
		return "comment_posted"
	case pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_UP:
		return "follow_up"
	case pb.NotificationType_NOTIFICATION_TYPE_SYSTEM:
		return "system"
	default:
		return "system"
	}
}

func typeToPB(s string) pb.NotificationType {
	switch s {
	case "question_assigned":
		return pb.NotificationType_NOTIFICATION_TYPE_QUESTION_ASSIGNED
	case "solution_posted":
		return pb.NotificationType_NOTIFICATION_TYPE_SOLUTION_POSTED
	case "question_escalated":
		return pb.NotificationType_NOTIFICATION_TYPE_QUESTION_ESCALATED
	case "comment_posted":
		return pb.NotificationType_NOTIFICATION_TYPE_COMMENT_POSTED
	case "follow_up":
		return pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_UP
	default:
		return pb.NotificationType_NOTIFICATION_TYPE_SYSTEM
	}
}

func channelToString(c pb.Channel) string {
	switch c {
	case pb.Channel_CHANNEL_IN_APP:
		return "in_app"
	case pb.Channel_CHANNEL_PUSH:
		return "push"
	case pb.Channel_CHANNEL_EMAIL:
		return "email"
	case pb.Channel_CHANNEL_SMS:
		return "sms"
	default:
		return "in_app"
	}
}
