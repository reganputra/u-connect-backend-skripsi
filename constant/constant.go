package constant

// Status constants for mentor requests, mentoring sessions, events,
// job applications, job postings and reports.
const (
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusEnded     = "ended"
	StatusCancelled = "cancelled"

	StatusScheduled = "scheduled"
	StatusCompleted = "completed"

	StatusUpcoming = "upcoming"
	StatusOngoing  = "ongoing"

	StatusWithdrawn = "withdrawn"

	StatusOpen   = "open"
	StatusClosed = "closed"
	StatusFilled = "filled"

	StatusResolved = "resolved"
)

// Role constants for users and group members.
const (
	RoleStudent = "student"
	RoleAlumni  = "alumni"
	RolePartner = "partner"
	RoleAdmin   = "admin"
	RoleOwner   = "owner"
)

// DirectoryRoles are the roles eligible for the directory listing.
var DirectoryRoles = []string{RoleStudent, RoleAlumni, RolePartner}

// IsValidRole reports whether role is one of the known user roles.
func IsValidRole(role string) bool {
	switch role {
	case RoleStudent, RoleAlumni, RolePartner, RoleAdmin:
		return true
	default:
		return false
	}
}

// IsValidDirectoryRole reports whether role is eligible for the directory.
func IsValidDirectoryRole(role string) bool {
	for _, r := range DirectoryRoles {
		if r == role {
			return true
		}
	}
	return false
}

// Reaction types shared by feed and group reactions.
const (
	ReactionLike  = "like"
	ReactionLove  = "love"
	ReactionHaha  = "haha"
	ReactionWow   = "wow"
	ReactionSad   = "sad"
	ReactionAngry = "angry"
)

// ValidReactionTypes is the set of accepted reaction types.
var ValidReactionTypes = map[string]bool{
	ReactionLike:  true,
	ReactionLove:  true,
	ReactionHaha:  true,
	ReactionWow:   true,
	ReactionSad:   true,
	ReactionAngry: true,
}

// IsValidReaction reports whether t is an accepted reaction type.
func IsValidReaction(t string) bool {
	return ValidReactionTypes[t]
}

// Job types.
const (
	JobTypeFullTime   = "full-time"
	JobTypePartTime   = "part-time"
	JobTypeInternship = "internship"
	JobTypeContract   = "contract"
	JobTypeFreelance  = "freelance"
)

// JobStatuses is the set of accepted job statuses.
var JobStatuses = map[string]bool{
	StatusOpen:   true,
	StatusClosed: true,
	StatusFilled: true,
}

// ApplicationStatuses is the set of accepted job-application statuses.
const (
	ApplicationStatusReviewed = "reviewed"
	ApplicationStatusAccepted = "accepted"
)

var ApplicationStatuses = map[string]bool{
	StatusPending:             true,
	ApplicationStatusReviewed: true,
	ApplicationStatusAccepted: true,
	StatusRejected:            true,
}

// SessionStatuses is the set of accepted mentoring-session statuses.
var SessionStatuses = map[string]bool{
	StatusScheduled: true,
	StatusCompleted: true,
	StatusCancelled: true,
}

// EventStatuses is the set of accepted event statuses.
var EventStatuses = map[string]bool{
	StatusUpcoming:  true,
	StatusOngoing:   true,
	StatusCompleted: true,
	StatusCancelled: true,
}

// ProfileJobStatuses are the allowed job_status values on a profile.
var ProfileJobStatuses = map[string]bool{
	"employed":         true,
	"entrepreneur":     true,
	"continuing_study": true,
	"unemployed":       true,
	"freelance":        true,
	"student":          true,
}

// MentorQuotas are the allowed mentor quota values.
var MentorQuotas = map[int]bool{
	1: true,
	2: true,
	3: true,
	5: true,
}

// Report target and report type constants.
const (
	ReportTargetPost         = "post"
	ReportTargetGroup        = "group"
	ReportTargetEvent        = "event"
	ReportTargetJob          = "job"
	ReportTargetComment      = "comment"
	ReportTargetGroupArticle = "group_article"
)

// ReportType constants.
const (
	ReportTypeHarassment     = "harassment"
	ReportTypeViolence       = "violence"
	ReportTypeHateSpeech     = "hate_speech"
	ReportTypeSpam           = "spam"
	ReportTypeInappropriate  = "inappropriate"
	ReportTypeMisinformation = "misinformation"
	ReportTypeCopyright      = "copyright"
	ReportTypeOther          = "other"
)

// Pagination bounds.
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// GeneralCategoryName is the default category name used when none is provided.
const GeneralCategoryName = "General"

// Misc config-style values used across services.
const (
	MaxResetAttempts    = 3
	MinPasswordLen     = 8
	MaxMentorsPerStudent = 2
)

// Auth/JWT durations.
const (
	JWTAccessExpiry = 24 * 60 * 60 // seconds (24h)
)

// Recommendation/cache/notification tuning values.
const (
	RecommendationCacheTTL    = 5 * 60 // seconds (5m)
	RecommendationMinScore    = 0.1
	NotificationThrottleShort = 30 * 60 // seconds (30m)
	NotificationThrottleLong  = 60 * 60 // seconds (60m)
	TopNDefault               = 10
)
