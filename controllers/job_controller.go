package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

// ─── Job CRUD ─────────────────────────────────────────────────────────────────

func CreateJobHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	role, err := getUserRoleFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	imageURL, err := uploadFileIfPresent(c, "image", "alumni-platform/jobs")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	req := service.JobRequest{
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		CompanyName: c.FormValue("company_name"),
		Location:    parseOptionalString(c.FormValue("location")),
		JobType:     c.FormValue("job_type"),
		Status:      c.FormValue("status"),
		SalaryRange: parseOptionalString(c.FormValue("salary_range")),
		ImageURL:    parseOptionalString(imageURL),
	}

	job, err := service.CreateJob(userID, role, req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya alumni atau partner yang dapat membuat lowongan" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, job)
}

func GetJobsHandler(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	jobType := c.Query("job_type", "")
	status := c.Query("status", "")

	jobs, total, err := service.GetJobs(search, jobType, status, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data lowongan")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"jobs":  jobs,
	})
}

func GetJobByIDHandler(c *fiber.Ctx) error {
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}
	job, err := service.GetJobByID(uint(jobID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, job)
}

func UpdateJobHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}

	imageURL, err := uploadFileIfPresent(c, "image", "alumni-platform/jobs")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	req := service.JobRequest{
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		CompanyName: c.FormValue("company_name"),
		Location:    parseOptionalString(c.FormValue("location")),
		JobType:     c.FormValue("job_type"),
		Status:      c.FormValue("status"),
		SalaryRange: parseOptionalString(c.FormValue("salary_range")),
		ImageURL:    parseOptionalString(imageURL),
	}

	job, err := service.UpdateJob(userID, uint(jobID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, job)
}

func DeleteJobHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}
	if err := service.DeleteJob(userID, uint(jobID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "lowongan berhasil dihapus"})
}

// ─── Applications ─────────────────────────────────────────────────────────────

func ApplyForJobHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	role, err := getUserRoleFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}

	// Resume: try file upload first, fall back to resume_url form field
	resumeURL, err := uploadRawFileIfPresent(c, "resume", "alumni-platform/resumes")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	if resumeURL == "" {
		resumeURL = c.FormValue("resume_url")
	}
	if resumeURL == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "resume wajib diunggah")
	}

	req := service.JobApplicationRequest{
		CoverLetter: parseOptionalString(c.FormValue("cover_letter")),
		ResumeURL:   resumeURL,
	}

	app, err := service.ApplyForJob(userID, role, uint(jobID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya student atau alumni yang dapat melamar" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, app)
}

func GetApplicantsHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}
	apps, err := service.GetApplicants(userID, uint(jobID))
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, apps)
}

func GetMyApplicationsHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	apps, err := service.GetMyApplications(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, apps)
}

func UpdateApplicationStatusHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	appID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lamaran tidak valid")
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil || body.Status == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	app, err := service.UpdateApplicationStatus(userID, uint(appID), body.Status)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, app)
}
