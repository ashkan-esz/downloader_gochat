package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type IWsHandler interface {
	AddClient(c *fiber.Ctx) error
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
	deviceId := c.Query("deviceId")

	err := w.wsService.AddClient(c.Context(), int64(userId), username, deviceId)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return err
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
