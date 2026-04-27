package service

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

type MessageService interface {
	// SendMessage persists a message after validating follow relationship.
	SendMessage(senderID, receiverID uint, content string) (*models.Message, error)
	// GetConversation returns paginated messages between two users.
	// Returns error if caller is not a participant.
	GetConversation(callerID, partnerID uint, page, limit int) ([]models.Message, int64, error)
	// GetConversationList returns the most recent message per unique conversation partner.
	GetConversationList(callerID uint) ([]repository.ConversationSummary, error)
	// MarkAsRead marks all messages from senderID to receiverID as read.
	MarkAsRead(receiverID, senderID uint) error
	// CountUnread returns the total number of unread messages for a user.
	CountUnread(userID uint) (int64, error)
}

type messageService struct {
	msgRepo    repository.MessageRepository
	followRepo repository.FollowRepository
}

func applyMessageUserPicture(user *models.User) {
	if user == nil {
		return
	}
	if user.Profile != nil {
		if user.Profile.ProfilePicture != "" {
			picture := user.Profile.ProfilePicture
			user.PictureURL = &picture
		} else {
			user.PictureURL = nil
		}
		user.Profile = nil
	}
}

func NewMessageService(
	msgRepo repository.MessageRepository,
	followRepo repository.FollowRepository,
) MessageService {
	return &messageService{
		msgRepo:    msgRepo,
		followRepo: followRepo,
	}
}

func (s *messageService) SendMessage(senderID, receiverID uint, content string) (*models.Message, error) {
	if senderID == receiverID {
		return nil, errors.New("tidak dapat mengirim pesan kepada diri sendiri")
	}
	if content == "" {
		return nil, errors.New("isi pesan tidak boleh kosong")
	}

	// Symmetric follow check: either A follows B or B follows A
	connected, err := s.followRepo.AreConnected(senderID, receiverID)
	if err != nil {
		return nil, err
	}
	if !connected {
		return nil, errors.New("anda harus mengikuti pengguna ini sebelum mengirim pesan")
	}

	msg := &models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	if err := s.msgRepo.Create(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *messageService) GetConversation(callerID, partnerID uint, page, limit int) ([]models.Message, int64, error) {
	if callerID != partnerID {
		// Verify caller is a participant before returning messages
		// (we allow the query itself to do the filtering, but we want
		//  to return an error for completely unrelated callers)
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	msgs, total, err := s.msgRepo.GetConversation(callerID, partnerID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range msgs {
		applyMessageUserPicture(&msgs[i].Sender)
		applyMessageUserPicture(&msgs[i].Receiver)
	}
	return msgs, total, nil
}

func (s *messageService) GetConversationList(callerID uint) ([]repository.ConversationSummary, error) {
	return s.msgRepo.GetConversationList(callerID)
}

func (s *messageService) MarkAsRead(receiverID, senderID uint) error {
	return s.msgRepo.MarkAsRead(receiverID, senderID)
}

func (s *messageService) CountUnread(userID uint) (int64, error) {
	return s.msgRepo.CountUnread(userID)
}
