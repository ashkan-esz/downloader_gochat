package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type IWsHandler interface {
	CreateRoom(c *fiber.Ctx) error
	JoinRoom(c *fiber.Ctx) error
	GetRooms(c *fiber.Ctx) error
	GetClients(c *fiber.Ctx) error
}

type WsHandler struct {
	wsService service.IWsService
}

func NewWsHandler(wsService service.IWsService) *WsHandler {
	return &WsHandler{
		wsService: wsService,
	}
}

//------------------------------------------
//------------------------------------------

func (w *WsHandler) CreateRoom(c *fiber.Ctx) error {
	var req model.CreateRoomReq
	if err := c.BodyParser(&req); err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	_ = w.wsService.CreateRoom(req.ID, req.Name)

	return response.ResponseOKWithData(c, req)
}

func (w *WsHandler) JoinRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	clientId := c.Query("userId")
	username := c.Query("username")

	err := w.wsService.JoinRoom(c.Context(), roomId, clientId, username)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return err
}

func (w *WsHandler) GetRooms(c *fiber.Ctx) error {
	rooms, _ := w.wsService.GetRooms()
	return response.ResponseOKWithData(c, rooms)
}

func (w *WsHandler) GetClients(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	clients, _ := w.wsService.GetRoomClient(roomId)
	return response.ResponseOKWithData(c, clients)
}
