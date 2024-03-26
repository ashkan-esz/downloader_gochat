package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"

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

// AddClient godoc
//
//	@Summary		Connect websocket
//	@Description	start websocket connection
//	@Tags			User-Websocket
//	@Param			deviceId	path		string				true	"unique id of the device"
//	@Param			messageBody	body		model.ClientMessage	true	"types of bodies can be handled in server"
//	@Success		200			{object}	model.ServerResultMessage
//	@Failure		400			{object}	response.ResponseErrorModel
//	@Router			/v1/ws/addClient/:deviceId [get]
func (w *WsHandler) AddClient(c *fiber.Ctx) error {
	deviceId := c.Params("deviceId", "")
	if deviceId == "" || deviceId == ":deviceId" {
		return response.ResponseError(c, response.InvalidDeviceId, fiber.StatusBadRequest)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	err := w.wsService.AddClient(c.Context(), jwtUserData.UserId, jwtUserData.Username, deviceId)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return err
}

// GetSingleChatMessages godoc
//
//	@Summary		Chat Messages
//	@Description	get messages of users chat
//	@Tags			User-Websocket
//	@Param			receiverId		query		integer	false	"receiverId"
//	@Param			date			query		time	false	"date"
//	@Param			skip			query		integer	false	"skip"
//	@Param			limit			query		integer	false	"limit"
//	@Param			messageState	query		integer	false	"messageState"
//	@Param			reverseOrder	query		boolean	false	"reverseOrder"
//	@Success		200				{object}	[]model.MessageDataModel
//	@Failure		400				{object}	response.ResponseErrorModel
//	@Router			/v1/ws/singleChat/messages [get]
func (w *WsHandler) GetSingleChatMessages(c *fiber.Ctx) error {
	var params model.GetSingleMessagesReq
	err := c.QueryParser(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	params.UserId = jwtUserData.UserId
	messages, err := w.wsService.GetSingleChatMessages(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, messages)
}

// GetSingleChatList godoc
//
//	@Summary		Chats List
//	@Description	get list of conversations
//	@Tags			User-Websocket
//	@Param			chatsSkip				query		integer	false	"chatsSkip"
//	@Param			chatsLimit				query		integer	false	"chatsLimit"
//	@Param			messagePerChatSkip		query		integer	false	"messagePerChatSkip"
//	@Param			messagePerChatLimit		query		integer	false	"messagePerChatLimit"
//	@Param			messageState			query		integer	false	"messageState"
//	@Param			includeProfileImages	query		boolean	false	"includeProfileImages"
//	@Success		200						{object}	[]model.ChatsCompressedDataModel
//	@Failure		400						{object}	response.ResponseErrorModel
//	@Router			/v1/ws/singleChat/list [get]
func (w *WsHandler) GetSingleChatList(c *fiber.Ctx) error {
	var params model.GetSingleChatListReq
	err := c.QueryParser(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	params.UserId = jwtUserData.UserId
	messages, err := w.wsService.GetSingleChatList(&params)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, messages)
}
