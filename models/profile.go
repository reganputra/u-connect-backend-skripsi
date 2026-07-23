package models

import "gorm.io/gorm"

type UserProfile struct {
	gorm.Model
	UserID uint `gorm:"uniqueIndex;not null"` // FK → users.id
	User   User `gorm:"foreignKey:UserID"`

	// Basic
	ProfilePicture string  `gorm:"default:null"`
	Bio            *string `gorm:"default:null"`
	Location       *string `gorm:"default:null"`
	LinkedinURL    *string `gorm:"default:null"`
	GithubURL      *string `gorm:"default:null"`

	// Professional (alumni)
	JobStatus       *string `gorm:"default:null"` // employed|entrepreneur|continuing_study|unemployed|freelance|student
	Position        *string `gorm:"default:null"`
	CompanyName     *string `gorm:"default:null"`
	CompanyLocation *string `gorm:"default:null"`
	CompanySize     *int    `gorm:"default:null"`
	IndustryName    *string `gorm:"default:null"`
	IndustryType    *string `gorm:"default:null"` // e.g. B2B, B2C, SaaS, etc.
	YearFounding    *int    `gorm:"default:null"`
	Salary          *int    `gorm:"default:null"` // optional

	// Academic (alumni continuing study)
	EducationalLevel       *string `gorm:"default:null"`
	AdvancedStudyProgram   *string `gorm:"default:null"`
	InstitutionName        *string `gorm:"default:null"`
	ExpectedGraduationYear *int    `gorm:"default:null"`

	// Skills & Interests
	Skills          *string `gorm:"default:null"` // comma-separated
	Interests       *string `gorm:"default:null"` // comma-separated
	CareerInterests *string `gorm:"default:null"` // comma-separated (Minat Karir)

	// Mentorship (alumni only)
	MentorQuota       *int    `gorm:"default:null"`
	MentorDescription *string `gorm:"default:null"`

	// Additional (alumni unemployed/freelance)
	StatusDescription *string `gorm:"default:null"`

	// Experiences (1:many)
	Experiences []UserExperience `gorm:"foreignKey:UserProfileID"`
}
