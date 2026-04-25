package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

// Deliverer is satisfied by *ws.Hub — defined here to avoid an import cycle
// between the service and ws packages.
type Deliverer interface {
	SendToUser(userID uint, payload []byte) bool
}

type NotificationService interface {
	// Notify creates a notification immediately (no throttle).
	Notify(userID uint, notifType, title, body, refType string, refID uint) error
	// NotifyThrottled skips creating a notification if one of the same type+reference
	// already exists within the throttle window.
	NotifyThrottled(userID uint, notifType, title, body, refType string, refID uint, throttle time.Duration) error

	GetMyNotifications(userID uint, page, limit int) ([]models.Notification, int64, error)
	MarkAsRead(notificationID, userID uint) error
	MarkAllAsRead(userID uint) error
	CountUnread(userID uint) (int64, error)
}

type notificationService struct {
	repo      repository.NotificationRepository
	deliverer Deliverer
	db        *gorm.DB
}

func NewNotificationService(repo repository.NotificationRepository, deliverer Deliverer, db *gorm.DB) NotificationService {
	return &notificationService{repo: repo, deliverer: deliverer, db: db}
}

func (s *notificationService) Notify(userID uint, notifType, title, body, refType string, refID uint) error {
	n := &models.Notification{
		UserID:           userID,
		NotificationType: notifType,
		Title:            title,
		Body:             body,
		ReferenceType:    refType,
		ReferenceID:      refID,
		RedirectURL:      s.buildNotificationRedirectURL(userID, refType, refID),
	}
	if err := s.repo.Create(n); err != nil {
		return err
	}
	s.deliver(n)
	return nil
}

func (s *notificationService) NotifyThrottled(userID uint, notifType, title, body, refType string, refID uint, throttle time.Duration) error {
	exists, err := s.repo.ExistsRecent(userID, notifType, refType, refID, throttle)
	if err != nil {
		return err
	}
	if exists {
		return nil // throttled — skip silently
	}
	return s.Notify(userID, notifType, title, body, refType, refID)
}

func (s *notificationService) GetMyNotifications(userID uint, page, limit int) ([]models.Notification, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.GetByUserID(userID, page, limit)
}

func (s *notificationService) MarkAsRead(notificationID, userID uint) error {
	return s.repo.MarkAsRead(notificationID, userID)
}

func (s *notificationService) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}

func (s *notificationService) CountUnread(userID uint) (int64, error) {
	return s.repo.CountUnread(userID)
}

// deliver pushes a notification to the WebSocket hub if the user is connected.
// Failures are silently ignored — the notification is already persisted.
func (s *notificationService) deliver(n *models.Notification) {
	type outgoing struct {
		Type string               `json:"type"`
		Data *models.Notification `json:"data"`
	}
	payload, err := json.Marshal(outgoing{Type: "notification", Data: n})
	if err != nil {
		return
	}
	s.deliverer.SendToUser(n.UserID, payload)
}

// ── Helper: truncate a string to max N runes ───────────────────────────────
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return fmt.Sprintf("%s…", string(runes[:n]))
}

func (s *notificationService) buildNotificationRedirectURL(userID uint, refType string, refID uint) string {
	switch refType {
	case "post":
		return fmt.Sprintf("/feed/%d", refID)
	case "comment":
		var comment models.Comment
		if err := s.db.Select("id", "post_id").First(&comment, refID).Error; err == nil {
			return fmt.Sprintf("/feed/%d", comment.PostID)
		}
		return fmt.Sprintf("/comments/%d", refID)
	case "group":
		return fmt.Sprintf("/groups/%d", refID)
	case "group_article":
		var article models.GroupArticle
		if err := s.db.Select("id", "group_id").First(&article, refID).Error; err == nil {
			return fmt.Sprintf("/groups/%d/article/%d", article.GroupID, article.ID)
		}
		return fmt.Sprintf("/groups/articles/%d", refID)
	case "group_comment":
		var comment models.GroupComment
		if err := s.db.Select("id", "article_id").First(&comment, refID).Error; err == nil {
			var article models.GroupArticle
			if err := s.db.Select("id", "group_id").First(&article, comment.ArticleID).Error; err == nil {
				return fmt.Sprintf("/groups/%d/article/%d", article.GroupID, article.ID)
			}
		}
		return fmt.Sprintf("/groups/comments/%d", refID)
	case "event":
		return fmt.Sprintf("/events/%d", refID)
	case "job":
		return fmt.Sprintf("/jobs/%d", refID)
	case "job_application":
		return "/jobs/applications/mine"
	case "follow":
		return fmt.Sprintf("/directory/%d", refID)
	case "mentor_request":
		return "/student/requests"
	case "report":
		return "/reports/mine"
	case "message":
		var msg models.Message
		if err := s.db.Select("id", "sender_id", "receiver_id").First(&msg, refID).Error; err == nil {
			if msg.ReceiverID == userID {
				return fmt.Sprintf("/messages/%d", msg.SenderID)
			}
			if msg.SenderID == userID {
				return fmt.Sprintf("/messages/%d", msg.ReceiverID)
			}
		}
		return "/messages"
	default:
		return ""
	}
}
