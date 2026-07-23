package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/constant"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/utils"
	"gorm.io/gorm"
)

var validJobTypes = map[string]bool{
	constant.JobTypeFullTime:   true,
	constant.JobTypePartTime:   true,
	constant.JobTypeInternship: true,
	constant.JobTypeContract:   true,
	constant.JobTypeFreelance:  true,
}

var validJobStatuses = constant.JobStatuses

var validApplicationStatuses = constant.ApplicationStatuses

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type JobRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	CompanyName string  `json:"company_name"`
	Openings    *int    `json:"openings"`
	Location    *string `json:"location"`
	JobType     string  `json:"job_type"`
	Status      string  `json:"status"`
	SalaryRange *string `json:"salary_range"`
	ImageURL    *string `json:"image_url"`
}

type JobApplicationRequest struct {
	CoverLetter *string `json:"cover_letter"`
	ResumeURL   string  `json:"resume_url"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type JobService interface {
	CreateJob(userID uint, role string, req JobRequest) (*models.Job, error)
	GetJobs(search, jobType, status string, page, limit int) ([]models.Job, int64, error)
	GetMyOwnedJobs(userID uint, page, limit int) ([]models.Job, int64, error)
	GetJobByID(id uint) (*models.Job, error)
	UpdateJob(userID, jobID uint, req JobRequest) (*models.Job, error)
	DeleteJob(userID, jobID uint) error
	ApplyForJob(userID uint, role string, jobID uint, req JobApplicationRequest) (*models.JobApplication, error)
	WithdrawApplication(userID uint, role string, jobID uint) error
	GetApplicants(userID, jobID uint) ([]models.JobApplication, error)
	GetMyApplications(userID uint) ([]models.JobApplication, error)
	CountMyApplications(userID uint) (int64, error)
	UpdateApplicationStatus(userID, applicationID uint, status string) (*models.JobApplication, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type jobService struct {
	jobRepo     repository.JobRepository
	jobAppRepo  repository.JobApplicationRepository
	companyRepo repository.CompanyRepository
	userRepo    repository.UserRepository
	notifSvc    NotificationService
}

func NewJobService(jobRepo repository.JobRepository, jobAppRepo repository.JobApplicationRepository, companyRepo repository.CompanyRepository, userRepo repository.UserRepository, notifSvc NotificationService) JobService {
	return &jobService{
		jobRepo:     jobRepo,
		jobAppRepo:  jobAppRepo,
		companyRepo: companyRepo,
		userRepo:    userRepo,
		notifSvc:    notifSvc,
	}
}

func (s *jobService) resolveJobCompany(userID uint, role string, reqCompanyName string) (string, *uint, error) {
	companyName := strings.TrimSpace(reqCompanyName)
	if companyName == "" {
		return "", nil, errors.New("nama perusahaan wajib diisi")
	}

	if role == constant.RolePartner {
		user, err := s.userRepo.FindUserByID(userID)
		if err != nil {
			return "", nil, errors.New("pengguna tidak ditemukan")
		}
		if user.CompanyName == nil || strings.TrimSpace(*user.CompanyName) == "" {
			return "", nil, errors.New("partner harus memiliki profil perusahaan sebelum membuat lowongan")
		}

		profile, err := s.companyRepo.FindCompanyProfileByName(strings.TrimSpace(*user.CompanyName))
		if err != nil {
			return "", nil, errors.New("profil perusahaan tidak ditemukan")
		}
		if strings.TrimSpace(profile.CompanyName) != companyName {
			return "", nil, errors.New("nama perusahaan harus sesuai dengan profil perusahaan Anda")
		}
		companyID := profile.ID
		return profile.CompanyName, &companyID, nil
	}

	if existing, err := s.companyRepo.FindCompanyProfileByName(companyName); err == nil && existing != nil {
		companyID := existing.ID
		return existing.CompanyName, &companyID, nil
	}

	return companyName, nil, nil
}

func (s *jobService) syncJobOpeningState(job *models.Job) {
	if job.Openings <= 0 {
		job.Openings = 0
		job.Status = constant.StatusFilled
		return
	}

	if job.Openings == 1 {
		job.Status = constant.StatusFilled
		return
	}
}

// ─── Job CRUD ─────────────────────────────────────────────────────────────────

func (s *jobService) CreateJob(userID uint, role string, req JobRequest) (*models.Job, error) {
	if role != constant.RoleAlumni && role != constant.RolePartner {
		return nil, errors.New("akses ditolak: hanya alumni atau partner yang dapat membuat lowongan")
	}
	if req.Title == "" {
		return nil, errors.New("judul wajib diisi")
	}
	if req.JobType == "" {
		return nil, errors.New("tipe pekerjaan wajib diisi")
	}
	if !validJobTypes[req.JobType] {
		return nil, errors.New("tipe pekerjaan tidak valid: harus full-time, part-time, internship, contract, atau freelance")
	}
	openings := 1
	if req.Openings != nil {
		if *req.Openings <= 0 {
			return nil, errors.New("jumlah lowongan harus lebih besar dari nol")
		}
		openings = *req.Openings
	}
	status := req.Status
	if status == "" {
		status = constant.StatusOpen
	}
	if !validJobStatuses[status] {
		return nil, errors.New("status tidak valid: harus open, closed, atau filled")
	}

	companyName, companyID, err := s.resolveJobCompany(userID, role, req.CompanyName)
	if err != nil {
		return nil, err
	}

	job := &models.Job{
		UserID:      userID,
		CompanyID:   companyID,
		Title:       req.Title,
		Description: req.Description,
		CompanyName: companyName,
		Location:    req.Location,
		JobType:     req.JobType,
		Status:      status,
		Openings:    openings,
		SalaryRange: req.SalaryRange,
		ImageURL:    req.ImageURL,
	}
	if err := s.jobRepo.CreateJob(job); err != nil {
		return nil, errors.New("gagal membuat lowongan")
	}
	return job, nil
}

func (s *jobService) GetJobs(search, jobType, status string, page, limit int) ([]models.Job, int64, error) {
	offset := (page - 1) * limit
	return s.jobRepo.FindJobs(search, jobType, status, offset, limit)
}

func (s *jobService) GetMyOwnedJobs(userID uint, page, limit int) ([]models.Job, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	return s.jobRepo.FindJobsByOwner(userID, offset, limit)
}

func (s *jobService) GetJobByID(id uint) (*models.Job, error) {
	job, err := s.jobRepo.FindJobByID(id)
	if err != nil {
		return nil, errors.New("lowongan tidak ditemukan")
	}
	return job, nil
}

func (s *jobService) UpdateJob(userID, jobID uint, req JobRequest) (*models.Job, error) {
	job, err := s.jobRepo.FindJobByID(jobID)
	if err != nil {
		return nil, errors.New("lowongan tidak ditemukan")
	}
	if !IsOwnerOrAdmin(s.userRepo, userID, job.UserID) {
		return nil, errors.New("akses ditolak")
	}

	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}

	if req.Title != "" {
		job.Title = req.Title
	}
	if req.Description != nil {
		job.Description = req.Description
	}
	if req.CompanyName != "" {
		companyName, companyID, err := s.resolveJobCompany(userID, user.Role, req.CompanyName)
		if err != nil {
			return nil, err
		}
		job.CompanyName = companyName
		job.CompanyID = companyID
	}
	if req.Location != nil {
		job.Location = req.Location
	}
	if req.ImageURL != nil {
		job.ImageURL = req.ImageURL
	}
	if req.SalaryRange != nil {
		job.SalaryRange = req.SalaryRange
	}
	if req.JobType != "" {
		if !validJobTypes[req.JobType] {
			return nil, errors.New("tipe pekerjaan tidak valid: harus full-time, part-time, internship, contract, atau freelance")
		}
		job.JobType = req.JobType
	}
	if req.Status != "" {
		if !validJobStatuses[req.Status] {
			return nil, errors.New("status tidak valid: harus open, closed, atau filled")
		}
		job.Status = req.Status
	}
	if req.Openings != nil {
		if *req.Openings <= 0 {
			return nil, errors.New("jumlah lowongan harus lebih besar dari nol")
		}
		job.Openings = *req.Openings
	}

	if err := s.jobRepo.UpdateJob(job); err != nil {
		return nil, errors.New("gagal memperbarui lowongan")
	}
	return job, nil
}

func (s *jobService) DeleteJob(userID, jobID uint) error {
	job, err := s.jobRepo.FindJobByID(jobID)
	if err != nil {
		return errors.New("lowongan tidak ditemukan")
	}
	if !IsOwnerOrAdmin(s.userRepo, userID, job.UserID) {
		return errors.New("akses ditolak")
	}
	if err := s.jobRepo.DeleteJob(jobID); err != nil {
		return errors.New("gagal menghapus lowongan")
	}
	return nil
}

// ─── Applications ─────────────────────────────────────────────────────────────

func (s *jobService) ApplyForJob(userID uint, role string, jobID uint, req JobApplicationRequest) (*models.JobApplication, error) {
	if role != constant.RoleStudent && role != constant.RoleAlumni {
		return nil, errors.New("akses ditolak: hanya student atau alumni yang dapat melamar")
	}
	if req.ResumeURL == "" {
		return nil, errors.New("resume wajib diunggah")
	}

	job, err := s.jobRepo.FindJobByID(jobID)
	if err != nil {
		return nil, errors.New("lowongan tidak ditemukan")
	}
	if job.Status != constant.StatusOpen {
		return nil, errors.New("lowongan sudah tidak menerima lamaran")
	}

	existing, err := s.jobAppRepo.FindJobApplication(jobID, userID)
	if err == nil && existing != nil {
		return nil, errors.New("sudah melamar pekerjaan ini")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.New("gagal memeriksa lamaran")
	}

	app := &models.JobApplication{
		JobID:       jobID,
		UserID:      userID,
		CoverLetter: req.CoverLetter,
		ResumeURL:   req.ResumeURL,
		Status:      constant.StatusPending,
	}
	if err := s.jobAppRepo.CreateJobApplication(app); err != nil {
		return nil, errors.New("gagal melamar pekerjaan")
	}
	return app, nil
}

func (s *jobService) WithdrawApplication(userID uint, role string, jobID uint) error {
	if role != constant.RoleStudent && role != constant.RoleAlumni {
		return errors.New("akses ditolak: hanya student atau alumni yang dapat menarik lamaran")
	}

	job, err := s.jobRepo.FindJobByID(jobID)
	if err != nil {
		return errors.New("lowongan tidak ditemukan")
	}
	if job.Status == constant.StatusFilled {
		return errors.New("lamaran tidak dapat ditarik untuk lowongan yang sudah penuh")
	}

	app, err := s.jobAppRepo.FindJobApplication(jobID, userID)
	if err != nil || app == nil {
		return errors.New("lamaran tidak ditemukan")
	}
	if app.Status != constant.StatusPending {
		return errors.New("lamaran hanya bisa ditarik saat masih pending")
	}

	app.Status = constant.StatusWithdrawn
	return s.jobAppRepo.UpdateJobApplication(app)
}

func (s *jobService) GetApplicants(userID, jobID uint) ([]models.JobApplication, error) {
	job, err := s.jobRepo.FindJobByID(jobID)
	if err != nil {
		return nil, errors.New("lowongan tidak ditemukan")
	}
	if !IsOwnerOrAdmin(s.userRepo, userID, job.UserID) {
		return nil, errors.New("akses ditolak")
	}
	apps, err := s.jobAppRepo.FindApplicantsByJobID(jobID)
	if err != nil {
		return nil, errors.New("gagal mengambil data pelamar")
	}

	for i := range apps {
		ApplyUserPicture(&apps[i].User)
		resumeURL := strings.TrimSpace(apps[i].ResumeURL)
		if resumeURL == "" || !strings.Contains(strings.ToLower(resumeURL), "cloudinary.com") {
			continue
		}

		signedURL, signErr := utils.BuildCloudinaryTemporaryDownloadURL(config.Cloudinary, resumeURL, time.Hour)
		if signErr == nil && signedURL != "" {
			apps[i].ResumeURL = signedURL
		}
	}

	return apps, nil
}

func (s *jobService) GetMyApplications(userID uint) ([]models.JobApplication, error) {
	apps, err := s.jobAppRepo.FindApplicationsByUserID(userID)
	if err != nil {
		return nil, errors.New("gagal mengambil lamaran saya")
	}
	return apps, nil
}

func (s *jobService) CountMyApplications(userID uint) (int64, error) {
	total, err := s.jobAppRepo.CountApplicationsByUserID(userID)
	if err != nil {
		return 0, errors.New("gagal menghitung lamaran saya")
	}
	return total, nil
}

func (s *jobService) UpdateApplicationStatus(userID, applicationID uint, status string) (*models.JobApplication, error) {
	if !validApplicationStatuses[status] {
		return nil, errors.New("status tidak valid: harus pending, reviewed, accepted, rejected, atau withdrawn")
	}

	app, err := s.jobAppRepo.FindJobApplicationByID(applicationID)
	if err != nil {
		return nil, errors.New("lamaran tidak ditemukan")
	}

	if !IsOwnerOrAdmin(s.userRepo, userID, app.Job.UserID) {
		return nil, errors.New("akses ditolak")
	}
	if app.Status == constant.StatusWithdrawn {
		return nil, errors.New("lamaran sudah ditarik")
	}
	previousStatus := app.Status

	app.Status = status
	if status == constant.ApplicationStatusAccepted && previousStatus != constant.ApplicationStatusAccepted {
		job, err := s.jobRepo.FindJobByID(app.JobID)
		if err != nil {
			return nil, errors.New("lowongan tidak ditemukan")
		}
		if job.Openings <= 0 {
			return nil, errors.New("lowongan sudah penuh")
		}
		job.Openings--
		if job.Openings <= 0 {
			job.Openings = 0
			job.Status = constant.StatusFilled
		}
		if err := s.jobRepo.UpdateJob(job); err != nil {
			return nil, errors.New("gagal memperbarui status lowongan")
		}
	} else if previousStatus == constant.ApplicationStatusAccepted && status != constant.ApplicationStatusAccepted {
		job, err := s.jobRepo.FindJobByID(app.JobID)
		if err != nil {
			return nil, errors.New("lowongan tidak ditemukan")
		}
		job.Openings++
		if job.Status == constant.StatusFilled && job.Openings > 0 {
			job.Status = constant.StatusOpen
		}
		if err := s.jobRepo.UpdateJob(job); err != nil {
			return nil, errors.New("gagal memperbarui status lowongan")
		}
	}

	if err := s.jobAppRepo.UpdateJobApplication(app); err != nil {
		return nil, errors.New("gagal memperbarui status lamaran")
	}
	// Notify applicant
	_ = s.notifSvc.Notify(
		app.UserID,
		"job_application_updated",
		"Status lamaran diperbarui",
		fmt.Sprintf("Lamaranmu untuk %s diperbarui menjadi %s", app.Job.Title, status),
		"job_application",
		app.ID,
	)
	return app, nil
}
