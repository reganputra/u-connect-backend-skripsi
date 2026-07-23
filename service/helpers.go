package service

import (
	"github.com/reganputra/skripsi-backend/models"
)

// ApplyUserPicture promotes profile.ProfilePicture to user.PictureURL and nils
// out the Profile field to prevent leaking raw profile data in API responses.
//
// All service files in this package use this single canonical implementation
// instead of their own private copies (feed, group, event, message services).
func ApplyUserPicture(user *models.User) {
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

// TruncateText shortens s to at most max runes, appending an ellipsis when
// truncated. It is the single canonical implementation previously duplicated
// as truncate (notification_service) and truncateNotifyText (admin_service).
func TruncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}
