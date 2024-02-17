package model

import (
	"time"
)

// from client to server
const MessageReadAction = "message-read"
const SendNewMessageAction = "send-new-message"

// from server to client
const ReceiveNewMessageAction = "receive-new-message"
const NewMessageSendResultAction = "new-message-send-result"
const ReceiveMessageStateAction = "receive-message-state"
const ErrorAction = "action-error"
const FollowNotifAction = "new-follow-notification"
const NewMessageNotifAction = "new-message-notification"

// both way
const SingleChatsListAction = "single-chats-list"
const SingleChatMessagesAction = "single-chat-messages"

type ClientMessage struct {
	Action          string               `json:"action,omitempty"`
	NewMessage      NewMessage           `json:"newMessage,omitempty"`
	MessageRead     *MessageRead         `json:"messageRead,omitempty"`
	ChatMessagesReq GetSingleMessagesReq `json:"chatMessagesReq,omitempty"`
	ChatsListReq    GetSingleChatListReq `json:"chatsListReq,omitempty"`
}

type ChannelMessage struct {
	Action               string                      `json:"action,omitempty"`
	ReceiveNewMessage    *ReceiveNewMessage          `json:"receiveNewMessage,omitempty"`
	NewMessageSendResult *NewMessageSendResult       `json:"newMessageSendResult,omitempty"`
	MessageRead          *MessageRead                `json:"messageRead,omitempty"`
	ChatMessagesReq      *GetSingleMessagesReq       `json:"chatMessagesReq,omitempty"`
	ChatsListReq         *GetSingleChatListReq       `json:"chatsListReq,omitempty"`
	ChatMessages         *[]MessageDataModel         `json:"chatMessages,omitempty"`
	Chats                *[]ChatsCompressedDataModel `json:"chats,omitempty"`
	ActionError          *ActionError                `json:"actionError,omitempty"`
	NotificationData     *NotificationDataModel      `json:"notificationData,omitempty"`
}

//------------------------------------------
//------------------------------------------

type NewMessage struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId"`
	ReceiverId int64  `json:"receiverId"`
	Uuid       string `json:"uuid"`
}

type ReceiveNewMessage struct {
	Id         int64     `json:"id"`
	Uuid       string    `json:"uuid"`
	Content    string    `json:"content"`
	RoomId     int64     `json:"roomId"`
	ReceiverId int64     `json:"receiverId"`
	State      int       `json:"state"`
	Date       time.Time `json:"date"`
	UserId     int64     `json:"userId"`
	Username   string    `json:"username"`
}

type NewMessageSendResult struct {
	Id           int64     `json:"id"`
	Uuid         string    `json:"uuid"`
	RoomId       int64     `json:"roomId"`
	ReceiverId   int64     `json:"receiverId"`
	State        int       `json:"state"`
	Date         time.Time `json:"date"`
	Code         int       `json:"code"`
	ErrorMessage string    `json:"errorMessage"`
}

type MessageRead struct {
	Id         int64     `json:"id"`
	RoomId     int64     `json:"roomId"`
	UserId     int64     `json:"userId"`
	ReceiverId int64     `json:"receiverId"`
	State      int       `json:"state"`
	Date       time.Time `json:"date"`
}

type ActionError struct {
	Action       string      `json:"action"`
	ActionData   interface{} `json:"actionData"`
	Code         int         `json:"code"`
	ErrorMessage string      `json:"errorMessage"`
}

//------------------------------------------
//------------------------------------------

func CreateReceiveNewMessageAction(message *ReceiveNewMessage) *ChannelMessage {
	return &ChannelMessage{
		Action:               ReceiveNewMessageAction,
		ReceiveNewMessage:    message,
		ChatsListReq:         nil,
		ChatMessages:         nil,
		ChatMessagesReq:      nil,
		NewMessageSendResult: nil,
		Chats:                nil,
		MessageRead:          nil,
	}
}

func CreateNewMessageSendResult(id int64, uuid string, roomId int64, receiverId int64, date time.Time, state int, code int, err string) *ChannelMessage {
	return &ChannelMessage{
		Action: NewMessageSendResultAction,
		NewMessageSendResult: &NewMessageSendResult{
			Id:           id,
			Uuid:         uuid,
			RoomId:       roomId,
			ReceiverId:   receiverId,
			Date:         date,
			State:        state,
			Code:         code,
			ErrorMessage: err,
		},
		ChatsListReq:      nil,
		ChatMessages:      nil,
		ChatMessagesReq:   nil,
		ReceiveNewMessage: nil,
		Chats:             nil,
		MessageRead:       nil,
	}
}

func CreateGetChatMessagesAction(params *GetSingleMessagesReq) *ChannelMessage {
	return &ChannelMessage{
		Action:               SingleChatMessagesAction,
		ChatMessagesReq:      params,
		ChatsListReq:         nil,
		ChatMessages:         nil,
		ReceiveNewMessage:    nil,
		NewMessageSendResult: nil,
		Chats:                nil,
		MessageRead:          nil,
	}
}

func CreateReturnChatMessagesAction(messages *[]MessageDataModel) *ChannelMessage {
	return &ChannelMessage{
		Action:               SingleChatMessagesAction,
		ChatMessages:         messages,
		ChatsListReq:         nil,
		ChatMessagesReq:      nil,
		ReceiveNewMessage:    nil,
		NewMessageSendResult: nil,
		Chats:                nil,
		MessageRead:          nil,
	}
}

func CreateGetChatListAction(params *GetSingleChatListReq) *ChannelMessage {
	return &ChannelMessage{
		Action:               SingleChatsListAction,
		ChatsListReq:         params,
		ChatMessages:         nil,
		ChatMessagesReq:      nil,
		ReceiveNewMessage:    nil,
		NewMessageSendResult: nil,
		Chats:                nil,
		MessageRead:          nil,
	}
}

func CreateReturnChatListAction(messages *[]ChatsCompressedDataModel) *ChannelMessage {
	return &ChannelMessage{
		Action:               SingleChatsListAction,
		Chats:                messages,
		ChatsListReq:         nil,
		ChatMessages:         nil,
		ChatMessagesReq:      nil,
		ReceiveNewMessage:    nil,
		NewMessageSendResult: nil,
		MessageRead:          nil,
	}
}

func CreateMessageReadAction(id int64, roomId int64, userId int64, receiverId int64, date time.Time, state int, isServerToClientAction bool) *ChannelMessage {
	action := MessageReadAction
	if isServerToClientAction {
		action = ReceiveMessageStateAction
	}
	return &ChannelMessage{
		Action: action,
		MessageRead: &MessageRead{
			Id:         id,
			RoomId:     roomId,
			UserId:     userId,
			ReceiverId: receiverId,
			State:      state,
			Date:       date,
		},
		ChatsListReq:         nil,
		ChatMessages:         nil,
		ChatMessagesReq:      nil,
		ReceiveNewMessage:    nil,
		NewMessageSendResult: nil,
		Chats:                nil,
	}
}

func CreateActionError(code int, errorMessage string, action string, actionData interface{}) *ChannelMessage {
	return &ChannelMessage{
		Action: ErrorAction,
		ActionError: &ActionError{
			Action:       action,
			ActionData:   actionData,
			Code:         code,
			ErrorMessage: errorMessage,
		},
		ChatsListReq:         nil,
		ChatMessages:         nil,
		ChatMessagesReq:      nil,
		ReceiveNewMessage:    nil,
		NewMessageSendResult: nil,
		Chats:                nil,
		MessageRead:          nil,
	}
}

func CreateFollowNotificationAction(userId int64, followId int64) *ChannelMessage {
	return &ChannelMessage{
		Action: FollowNotifAction,
		NotificationData: &NotificationDataModel{
			Id:           0,
			CreatorId:    userId,
			ReceiverId:   followId,
			Date:         time.Now(),
			Status:       1,
			EntityId:     userId,
			EntityTypeId: FollowNotificationTypeId,
			Message:      "",
		},
		ReceiveNewMessage:    nil,
		ChatsListReq:         nil,
		ChatMessages:         nil,
		ChatMessagesReq:      nil,
		NewMessageSendResult: nil,
		Chats:                nil,
		MessageRead:          nil,
		ActionError:          nil,
	}
}

func CreateNewMessageNotificationAction(message *ReceiveNewMessage) *ChannelMessage {
	//todo : implement
	return &ChannelMessage{
		Action:               NewMessageNotifAction,
		ReceiveNewMessage:    message,
		ChatsListReq:         nil,
		ChatMessages:         nil,
		ChatMessagesReq:      nil,
		NewMessageSendResult: nil,
		Chats:                nil,
		MessageRead:          nil,
		ActionError:          nil,
	}
}
