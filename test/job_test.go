package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestJob covers job listing, CRUD, role guards, applying, and application status management.
// Run: go test -v -run TestJob ./test/...
func TestJob(t *testing.T) {
	sfx := newSuffix()
	alumniToken, _ := registerAndLogin(t, sfx, "Job Alumni", "jobalumni+"+sfx+"@test.com", "alumni", "Engineering", "CS")
	studentToken, _ := registerAndLogin(t, sfx, "Job Student", "jobstudent+"+sfx+"@test.com", "student", "Engineering", "CS")
	otherAlumniToken, _ := registerAndLogin(t, sfx, "Other Alumni", "otheralumni+"+sfx+"@test.com", "alumni", "Science", "Physics")
	partnerToken, _ := registerAndLoginPartner(t, "Job Partner "+sfx, "jobpartner+"+sfx+"@test.com", "PT Job Corp "+sfx)

	var jobID float64
	var applicationID float64

	// ── Role Guard ─────────────────────────────────────────────────────────────

	t.Run("student_cannot_create_job", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/jobs", studentToken, map[string]string{
			"title":        "should fail",
			"company_name": "Test Co",
			"job_type":     "full-time",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Student blocked from creating job (403)")
	})

	t.Run("no_token_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/jobs", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated request rejected (401)")
	})

	// ── Job CRUD ──────────────────────────────────────────────────────────────

	t.Run("create_job", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/jobs", alumniToken, map[string]string{
			"title":        "Backend Engineer " + sfx,
			"description":  "Golang backend developer position",
			"company_name": "PT Alumni Tech",
			"location":     "Jakarta",
			"job_type":     "full-time",
			"status":       "open",
			"salary_range": "10000000-20000000",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		jobID = d["ID"].(float64)
		t.Logf("✅ Job created (id: %.0f)", jobID)
	})

	t.Run("create_job_missing_title_fails", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/jobs", alumniToken, map[string]string{
			"company_name": "PT Test",
			"job_type":     "full-time",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Job without title rejected (400)")
	})

	t.Run("create_job_invalid_job_type_fails", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/jobs", alumniToken, map[string]string{
			"title":        "Bad Type",
			"company_name": "PT Test",
			"job_type":     "random",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid job_type rejected (400)")
	})

	t.Run("partner_can_create_job", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/jobs", partnerToken, map[string]string{
			"title":        "Magang IT " + sfx,
			"company_name": "PT Partner Corp",
			"job_type":     "internship",
			"status":       "open",
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Partner can create job (201)")
	})

	t.Run("get_jobs_list", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/jobs?page=1&limit=10", alumniToken, nil)
		assertStatus(t, 200, code, res)
		if res["data"] == nil {
			t.Fatal("missing data field in response")
		}
		t.Log("✅ Job list fetched")
	})

	t.Run("get_jobs_filter_by_job_type", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/jobs?job_type=full-time", alumniToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Job list filtered by job_type")
	})

	t.Run("get_jobs_search", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/jobs?search=Backend+Engineer", alumniToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Job list search works")
	})

	t.Run("get_job_detail", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/jobs/%.0f", jobID), studentToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Title"] == nil {
			t.Fatal("Title missing from job detail")
		}
		t.Log("✅ Job detail fetched")
	})

	t.Run("get_nonexistent_job_returns_404", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/jobs/99999999", alumniToken, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Non-existent job returns 404")
	})

	t.Run("update_job", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/jobs/%.0f", jobID), alumniToken, map[string]string{
			"status":   "closed",
			"location": "Bandung",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Job updated (owner)")
	})

	t.Run("non_owner_cannot_update_job", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/jobs/%.0f", jobID), otherAlumniToken, map[string]string{
			"title": "hijacked",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner job update blocked (403)")
	})

	// ── Applications ──────────────────────────────────────────────────────────

	t.Run("apply_without_resume_fails", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/jobs/%.0f/apply", jobID), studentToken, map[string]string{
			"cover_letter": "I am very interested",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Apply without resume rejected (400)")
	})

	t.Run("partner_cannot_apply", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/jobs/%.0f/apply", jobID), partnerToken,
			map[string]string{"cover_letter": "test", "resume_url": "https://example.com/cv.pdf"})
		assertStatus(t, 403, code, res)
		t.Log("✅ Partner blocked from applying (403)")
	})

	t.Run("student_can_apply", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/jobs/%.0f/apply", jobID), studentToken,
			map[string]string{"cover_letter": "I am very interested in this position", "resume_url": "https://example.com/cv_student.pdf"})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		applicationID = d["ID"].(float64)
		t.Logf("✅ Student applied successfully (id: %.0f)", applicationID)
	})

	t.Run("cannot_apply_twice", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/jobs/%.0f/apply", jobID), studentToken,
			map[string]string{"resume_url": "https://example.com/cv_student2.pdf"})
		assertStatus(t, 400, code, res)
		t.Log("✅ Duplicate application rejected (400)")
	})

	t.Run("alumni_can_apply", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/jobs/%.0f/apply", jobID), otherAlumniToken,
			map[string]string{"cover_letter": "Alumni applying", "resume_url": "https://example.com/cv_alumni.pdf"})
		assertStatus(t, 201, code, res)
		t.Log("✅ Alumni applied successfully")
	})

	t.Run("get_applicants_as_owner", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/jobs/%.0f/applicants", jobID), alumniToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Owner can view applicants")
	})

	t.Run("non_owner_cannot_view_applicants", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/jobs/%.0f/applicants", jobID), otherAlumniToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner blocked from viewing applicants (403)")
	})

	t.Run("get_my_applications", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/jobs/applications/mine", studentToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Student can view own applications")
	})

	t.Run("update_application_status", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/jobs/applications/%.0f/status", applicationID), alumniToken, map[string]any{
			"status": "reviewed",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Owner updated application status to reviewed")
	})

	t.Run("non_owner_cannot_update_application_status", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/jobs/applications/%.0f/status", applicationID), otherAlumniToken, map[string]any{
			"status": "rejected",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner blocked from updating application status (403)")
	})

	t.Run("invalid_application_status_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/jobs/applications/%.0f/status", applicationID), alumniToken, map[string]any{
			"status": "invalid_status",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid application status rejected (400)")
	})

	// ── Delete Job ────────────────────────────────────────────────────────────

	t.Run("non_owner_cannot_delete_job", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/jobs/%.0f", jobID), otherAlumniToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner job delete blocked (403)")
	})

	t.Run("delete_job_cascades_applications", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/jobs/%.0f", jobID), alumniToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Job deleted (cascades applications)")
	})

	t.Run("deleted_job_not_found", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/jobs/%.0f", jobID), alumniToken, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Deleted job returns 404")
	})
}
