package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type IWsHandler interface {
	AddClient(c *fiber.Ctx) error
	CreateRoom(c *fiber.Ctx) error
	JoinRoom(c *fiber.Ctx) error
	GetRooms(c *fiber.Ctx) error
	GetClients(c *fiber.Ctx) error
	GetSingleChatMessages(c *fiber.Ctx) error
	GetSingleChatList(c *fiber.Ctx) error
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

func (w *WsHandler) AddClient(c *fiber.Ctx) error {
	userId := c.QueryInt("userId")
	username := c.Query("username")

	err := w.wsService.AddClient(c.Context(), int64(userId), username)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return err
}

func (w *WsHandler) CreateRoom(c *fiber.Ctx) error {
	var req model.CreateRoomReq
	if err := c.BodyParser(&req); err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	roomId, err := w.wsService.CreateRoom(req.SenderId, req.ReceiverId)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	res := model.CreateRoomRes{RoomId: roomId}

	return response.ResponseOKWithData(c, res)
}

func (w *WsHandler) JoinRoom(c *fiber.Ctx) error {
	roomId, _ := c.ParamsInt("roomId")
	userId := c.QueryInt("userId")
	username := c.Query("username")

	err := w.wsService.JoinRoom(c.Context(), int64(roomId), int64(userId), username)
	if err != nil {
		if err.Error() == "not found" {
			return response.ResponseError(c, err.Error(), fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return err
}

func (w *WsHandler) GetRooms(c *fiber.Ctx) error {
	rooms, _ := w.wsService.GetRooms()
	return response.ResponseOKWithData(c, rooms)
}

func (w *WsHandler) GetClients(c *fiber.Ctx) error {
	roomId, _ := c.ParamsInt("roomId")
	clients, _ := w.wsService.GetRoomClient(int64(roomId))
	return response.ResponseOKWithData(c, clients)
}

func (w *WsHandler) GetSingleChatMessages(c *fiber.Ctx) error {
	var params model.GetSingleMessagesReq
	err := c.QueryParser(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	messages, err := w.wsService.GetSingleChatMessages(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, messages)
}

func (w *WsHandler) GetSingleChatList(c *fiber.Ctx) error {
	var params model.GetSingleChatListReq
	err := c.QueryParser(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	messages, err := w.wsService.GetSingleChatList(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, messages)
}
