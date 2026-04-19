package controllers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type JobController struct {
	jobSvc service.JobService
}

func NewJobController(jobSvc service.JobService) *JobController {
	return &JobController{jobSvc: jobSvc}
}

// ─── Job CRUD ─────────────────────────────────────────────────────────────────

func (ctrl *JobController) CreateJobHandler(c *fiber.Ctx) error {
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

	job, err := ctrl.jobSvc.CreateJob(userID, role, req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya alumni atau partner yang dapat membuat lowongan" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, job)
}

func (ctrl *JobController) GetJobsHandler(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	jobType := c.Query("job_type", "")
	status := c.Query("status", "")

	jobs, total, err := ctrl.jobSvc.GetJobs(search, jobType, status, page, limit)
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

func (ctrl *JobController) GetJobByIDHandler(c *fiber.Ctx) error {
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}
	job, err := ctrl.jobSvc.GetJobByID(uint(jobID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, job)
}

func (ctrl *JobController) UpdateJobHandler(c *fiber.Ctx) error {
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

	job, err := ctrl.jobSvc.UpdateJob(userID, uint(jobID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, job)
}

func (ctrl *JobController) DeleteJobHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}
	if err := ctrl.jobSvc.DeleteJob(userID, uint(jobID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "lowongan berhasil dihapus"})
}

// ─── Applications ─────────────────────────────────────────────────────────────

func (ctrl *JobController) ApplyForJobHandler(c *fiber.Ctx) error {
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

	resumeURL, err := uploadRawFileIfPresent(c, "resume", "alumni-platform/resumes")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	usedResumeURLFallback := false
	if resumeURL == "" {
		resumeURL = strings.TrimSpace(c.FormValue("resume_url"))
		usedResumeURLFallback = resumeURL != ""
	}
	if usedResumeURLFallback {
		isCloudinaryURL := strings.Contains(resumeURL, "res.cloudinary.com")
		isImageDelivery := strings.Contains(resumeURL, "/image/upload/")
		isAuthenticatedRaw := strings.Contains(resumeURL, "/raw/authenticated/")
		if isCloudinaryURL && (isImageDelivery || isAuthenticatedRaw) {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "resume_url Cloudinary tidak valid: gunakan file upload `resume` atau URL `/raw/upload/`")
		}
	}
	if resumeURL == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "resume wajib diunggah")
	}

	req := service.JobApplicationRequest{
		CoverLetter: parseOptionalString(c.FormValue("cover_letter")),
		ResumeURL:   resumeURL,
	}

	app, err := ctrl.jobSvc.ApplyForJob(userID, role, uint(jobID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya student atau alumni yang dapat melamar" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, app)
}

func (ctrl *JobController) GetApplicantsHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	jobID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID lowongan tidak valid")
	}
	apps, err := ctrl.jobSvc.GetApplicants(userID, uint(jobID))
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, apps)
}

func (ctrl *JobController) GetMyApplicationsHandler(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	apps, err := ctrl.jobSvc.GetMyApplications(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, apps)
}

func (ctrl *JobController) UpdateApplicationStatusHandler(c *fiber.Ctx) error {
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

	app, err := ctrl.jobSvc.UpdateApplicationStatus(userID, uint(appID), body.Status)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, app)
}
