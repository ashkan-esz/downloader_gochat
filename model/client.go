package model

// from client to server
const MessageReadAction = "message-read"
const SendNewMessageAction = "send-new-message"

// from server to client
const ReceiveNewMessageAction = "receive-new-message"
const NewMessageSendResultAction = "new-message-send-result"

// both way
const SingleChatsListAction = "single-chats-list"
const SingleChatMessagesAction = "single-chat-messages"

type ClientMessage struct {
	Action          string               `json:"action,omitempty"`
	NewMessage      NewMessage           `json:"newMessage,omitempty"`
	ChatMessagesReq GetSingleMessagesReq `json:"chatMessagesReq,omitempty"`
	ChatsListReq    GetSingleChatListReq `json:"chatsListReq,omitempty"`
}

type ChannelMessage struct {
	Action               string                      `json:"action,omitempty"`
	ReceiveNewMessage    *ReceiveNewMessage          `json:"receiveNewMessage,omitempty"`
	NewMessageSendResult *NewMessageSendResult       `json:"newMessageSendResult,omitempty"`
	ChatMessagesReq      *GetSingleMessagesReq       `json:"chatMessagesReq,omitempty"`
	ChatsListReq         *GetSingleChatListReq       `json:"chatsListReq,omitempty"`
	ChatMessages         *[]MessageDataModel         `json:"chatMessages,omitempty"`
	Chats                *[]ChatsCompressedDataModel `json:"chats,omitempty"`
}

//------------------------------------------
//------------------------------------------

type NewMessage struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId"`
	ReceiverId int64  `json:"receiverId"`
}

type ReceiveNewMessage struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId"`
	ReceiverId int64  `json:"receiverId"`
	State      int    `json:"state"`
	UserId     int64  `json:"userId"`
	Username   string `json:"username"`
}

type NewMessageSendResult struct {
	RoomId       int64  `json:"roomId"`
	ReceiverId   int64  `json:"receiverId"`
	State        int    `json:"state"`
	Code         int    `json:"code"`
	ErrorMessage string `json:"errorMessage"`
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
	}
}

func CreateNewMessageSendResult(roomId int64, receiverId int64, state int, code int, err string) *ChannelMessage {
	return &ChannelMessage{
		Action: NewMessageSendResultAction,
		NewMessageSendResult: &NewMessageSendResult{
			//Id:       m.RoomId, //todo : need a way to understand which message failed/received on client side
			RoomId:       roomId,
			ReceiverId:   receiverId,
			State:        state,
			Code:         code,
			ErrorMessage: err,
		},
		ChatsListReq:      nil,
		ChatMessages:      nil,
		ChatMessagesReq:   nil,
		ReceiveNewMessage: nil,
		Chats:             nil,
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
	}
}
