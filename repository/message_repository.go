package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// ConversationSummary is returned by GetConversationList: last message per conversation partner.
type ConversationSummary struct {
	PartnerID      uint   `json:"partner_id"`
	PartnerName    string `json:"partner_name"`
	LastMessage    string `json:"last_message"`
	LastMessageAt  string `json:"last_message_at"`
	UnreadCount    int64  `json:"unread_count"`
}

type MessageRepository interface {
	Create(msg *models.Message) error
	// GetConversation returns paginated messages between two users, newest first.
	GetConversation(userA, userB uint, page, limit int) ([]models.Message, int64, error)
	// GetConversationList returns one summary row per unique conversation partner for userID.
	GetConversationList(userID uint) ([]ConversationSummary, error)
	// MarkAsRead marks all unread messages sent by senderID to receiverID as read.
	MarkAsRead(receiverID, senderID uint) error
	// CountUnread returns the total number of unread messages for receiverID.
	CountUnread(receiverID uint) (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(msg *models.Message) error {
	return r.db.Create(msg).Error
}

func (r *messageRepository) GetConversation(userA, userB uint, page, limit int) ([]models.Message, int64, error) {
	var msgs []models.Message
	var total int64

	base := r.db.Model(&models.Message{}).
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userA, userB, userB, userA,
		)

	base.Count(&total)

	offset := (page - 1) * limit
	err := base.
		Preload("Sender").
		Preload("Receiver").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&msgs).Error

	return msgs, total, err
}

func (r *messageRepository) GetConversationList(userID uint) ([]ConversationSummary, error) {
	var summaries []ConversationSummary

	// For each partner, grab the latest message and unread count.
	// We derive partner = the other side of the message.
	err := r.db.Raw(`
		SELECT
			partner_id,
			u.name AS partner_name,
			last_msg.content AS last_message,
			TO_CHAR(last_msg.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS last_message_at,
			COALESCE(unread.cnt, 0) AS unread_count
		FROM (
			-- All distinct partner IDs for this user
			SELECT DISTINCT
				CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS partner_id
			FROM messages
			WHERE (sender_id = ? OR receiver_id = ?)
			  AND deleted_at IS NULL
		) partners
		JOIN LATERAL (
			-- Latest message for each partner
			SELECT content, created_at
			FROM messages
			WHERE ((sender_id = ? AND receiver_id = partners.partner_id)
			    OR (sender_id = partners.partner_id AND receiver_id = ?))
			  AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		) last_msg ON true
		LEFT JOIN LATERAL (
			-- Unread messages from this partner to your user
			SELECT COUNT(*) AS cnt
			FROM messages
			WHERE sender_id = partners.partner_id
			  AND receiver_id = ?
			  AND is_read = false
			  AND deleted_at IS NULL
		) unread ON true
		JOIN users u ON u.id = partners.partner_id AND u.deleted_at IS NULL
		ORDER BY last_msg.created_at DESC
	`, userID, userID, userID, userID, userID, userID).
		Scan(&summaries).Error

	return summaries, err
}

func (r *messageRepository) MarkAsRead(receiverID, senderID uint) error {
	return r.db.Model(&models.Message{}).
		Where("receiver_id = ? AND sender_id = ? AND is_read = false", receiverID, senderID).
		Update("is_read", true).Error
}

func (r *messageRepository) CountUnread(receiverID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Where("receiver_id = ? AND is_read = false", receiverID).
		Count(&count).Error
	return count, err
}
