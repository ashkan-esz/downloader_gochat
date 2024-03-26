package model

import (
	"strings"
	"time"
)

type ActionType string

// from client to server
const MessageReadAction ActionType = "message-read"
const SendNewMessageAction ActionType = "send-new-message"

// from server to client
const ReceiveNewMessageAction ActionType = "receive-new-message"
const NewMessageSendResultAction ActionType = "new-message-send-result"
const ReceiveMessageStateAction ActionType = "receive-message-state"
const ErrorAction ActionType = "action-error"
const FollowNotifAction ActionType = "new-follow-notification"
const NewMessageNotifAction ActionType = "new-message-notification"
const UpdateProfileImagesAction ActionType = "update-profile-images"
const UpdateProfileAction ActionType = "update-profile"

// both way
const SingleChatsListAction ActionType = "single-chats-list"
const SingleChatMessagesAction ActionType = "single-chat-messages"
const NotificationSettingsAction ActionType = "notification-settings"
const UserStatusAction ActionType = "user-status"
const UserIsTypingAction ActionType = "user-status-isTyping"

type UserStatusResultType string

const (
	UserStatusOnlineUsers UserStatusResultType = "onlineUsers"
	UserStatusIsTyping    UserStatusResultType = "isTyping"
	UserStatusStopTyping  UserStatusResultType = "stopTyping"
)

type ClientMessage struct {
	Action          ActionType           `json:"action,omitempty"`
	NewMessage      NewMessage           `json:"newMessage,omitempty"`      //action is SendNewMessageAction
	MessageRead     *MessageRead         `json:"messageRead,omitempty"`     //action is MessageReadAction
	ChatMessagesReq GetSingleMessagesReq `json:"chatMessagesReq,omitempty"` //action is SingleChatMessagesAction
	ChatsListReq    GetSingleChatListReq `json:"chatsListReq,omitempty"`    //action is SingleChatsListAction
	UserStatusReq   *UserStatusReq       `json:"userStatusReq,omitempty"`   //action is UserStatusAction
}

type ChannelMessage struct {
	Action               ActionType                  `json:"action,omitempty"`
	ReceiveNewMessage    *ReceiveNewMessage          `json:"receiveNewMessage,omitempty"`
	NewMessageSendResult *NewMessageSendResult       `json:"newMessageSendResult,omitempty"`
	MessageRead          *MessageRead                `json:"messageRead,omitempty"`
	ChatMessagesReq      *GetSingleMessagesReq       `json:"chatMessagesReq,omitempty"`
	ChatsListReq         *GetSingleChatListReq       `json:"chatsListReq,omitempty"`
	ChatMessages         *[]MessageDataModel         `json:"chatMessages,omitempty"`
	Chats                *[]ChatsCompressedDataModel `json:"chats,omitempty"`
	ActionError          *ActionError                `json:"actionError,omitempty"`
	NotificationData     *NotificationDataModel      `json:"notificationData,omitempty"`
	NotificationSettings *NotificationSettings       `json:"notificationSettings,omitempty"`
	ProfileImages        *[]ProfileImageDataModel    `json:"profileImages,omitempty"`
	EditProfile          *EditProfileReq             `json:"editProfile,omitempty"`
	UserStatusReq        *UserStatusReq              `json:"userStatusReq,omitempty"`
	UserStatusRes        *UserStatusRes              `json:"userStatusRes,omitempty"`
}

// for documentation usage
type ServerResultMessage struct {
	Action               ActionType                  `json:"action,omitempty"`
	ReceiveNewMessage    *ReceiveNewMessage          `json:"receiveNewMessage,omitempty"`    //action is ReceiveNewMessageAction
	NewMessageSendResult *NewMessageSendResult       `json:"newMessageSendResult,omitempty"` //action is NewMessageSendResultAction
	MessageRead          *MessageRead                `json:"messageRead,omitempty"`          //action is ReceiveMessageStateAction
	ChatMessages         *[]MessageDataModel         `json:"chatMessages,omitempty"`         //action is SingleChatMessagesAction
	Chats                *[]ChatsCompressedDataModel `json:"chats,omitempty"`                //action is SingleChatsListAction
	ActionError          *ActionError                `json:"actionError,omitempty"`          //action is ErrorAction
	NotificationSettings *NotificationSettings       `json:"notificationSettings,omitempty"` //action is NotificationSettingsAction
	ProfileImages        *[]ProfileImageDataModel    `json:"profileImages,omitempty"`        //action is UpdateProfileImagesAction
	EditProfile          *EditProfileReq             `json:"editProfile,omitempty"`          //action is UpdateProfileAction
	UserStatusRes        *UserStatusRes              `json:"userStatusRes,omitempty"`        //action is UserStatusAction
}

//------------------------------------------
//------------------------------------------

type NewMessage struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId" minimum:"-1"` // value -1 means its user-to-user message
	ReceiverId int64  `json:"receiverId" minimum:"1"`
	Uuid       string `json:"uuid"`
}

func (m *NewMessage) Validate() string {
	errors := make([]string, 0)
	if m.RoomId < -1 {
		errors = append(errors, "roomId cannot be smaller than -1")
	}
	if m.ReceiverId < 1 {
		errors = append(errors, "receiverId cannot be smaller than 1")
	}

	return strings.Join(errors, ", ")
}

type ReceiveNewMessage struct {
	Id           int64       `json:"id"`
	Uuid         string      `json:"uuid"`
	Content      string      `json:"content"`
	RoomId       int64       `json:"roomId"`
	ReceiverId   int64       `json:"receiverId"`
	State        int         `json:"state"`
	Date         time.Time   `json:"date"`
	UserId       int64       `json:"userId"`
	Username     string      `json:"username"`
	CreatorImage string      `json:"creatorImage"`
	Medias       []MediaFile `json:"medias"`
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
	Id         int64     `json:"id" minimum:"1"`
	RoomId     int64     `json:"roomId" minimum:"-1"` // value -1 means its user-to-user message
	UserId     int64     `json:"userId" minimum:"1"`
	ReceiverId int64     `json:"receiverId" swaggerignore:"true"`
	State      int       `json:"state" minimum:"0" maximum:"2"` // 0: pending, 1: saved, 2: receiver read
	Date       time.Time `json:"date"`
}

func (m *MessageRead) Validate() string {
	errors := make([]string, 0)
	if m.Id < 1 {
		errors = append(errors, "id cannot be smaller than 1")
	}
	if m.RoomId < -1 {
		errors = append(errors, "roomId cannot be smaller than -1")
	}
	if m.UserId < 1 {
		errors = append(errors, "userId cannot be smaller than 1")
	}
	if m.State < 0 || m.State > 2 {
		errors = append(errors, "state must be in range of 0-2")
	}

	return strings.Join(errors, ", ")
}

type UserStatusReq struct {
	Type    UserStatusResultType `json:"type"`
	UserId  int64                `json:"userId" swaggerignore:"true"`
	UserIds []int64              `json:"userIds" maximum:"12"`
}

func (m *UserStatusReq) Validate() string {
	errors := make([]string, 0)
	if len(m.UserIds) > 12 {
		errors = append(errors, "userIds length cannot be more than 12")
	}

	return strings.Join(errors, ", ")
}

type UserStatusRes struct {
	Type            UserStatusResultType `json:"type"`
	OnlineUserIds   []int64              `json:"onlineUserIds"`
	IsTypingUserIds []int64              `json:"isTypingUserIds"`
}

type ActionError struct {
	Action       ActionType  `json:"action"`
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

func CreateActionError(code int, errorMessage string, action ActionType, actionData interface{}) *ChannelMessage {
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
	return &ChannelMessage{
		Action: NewMessageNotifAction,
		NotificationData: &NotificationDataModel{
			Id:           0,
			CreatorId:    message.UserId,
			ReceiverId:   message.ReceiverId,
			Date:         message.Date,
			Status:       1,
			EntityId:     message.Id,
			EntityTypeId: NewMessageNotificationTypeId,
			Message:      message.Content,
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

func CreateNotificationSettingsAction(notificationSettings *NotificationSettings) *ChannelMessage {
	return &ChannelMessage{
		Action:               NotificationSettingsAction,
		NotificationSettings: notificationSettings,
		NotificationData:     nil,
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

func CreateUpdateProfileImagesAction(profileImages *[]ProfileImageDataModel) *ChannelMessage {
	return &ChannelMessage{
		Action:               UpdateProfileImagesAction,
		ProfileImages:        profileImages,
		NotificationSettings: nil,
		NotificationData:     nil,
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

func CreateUpdateProfileAction(editProfile *EditProfileReq) *ChannelMessage {
	return &ChannelMessage{
		Action:               UpdateProfileAction,
		EditProfile:          editProfile,
		ProfileImages:        nil,
		NotificationSettings: nil,
		NotificationData:     nil,
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

func CreateGetUserStatusAction(userStatusReq *UserStatusReq) *ChannelMessage {
	return &ChannelMessage{
		Action:               UserStatusAction,
		UserStatusReq:        userStatusReq,
		UserStatusRes:        nil,
		EditProfile:          nil,
		ProfileImages:        nil,
		NotificationSettings: nil,
		NotificationData:     nil,
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

func CreateSendUserStatusAction(userStatusRes *UserStatusRes) *ChannelMessage {
	return &ChannelMessage{
		Action:               UserStatusAction,
		UserStatusRes:        userStatusRes,
		UserStatusReq:        nil,
		EditProfile:          nil,
		ProfileImages:        nil,
		NotificationSettings: nil,
		NotificationData:     nil,
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

func CreateSendUserIsTypingAction(userStatusReq *UserStatusReq) *ChannelMessage {
	return &ChannelMessage{
		Action:               UserIsTypingAction,
		UserStatusReq:        userStatusReq,
		UserStatusRes:        nil,
		EditProfile:          nil,
		ProfileImages:        nil,
		NotificationSettings: nil,
		NotificationData:     nil,
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
