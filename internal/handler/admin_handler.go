package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type IAdminHandler interface {
	GetServerStatus(c *fiber.Ctx) error
}

type AdminHandler struct {
	adminService service.IAdminService
}

func NewAdminHandler(adminService service.IAdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

//------------------------------------------
//------------------------------------------

// GetServerStatus godoc
//
//	@Summary		Server Status
//	@Description	Return status of server resources and services
//	@Tags			Admin-Status
//	@Success		200	{object}	model.Status
//	@Failure		400	{object}	response.ResponseErrorModel
//	@Router			/v1/admin/status [get]
func (a *AdminHandler) GetServerStatus(c *fiber.Ctx) error {
	result := a.adminService.GetServerStatus()

	return response.ResponseOKWithData(c, result)
}
