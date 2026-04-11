package service

import (
	"errors"
	"fmt"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

var validJobTypes = map[string]bool{
	"full-time":  true,
	"part-time":  true,
	"internship": true,
	"contract":   true,
	"freelance":  true,
}

var validJobStatuses = map[string]bool{
	"open":   true,
	"closed": true,
	"filled": true,
}

var validApplicationStatuses = map[string]bool{
	"pending":  true,
	"reviewed": true,
	"accepted": true,
	"rejected": true,
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type JobRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	CompanyName string  `json:"company_name"`
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
	GetJobByID(id uint) (*models.Job, error)
	UpdateJob(userID, jobID uint, req JobRequest) (*models.Job, error)
	DeleteJob(userID, jobID uint) error
	ApplyForJob(userID uint, role string, jobID uint, req JobApplicationRequest) (*models.JobApplication, error)
	GetApplicants(userID, jobID uint) ([]models.JobApplication, error)
	GetMyApplications(userID uint) ([]models.JobApplication, error)
	UpdateApplicationStatus(userID, applicationID uint, status string) (*models.JobApplication, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type jobService struct {
	jobRepo    repository.JobRepository
	jobAppRepo repository.JobApplicationRepository
	notifSvc   NotificationService
}

func NewJobService(jobRepo repository.JobRepository, jobAppRepo repository.JobApplicationRepository, notifSvc NotificationService) JobService {
	return &jobService{
		jobRepo:    jobRepo,
		jobAppRepo: jobAppRepo,
		notifSvc:   notifSvc,
	}
}

// ─── Job CRUD ─────────────────────────────────────────────────────────────────

func (s *jobService) CreateJob(userID uint, role string, req JobRequest) (*models.Job, error) {
	if role != "alumni" && role != "partner" {
		return nil, errors.New("akses ditolak: hanya alumni atau partner yang dapat membuat lowongan")
	}
	if req.Title == "" {
		return nil, errors.New("judul wajib diisi")
	}
	if req.CompanyName == "" {
		return nil, errors.New("nama perusahaan wajib diisi")
	}
	if req.JobType == "" {
		return nil, errors.New("tipe pekerjaan wajib diisi")
	}
	if !validJobTypes[req.JobType] {
		return nil, errors.New("tipe pekerjaan tidak valid: harus full-time, part-time, internship, contract, atau freelance")
	}
	status := req.Status
	if status == "" {
		status = "open"
	}
	if !validJobStatuses[status] {
		return nil, errors.New("status tidak valid: harus open, closed, atau filled")
	}

	job := &models.Job{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		CompanyName: req.CompanyName,
		Location:    req.Location,
		JobType:     req.JobType,
		Status:      status,
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
	if job.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	if req.Title != "" {
		job.Title = req.Title
	}
	if req.Description != nil {
		job.Description = req.Description
	}
	if req.CompanyName != "" {
		job.CompanyName = req.CompanyName
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
	if job.UserID != userID {
		return errors.New("akses ditolak")
	}
	if err := s.jobRepo.DeleteJob(jobID); err != nil {
		return errors.New("gagal menghapus lowongan")
	}
	return nil
}

// ─── Applications ─────────────────────────────────────────────────────────────

func (s *jobService) ApplyForJob(userID uint, role string, jobID uint, req JobApplicationRequest) (*models.JobApplication, error) {
	if role != "student" && role != "alumni" {
		return nil, errors.New("akses ditolak: hanya student atau alumni yang dapat melamar")
	}
	if req.ResumeURL == "" {
		return nil, errors.New("resume wajib diunggah")
	}

	if _, err := s.jobRepo.FindJobByID(jobID); err != nil {
		return nil, errors.New("lowongan tidak ditemukan")
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
		Status:      "pending",
	}
	if err := s.jobAppRepo.CreateJobApplication(app); err != nil {
		return nil, errors.New("gagal melamar pekerjaan")
	}
	return app, nil
}

func (s *jobService) GetApplicants(userID, jobID uint) ([]models.JobApplication, error) {
	job, err := s.jobRepo.FindJobByID(jobID)
	if err != nil {
		return nil, errors.New("lowongan tidak ditemukan")
	}
	if job.UserID != userID {
		return nil, errors.New("akses ditolak")
	}
	apps, err := s.jobAppRepo.FindApplicantsByJobID(jobID)
	if err != nil {
		return nil, errors.New("gagal mengambil data pelamar")
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

func (s *jobService) UpdateApplicationStatus(userID, applicationID uint, status string) (*models.JobApplication, error) {
	if !validApplicationStatuses[status] {
		return nil, errors.New("status tidak valid: harus pending, reviewed, accepted, atau rejected")
	}

	app, err := s.jobAppRepo.FindJobApplicationByID(applicationID)
	if err != nil {
		return nil, errors.New("lamaran tidak ditemukan")
	}

	if app.Job.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	app.Status = status
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
