package service

import "github.com/reganputra/skripsi-backend/models"

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
