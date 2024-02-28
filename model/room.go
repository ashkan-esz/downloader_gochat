package model

import "time"

type Room struct {
	RoomId     int64     `gorm:"column:roomId;type:serial;autoIncrement;primaryKey;"`
	CreatorId  int64     `gorm:"column:creatorId;type:integer;not null;uniqueIndex:Room_creatorId_receiverId_key;"`
	ReceiverId int64     `gorm:"column:receiverId;type:integer;not null;uniqueIndex:Room_creatorId_receiverId_key;"`
	Messages   []Message `gorm:"foreignKey:RoomId;references:RoomId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Room) TableName() string {
	return "Room"
}

type Message struct {
	Id         int64     `gorm:"column:id;type:serial;autoIncrement;primaryKey;"`
	Content    string    `gorm:"column:content;type:text;not null;"`
	Date       time.Time `gorm:"column:date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;uniqueIndex:Message_date_state_idx;"`
	State      int       `gorm:"column:state;type:integer;default:0;not null;uniqueIndex:Message_date_state_idx;"`
	RoomId     *int64    `gorm:"column:roomId;type:integer;"`
	CreatorId  int64     `gorm:"column:creatorId;type:integer;not null;"`
	ReceiverId int64     `gorm:"column:receiverId;type:integer;not null;"`
	//-----------------------------------
	Medias []MediaFile `gorm:"foreignKey:MessageId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Message) TableName() string {
	return "Message"
}

type UserMessageRead struct {
	UserId              int64     `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:UserMessageRead_userId_key;"`
	LastTimeRead        time.Time `gorm:"column:lastTimeRead;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	LastMessageReceived time.Time `gorm:"column:lastMessageReceived;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
}

func (UserMessageRead) TableName() string {
	return "UserMessageRead"
}

//---------------------------------------
//---------------------------------------

type GetSingleMessagesReq struct {
	UserId       int64     `json:"userId"`
	ReceiverId   int64     `json:"receiverId"`
	Date         time.Time `json:"date"`
	Skip         int       `json:"skip"`
	Limit        int       `json:"limit"`
	MessageState int       `json:"messageState"`
	ReverseOrder bool      `json:"reverseOrder,omitempty"`
}

type GetSingleChatListReq struct {
	UserId               int64 `json:"userId"`
	ChatsSkip            int   `json:"chatsSkip"`
	ChatsLimit           int   `json:"chatsLimit"`
	MessagePerChatSkip   int   `json:"messagePerChatSkip"`
	MessagePerChatLimit  int   `json:"messagePerChatLimit"`
	MessageState         int   `json:"messageState"`
	IncludeProfileImages bool  `json:"includeProfileImages"`
}

type MessageDataModel struct {
	Id         int64       `gorm:"column:id" json:"id"`
	Content    string      `gorm:"column:content" json:"content"`
	Date       time.Time   `gorm:"column:date" json:"date"`
	State      int         `gorm:"column:state" json:"state"`
	RoomId     *int64      `gorm:"column:roomId" json:"roomId,omitempty"`
	CreatorId  int64       `gorm:"column:creatorId" json:"creatorId"`
	ReceiverId int64       `gorm:"column:receiverId" json:"receiverId"`
	Medias     []MediaFile `gorm:"foreignKey:MessageId;references:Id;" json:"medias"`
}

type ChatsDataModel struct {
	UserId     int64     `gorm:"column:userId;" json:"userId"`
	Username   string    `gorm:"column:username;" json:"username"`
	PublicName string    `gorm:"column:publicName;" json:"publicName"`
	Role       string    `gorm:"column:role;" json:"role"`
	Id         int64     `gorm:"column:id" json:"id"`
	Content    string    `gorm:"column:content" json:"content"`
	Date       time.Time `gorm:"column:date" json:"date"`
	State      int       `gorm:"column:state" json:"state"`
	RoomId     *int64    `gorm:"column:roomId" json:"roomId"`
	CreatorId  int64     `gorm:"column:creatorId" json:"creatorId"`
	ReceiverId int64     `gorm:"column:receiverId" json:"receiverId"`
	//Medias     []MediaFile `gorm:"foreignKey:MessageId;references:Id;" json:"medias"`
	MediaFileId   int64     `gorm:"column:id;" json:"mediaFileId"`
	MediaFileDate time.Time `gorm:"column:date;" json:"mediaFileDate"`
	Url           string    `gorm:"column:url;" json:"url"`
	Type          string    `gorm:"column:type;" json:"type"`
	Size          int64     `gorm:"column:size;" json:"size"`
	Thumbnail     string    `gorm:"column:thumbnail;" json:"thumbnail"`
	BlurHash      string    `gorm:"column:blurHash;" json:"blurHash"`
}

type ChatsCompressedDataModel struct {
	UserId              int64                   `json:"userId"`
	Username            string                  `json:"username"`
	PublicName          string                  `json:"publicName"`
	Role                string                  `json:"role"`
	ProfileImages       []ProfileImageDataModel `json:"profileImages"`
	Messages            []MessageDataModel      `json:"messages"`
	UnreadMessagesCount int                     `json:"unreadMessagesCount"`
}

type MessagesCountDataModel struct {
	Count      int    `gorm:"column:count;" json:"count"`
	CreatorId  int64  `gorm:"column:creatorId;" json:"creatorId"`
	ReceiverId int64  `gorm:"column:receiverId;" json:"receiverId"`
	State      int    `gorm:"column:state;" json:"state"`
	RoomId     *int64 `gorm:"column:roomId;" json:"roomId"`
}
