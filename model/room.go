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
	Date       time.Time `gorm:"column:date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	State      int       `gorm:"column:state;type:integer;default:0;not null;"`
	RoomId     *int64    `gorm:"column:roomId;type:integer;"`
	CreatorId  int64     `gorm:"column:creatorId;type:integer;not null;"`
	ReceiverId int64     `gorm:"column:receiverId;type:integer;not null;"`
}

func (Message) TableName() string {
	return "Message"
}

//---------------------------------------
//---------------------------------------

type ClientMessage struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId"`
	ReceiverId int64  `json:"receiverId"`
}

type ChannelMessage struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId"`
	ReceiverId int64  `json:"receiverId"`
	State      int    `json:"state"`
	UserId     int64  `json:"userId"`
	Username   string `json:"username"`
}

//---------------------------------------
//---------------------------------------

type CreateRoomReq struct {
	SenderId   int64 `json:"senderId"`
	ReceiverId int64 `json:"receiverId"`
}

type CreateRoomRes struct {
	RoomId int64 `json:"roomId"`
}

type RoomRes struct {
	ID int64 `json:"id"`
}

type ClientRes struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}
