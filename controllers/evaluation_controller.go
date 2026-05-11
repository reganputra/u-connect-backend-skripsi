package controllers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type EvaluationController struct {
	svc service.EvaluationService
}

func NewEvaluationController(svc service.EvaluationService) *EvaluationController {
	return &EvaluationController{svc: svc}
}

// Mengevaluasi sistem rekomendasi CBF menggunakan permintaan mentoring aktual sebagai ground truth dan mengembalikan MAP, P@1, P@3, P@5, serta rincian per mahasiswa.
// student_ids bersifat opsional — jika tidak diberikan, semua mahasiswa dengan riwayat permintaan akan dievaluasi.
func (ctrl *EvaluationController) EvaluateCBF(c *fiber.Ctx) error {
	topN, _ := strconv.Atoi(c.Query("top_n", "10"))
	if topN <= 0 || topN > 50 {
		topN = 10
	}

	var studentIDs []uint
	if raw := c.Query("student_ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.ParseUint(part, 10, 64); err == nil && id > 0 {
				studentIDs = append(studentIDs, uint(id))
			}
		}
	}

	result, err := ctrl.svc.EvaluateCBF(topN, studentIDs)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, result)
}
