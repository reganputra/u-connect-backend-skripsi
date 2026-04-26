package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

// Deliverer is satisfied by *ws.Hub — avoids import cycle between service and ws packages.
type Deliverer interface {
	SendToUser(userID uint, payload []byte) bool
}

type NotificationService interface {
	Notify(userID uint, notifType, title, body, refType string, refID uint) error
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

// Notify persists and delivers a notification.
// Redirect URL resolution and WS delivery are async (fire-and-forget).
func (s *notificationService) Notify(userID uint, notifType, title, body, refType string, refID uint) error {
	n := &models.Notification{
		UserID:           userID,
		NotificationType: notifType,
		Title:            title,
		Body:             body,
		ReferenceType:    refType,
		ReferenceID:      refID,
	}
	if err := s.repo.Create(n); err != nil {
		return err
	}
	// 1.1 + 1.2: resolve redirect URL and push to WS off the request path.
	go func() {
		n.RedirectURL = s.buildRedirectURL(userID, refType, refID)
		if err := s.repo.UpdateRedirectURL(n.ID, n.RedirectURL); err != nil {
			log.Printf("[NOTIF] redirect update failed id=%d: %v", n.ID, err)
		}
		s.deliver(n)
	}()
	return nil
}

func (s *notificationService) NotifyThrottled(userID uint, notifType, title, body, refType string, refID uint, throttle time.Duration) error {
	exists, err := s.repo.ExistsRecent(userID, notifType, refType, refID, throttle)
	if err != nil {
		return err
	}
	if exists {
		return nil
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

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return fmt.Sprintf("%s…", string(runes[:n]))
}

// buildRedirectURL computes the frontend route for a notification reference.
// 1.4: group_comment uses a single JOIN instead of two sequential queries.
func (s *notificationService) buildRedirectURL(userID uint, refType string, refID uint) string {
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
		// 1.4: single JOIN replaces two sequential queries.
		var row struct {
			ArticleID uint
			GroupID   uint
		}
		s.db.Raw(`
			SELECT gc.article_id, ga.group_id
			FROM group_comments gc
			JOIN group_articles ga ON ga.id = gc.article_id
			WHERE gc.id = ? AND gc.deleted_at IS NULL AND ga.deleted_at IS NULL
		`, refID).Scan(&row)
		if row.ArticleID != 0 {
			return fmt.Sprintf("/groups/%d/article/%d", row.GroupID, row.ArticleID)
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
