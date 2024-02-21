package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"

	"github.com/gofiber/fiber/v2"
)

type INotificationHandler interface {
	GetUserNotifications(c *fiber.Ctx) error
	BatchUpdateUserNotificationStatus(c *fiber.Ctx) error
}

type NotificationHandler struct {
	notifService service.INotificationService
}

func NewNotificationHandler(notifService service.INotificationService) *NotificationHandler {
	return &NotificationHandler{
		notifService: notifService,
	}
}

//------------------------------------------
//------------------------------------------

// GetUserNotifications godoc
//
//	@Summary		Follow events
//	@Description	get user followers/followings events
//	@Tags			User
//	@Param			skip		path		integer				true   "skip"
//	@Param			limit		path		integer				true   "limit"
//	@Param			entityTypeId		query		integer				true   "entityTypeId"
//	@Param			status		query		integer				true   "status"
//	@Success		200				{object}	model.NotificationDataModel
//	@Failure		400,401,404			{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/notifications/:skip/:limit [get]
func (n *NotificationHandler) GetUserNotifications(c *fiber.Ctx) error {
	skip, err := c.ParamsInt("skip", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	limit, err := c.ParamsInt("limit", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	entityTypeId := c.QueryInt("entityTypeId", 0)
	status := c.QueryInt("status", 0)
	autoUpdateStatus := c.QueryBool("autoUpdateStatus", false)

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	result, err := n.notifService.GetUserNotifications(jwtUserData.UserId, skip, limit, entityTypeId, status, autoUpdateStatus)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, result)
}

// BatchUpdateUserNotificationStatus godoc
//
//	@Summary		Notification Status update
//	@Description	update the status of notifications
//	@Tags			User
//	@Param			id		path		integer				true   "notificationId"
//	@Param			entityTypeId		path		integer				true   "type of notification"
//	@Param			status		path		integer				true   "new value of status"
//	@Success		200				{object}	response.ResponseOKModel
//	@Failure		400,401,404			{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/notifications/batchUpdateStatus/:id/:entityTypeId/:status [put]
func (n *NotificationHandler) BatchUpdateUserNotificationStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	entityTypeId, err := c.ParamsInt("entityTypeId", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	status, err := c.ParamsInt("status", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	if status > 2 {
		status = 2
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	err = n.notifService.BatchUpdateUserNotificationStatus(jwtUserData.UserId, int64(id), entityTypeId, status)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOK(c, "")
}
