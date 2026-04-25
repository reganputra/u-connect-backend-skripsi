package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

type FollowService interface {
	Follow(followerID, followingID uint, followerRole string) error
	Unfollow(followerID, followingID uint) error
	GetFollowers(userID uint) ([]models.User, error)
	GetFollowing(userID uint) ([]models.User, error)
}

type followService struct {
	repo     repository.FollowRepository
	userRepo repository.UserRepository
	notifSvc NotificationService
}

func NewFollowService(repo repository.FollowRepository, userRepo repository.UserRepository, notifSvc NotificationService) FollowService {
	return &followService{repo: repo, userRepo: userRepo, notifSvc: notifSvc}
}

func (s *followService) Follow(followerID, followingID uint, followerRole string) error {
	// Role guard
	if followerRole != "student" && followerRole != "alumni" {
		return errors.New("hanya student dan alumni yang dapat mengikuti pengguna")
	}
	// Self-follow guard
	if followerID == followingID {
		return errors.New("tidak dapat mengikuti diri sendiri")
	}
	// Duplicate guard
	already, err := s.repo.IsFollowing(followerID, followingID)
	if err != nil {
		return err
	}
	if already {
		return errors.New("anda sudah mengikuti pengguna ini")
	}
	if err := s.repo.Follow(followerID, followingID); err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "idx_follow_pair") || strings.Contains(lowerErr, "duplicate key") {
			return errors.New("anda sudah mengikuti pengguna ini")
		}
		return err
	}
	// Notify the followed user
	if follower, err := s.userRepo.FindUserByID(followerID); err == nil {
		_ = s.notifSvc.Notify(
			followingID,
			"new_follower",
			"Pengikut baru",
			fmt.Sprintf("%s mulai mengikutimu", follower.Name),
			"follow",
			followerID,
		)
	}
	return nil
}

func (s *followService) Unfollow(followerID, followingID uint) error {
	if followerID == followingID {
		return errors.New("tidak dapat berhenti mengikuti diri sendiri")
	}
	following, err := s.repo.IsFollowing(followerID, followingID)
	if err != nil {
		return err
	}
	if !following {
		return errors.New("anda tidak mengikuti pengguna ini")
	}
	return s.repo.Unfollow(followerID, followingID)
}

func (s *followService) GetFollowers(userID uint) ([]models.User, error) {
	return s.repo.GetFollowers(userID)
}

func (s *followService) GetFollowing(userID uint) ([]models.User, error) {
	return s.repo.GetFollowing(userID)
}
