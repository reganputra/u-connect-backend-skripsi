package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type ActivityController struct {
	eventSvc service.EventService
	jobSvc   service.JobService
	groupSvc service.GroupService
}

func NewActivityController(eventSvc service.EventService, jobSvc service.JobService, groupSvc service.GroupService) *ActivityController {
	return &ActivityController{
		eventSvc: eventSvc,
		jobSvc:   jobSvc,
		groupSvc: groupSvc,
	}
}

func (ctrl *ActivityController) GetMyActivitySummary(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	_, ownedEventsTotal, err := ctrl.eventSvc.GetMyOwnedEvents(userID, 1, 1)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal menghitung aktivitas acara milik pengguna")
	}

	_, registeredEventsTotal, err := ctrl.eventSvc.GetMyRegisteredEvents(userID, 1, 1)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal menghitung aktivitas pendaftaran acara")
	}

	_, ownedJobsTotal, err := ctrl.jobSvc.GetMyOwnedJobs(userID, 1, 1)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal menghitung aktivitas lowongan milik pengguna")
	}

	appliedJobsTotal, err := ctrl.jobSvc.CountMyApplications(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal menghitung aktivitas lamaran pekerjaan")
	}

	_, joinedGroupsTotal, err := ctrl.groupSvc.GetJoinedGroups(userID, 1, 1)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal menghitung aktivitas grup diikuti")
	}

	_, ownedGroupsTotal, err := ctrl.groupSvc.GetOwnedGroups(userID, 1, 1)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal menghitung aktivitas grup milik pengguna")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"events_owned":      ownedEventsTotal,
		"events_registered": registeredEventsTotal,
		"jobs_owned":        ownedJobsTotal,
		"jobs_applied":      appliedJobsTotal,
		"groups_owned":      ownedGroupsTotal,
		"groups_joined":     joinedGroupsTotal,
	})
}
