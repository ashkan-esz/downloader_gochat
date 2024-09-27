package model

import (
	"strings"
	"time"
)

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
	UserId       int64     `json:"userId" query:"-" swaggerignore:"true"`
	ReceiverId   int64     `json:"receiverId" minimum:"1"`
	Date         time.Time `json:"date"`
	Skip         int       `json:"skip" minimum:"0"`
	Limit        int       `json:"limit" minimum:"1"`
	MessageState int       `json:"messageState" default:"0" minimum:"0" maximum:"2"` // 0: pending, 1: saved, 2: receiver read || on 0 value, this filter won't apply
	ReverseOrder bool      `json:"reverseOrder,omitempty" default:"false"`
}

func (m *GetSingleMessagesReq) Validate() string {
	errors := make([]string, 0)
	if m.ReceiverId < 1 {
		errors = append(errors, "receiverId cannot be smaller than 1")
	}
	if m.Skip < 0 {
		errors = append(errors, "skip cannot be smaller than 0")
	}
	if m.Limit < 1 {
		errors = append(errors, "limit cannot be smaller than 1")
	}
	if m.MessageState < 0 || m.MessageState > 2 {
		errors = append(errors, "messageState must be in range of 0-2")
	}

	return strings.Join(errors, ", ")
}

type GetSingleChatListReq struct {
	UserId               int64 `json:"userId" query:"-" swaggerignore:"true"`
	ChatsSkip            int   `json:"chatsSkip" minimum:"0"`
	ChatsLimit           int   `json:"chatsLimit" minimum:"1"`
	MessagePerChatSkip   int   `json:"messagePerChatSkip" minimum:"0"`
	MessagePerChatLimit  int   `json:"messagePerChatLimit" minimum:"1" maximum:"6"`
	MessageState         int   `json:"messageState" default:"0" minimum:"0" maximum:"2"` // 0: pending, 1: saved, 2: receiver read
	IncludeProfileImages bool  `json:"includeProfileImages" default:"false"`
}

func (m *GetSingleChatListReq) Validate() string {
	errors := make([]string, 0)
	if m.ChatsSkip < 0 {
		errors = append(errors, "chatsSkip cannot be smaller than 0")
	}
	if m.ChatsLimit < 1 {
		errors = append(errors, "chatsLimit cannot be smaller than 1")
	}
	if m.MessagePerChatSkip < 0 {
		errors = append(errors, "messagePerChatSkip cannot be smaller than 0")
	}
	if m.MessagePerChatLimit < 1 || m.MessagePerChatLimit > 6 {
		errors = append(errors, "messagePerChatLimit must be in range of 1-6")
	}
	if m.MessageState < 0 || m.MessageState > 2 {
		errors = append(errors, "messageState must be in range of 0-2")
	}

	return strings.Join(errors, ", ")
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
	UserId       int64     `gorm:"column:userId;" json:"userId"`
	Username     string    `gorm:"column:username;" json:"username"`
	PublicName   string    `gorm:"column:publicName;" json:"publicName"`
	LastSeenDate time.Time `gorm:"column:lastSeenDate;" json:"lastSeenDate"`
	Id           int64     `gorm:"column:id" json:"id"`
	Content      string    `gorm:"column:content" json:"content"`
	Date         time.Time `gorm:"column:date" json:"date"`
	State        int       `gorm:"column:state" json:"state"`
	RoomId       *int64    `gorm:"column:roomId" json:"roomId"`
	CreatorId    int64     `gorm:"column:creatorId" json:"creatorId"`
	ReceiverId   int64     `gorm:"column:receiverId" json:"receiverId"`
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
	LastSeenDate        time.Time               `json:"lastSeenDate"`
	ProfileImages       []ProfileImageDataModel `json:"profileImages"`
	Messages            []MessageDataModel      `json:"messages"`
	UnreadMessagesCount int                     `json:"unreadMessagesCount"`
	IsOnline            bool                    `json:"isOnline"`
}

type MessagesCountDataModel struct {
	Count      int    `gorm:"column:count;" json:"count"`
	CreatorId  int64  `gorm:"column:creatorId;" json:"creatorId"`
	ReceiverId int64  `gorm:"column:receiverId;" json:"receiverId"`
	State      int    `gorm:"column:state;" json:"state"`
	RoomId     *int64 `gorm:"column:roomId;" json:"roomId"`
}
