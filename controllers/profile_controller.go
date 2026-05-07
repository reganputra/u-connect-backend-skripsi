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

// ─── Token helpers (shared across all controllers in this package) ────────────

func getUserIDFromToken(c *fiber.Ctx) (uint, error) {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return 0, fmt.Errorf("token tidak valid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("klaim token tidak valid")
	}
	idFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id tidak valid dalam token")
	}
	return uint(idFloat), nil
}

func getUserRoleFromToken(c *fiber.Ctx) (string, error) {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("token tidak valid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("klaim token tidak valid")
	}
	if role, ok := claims["role"].(string); ok && role != "" {
		return role, nil
	}
	if role, ok := claims["Role"].(string); ok && role != "" {
		return role, nil
	}
	return "", fmt.Errorf("peran pengguna tidak ditemukan")
}

// parseOptionalString returns a *string if the form value is non-empty, else nil.
func parseOptionalString(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}

// parseOptionalInt returns a *int if the form value is a valid non-zero int, else nil.
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

// uploadFileIfPresent uploads an image form file to Cloudinary and returns its URL.
// Returns empty string and nil error if no file was provided.
func uploadFileIfPresent(c *fiber.Ctx, fieldName, folder string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil || file == nil {
		return "", nil
	}
	url, err := utils.UploadImage(config.Cloudinary, file, folder)
	if err != nil {
		return "", err
	}
	return url, nil
}

// uploadRawFileIfPresent uploads any file type (e.g. PDF resume) to Cloudinary.
// Returns empty string and nil error if no file was provided.
func uploadRawFileIfPresent(c *fiber.Ctx, fieldName, folder string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil || file == nil {
		return "", nil
	}
	url, err := utils.UploadFile(config.Cloudinary, file, folder)
	if err != nil {
		return "", err
	}
	return url, nil
}

// ─── ProfileController ────────────────────────────────────────────────────────

type ProfileController struct {
	profileSvc service.ProfileService
}

func NewProfileController(profileSvc service.ProfileService) *ProfileController {
	return &ProfileController{profileSvc: profileSvc}
}

func (ctrl *ProfileController) CreateProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	pictureURL, err := uploadFileIfPresent(c, "picture", "alumni-platform/profiles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	yearEnroll := parseOptionalInt(c.FormValue("year_enroll"))
	salary := parseOptionalInt(c.FormValue("salary"))
	yearFounding := parseOptionalInt(c.FormValue("year_founding"))
	companySize := parseOptionalInt(c.FormValue("company_size"))
	mentorQuota := parseOptionalInt(c.FormValue("mentor_quota"))
	expectedGradYear := parseOptionalInt(c.FormValue("expected_graduation_year"))

	req := service.ProfileRequest{
		ProfilePicture:         parseOptionalString(pictureURL),
		Bio:                    parseOptionalString(c.FormValue("bio")),
		Location:               parseOptionalString(c.FormValue("location")),
		JobStatus:              parseOptionalString(c.FormValue("job_status")),
		Position:               parseOptionalString(c.FormValue("position")),
		CompanyName:            parseOptionalString(c.FormValue("company_name")),
		CompanyLocation:        parseOptionalString(c.FormValue("company_location")),
		CompanySize:            companySize,
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

	_ = yearEnroll

	profile, err := ctrl.profileSvc.CreateProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, profile)
}

func (ctrl *ProfileController) GetProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	profile, err := ctrl.profileSvc.GetProfile(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

func (ctrl *ProfileController) UpdateProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	pictureURL, err := uploadFileIfPresent(c, "picture", "alumni-platform/profiles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	salary := parseOptionalInt(c.FormValue("salary"))
	yearFounding := parseOptionalInt(c.FormValue("year_founding"))
	companySize := parseOptionalInt(c.FormValue("company_size"))
	mentorQuota := parseOptionalInt(c.FormValue("mentor_quota"))
	expectedGradYear := parseOptionalInt(c.FormValue("expected_graduation_year"))

	req := service.ProfileRequest{
		Name:                   parseOptionalString(c.FormValue("name")),
		ProfilePicture:         parseOptionalString(pictureURL),
		Bio:                    parseOptionalString(c.FormValue("bio")),
		Location:               parseOptionalString(c.FormValue("location")),
		JobStatus:              parseOptionalString(c.FormValue("job_status")),
		Position:               parseOptionalString(c.FormValue("position")),
		CompanyName:            parseOptionalString(c.FormValue("company_name")),
		CompanyLocation:        parseOptionalString(c.FormValue("company_location")),
		CompanySize:            companySize,
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

	profile, err := ctrl.profileSvc.UpdateProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

func (ctrl *ProfileController) DeleteProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	if err := ctrl.profileSvc.DeleteProfile(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "profil berhasil dihapus"})
}

func (ctrl *ProfileController) AddExperience(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	var req service.ExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	exp, err := ctrl.profileSvc.AddExperience(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, exp)
}

func (ctrl *ProfileController) UpdateExperience(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	expID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID pengalaman tidak valid")
	}

	var req service.ExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	exp, err := ctrl.profileSvc.UpdateExperience(userID, uint(expID), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, exp)
}

func (ctrl *ProfileController) DeleteExperience(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	expID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID pengalaman tidak valid")
	}

	if err := ctrl.profileSvc.DeleteExperience(userID, uint(expID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "pengalaman berhasil dihapus"})
}
