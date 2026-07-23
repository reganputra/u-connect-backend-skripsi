package controllers

import (
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
	topN, studentIDs := utils.ParseEvaluationParams(c, 50)

	result, err := ctrl.svc.EvaluateCBF(topN, studentIDs)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, result)
}

// EvaluateCBFWithoutLemmatizer mengevaluasi sistem rekomendasi CBF tanpa menggunakan Bilingual Lemmatizer.
func (ctrl *EvaluationController) EvaluateCBFWithoutLemmatizer(c *fiber.Ctx) error {
	topN, studentIDs := utils.ParseEvaluationParams(c, 50)

	result, err := ctrl.svc.EvaluateCBFWithoutLemmatizer(topN, studentIDs)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, result)
}

// EvaluateCBFMRR mengevaluasi sistem rekomendasi CBF menggunakan metrik MRR.
func (ctrl *EvaluationController) EvaluateCBFMRR(c *fiber.Ctx) error {
	topN, studentIDs := utils.ParseEvaluationParams(c, 50)

	result, err := ctrl.svc.EvaluateCBFMRR(topN, studentIDs)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, result)
}

// EvaluateCBFMRRWithoutLemmatizer mengevaluasi sistem rekomendasi CBF menggunakan metrik MRR tanpa Bilingual Lemmatizer.
func (ctrl *EvaluationController) EvaluateCBFMRRWithoutLemmatizer(c *fiber.Ctx) error {
	topN, studentIDs := utils.ParseEvaluationParams(c, 50)

	result, err := ctrl.svc.EvaluateCBFMRRWithoutLemmatizer(topN, studentIDs)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, result)
}

// ExplainCBF menjelaskan tahapan perhitungan rekomendasi CBF per mahasiswa terhadap mentor tertentu.
func (ctrl *EvaluationController) ExplainCBF(c *fiber.Ctx) error {
	studentID, ok := utils.MustParseIDParam(c, "student_id", "student")
	if !ok {
		return nil
	}
	mentorID, ok := utils.MustParseIDParam(c, "mentor_id", "mentor")
	if !ok {
		return nil
	}
	customQuery := c.Query("q", "")

	result, err := ctrl.svc.ExplainCBF(studentID, mentorID, customQuery)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, result)
}
