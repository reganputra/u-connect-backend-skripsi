package controllers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

func getUserIDFromToken(c *fiber.Ctx) (uint, error) {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return 0, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}
	idFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user_id in token")
	}
	return uint(idFloat), nil
}

// parseOptionalString returns a *string if the form value is non-empty, else nil
func parseOptionalString(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}

// parseOptionalInt returns a *int if the form value is a valid non-zero int, else nil
func parseOptionalInt(val string) *int {
	if val == "" {
		return nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return nil
	}
	return &n
}

// uploadFileIfPresent uploads a form file to Cloudinary and returns its URL.
// Returns empty string and nil error if no file was provided.
func uploadFileIfPresent(c *fiber.Ctx, fieldName, folder string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil || file == nil {
		return "", nil // no file — not an error
	}
	url, err := utils.UploadImage(config.Cloudinary, file, folder)
	if err != nil {
		return "", err
	}
	return url, nil
}

func CreateProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	// Handle optional profile picture upload
	pictureURL, err := uploadFileIfPresent(c, "picture", "alumni-platform/profiles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	yearEnroll := parseOptionalInt(c.FormValue("year_enroll"))
	salary := parseOptionalInt(c.FormValue("salary"))
	yearFounding := parseOptionalInt(c.FormValue("year_founding"))
	mentorQuota := parseOptionalInt(c.FormValue("mentor_quota"))
	expectedGradYear := parseOptionalInt(c.FormValue("expected_graduation_year"))

	req := service.ProfileRequest{
		ProfilePicture:         parseOptionalString(pictureURL),
		Bio:                    parseOptionalString(c.FormValue("bio")),
		Location:               parseOptionalString(c.FormValue("location")),
		JobStatus:              parseOptionalString(c.FormValue("job_status")),
		Position:               parseOptionalString(c.FormValue("position")),
		CompanyName:            parseOptionalString(c.FormValue("company_name")),
		IndustryName:           parseOptionalString(c.FormValue("industry_name")),
		IndustryType:           parseOptionalString(c.FormValue("industry_type")),
		YearFounding:           yearFounding,
		Salary:                 salary,
		EducationalLevel:       parseOptionalString(c.FormValue("educational_level")),
		AdvancedStudyProgram:   parseOptionalString(c.FormValue("advanced_study_program")),
		InstitutionName:        parseOptionalString(c.FormValue("institution_name")),
		ExpectedGraduationYear: expectedGradYear,
		Skills:                 parseOptionalString(c.FormValue("skills")),
		Interests:              parseOptionalString(c.FormValue("interests")),
		MentorQuota:            mentorQuota,
		MentorDescription:      parseOptionalString(c.FormValue("mentor_description")),
		StatusDescription:      parseOptionalString(c.FormValue("status_description")),
	}

	// year_enroll is part of User, not profile — ignore here
	_ = yearEnroll

	profile, err := service.CreateProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, profile)
}

func GetProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	profile, err := service.GetProfile(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

func UpdateProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	// Handle optional profile picture upload
	pictureURL, err := uploadFileIfPresent(c, "picture", "alumni-platform/profiles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	salary := parseOptionalInt(c.FormValue("salary"))
	yearFounding := parseOptionalInt(c.FormValue("year_founding"))
	mentorQuota := parseOptionalInt(c.FormValue("mentor_quota"))
	expectedGradYear := parseOptionalInt(c.FormValue("expected_graduation_year"))

	req := service.ProfileRequest{
		ProfilePicture:         parseOptionalString(pictureURL),
		Bio:                    parseOptionalString(c.FormValue("bio")),
		Location:               parseOptionalString(c.FormValue("location")),
		JobStatus:              parseOptionalString(c.FormValue("job_status")),
		Position:               parseOptionalString(c.FormValue("position")),
		CompanyName:            parseOptionalString(c.FormValue("company_name")),
		IndustryName:           parseOptionalString(c.FormValue("industry_name")),
		IndustryType:           parseOptionalString(c.FormValue("industry_type")),
		YearFounding:           yearFounding,
		Salary:                 salary,
		EducationalLevel:       parseOptionalString(c.FormValue("educational_level")),
		AdvancedStudyProgram:   parseOptionalString(c.FormValue("advanced_study_program")),
		InstitutionName:        parseOptionalString(c.FormValue("institution_name")),
		ExpectedGraduationYear: expectedGradYear,
		Skills:                 parseOptionalString(c.FormValue("skills")),
		Interests:              parseOptionalString(c.FormValue("interests")),
		MentorQuota:            mentorQuota,
		MentorDescription:      parseOptionalString(c.FormValue("mentor_description")),
		StatusDescription:      parseOptionalString(c.FormValue("status_description")),
	}

	profile, err := service.UpdateProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

func DeleteProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	if err := service.DeleteProfile(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "profile deleted successfully"})
}

func AddExperience(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	var req service.ExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	exp, err := service.AddExperience(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, exp)
}

func UpdateExperience(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	expID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid experience id")
	}

	var req service.ExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	exp, err := service.UpdateExperience(userID, uint(expID), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, exp)
}

func DeleteExperience(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	expID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid experience id")
	}

	if err := service.DeleteExperience(userID, uint(expID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "experience deleted successfully"})
}
