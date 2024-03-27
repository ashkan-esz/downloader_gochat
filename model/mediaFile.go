package model

import (
	"strings"
	"time"
)

type MediaFile struct {
	Id        int64     `gorm:"column:id;type:serial;autoIncrement;primaryKey;" json:"id"`
	MessageId int64     `gorm:"column:messageId;type:integer;not null;uniqueIndex:MediaFile_messageId_idx;" json:"messageId"`
	Date      time.Time `gorm:"column:date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;" json:"date"`
	Url       string    `gorm:"column:url;type:text;not null;" json:"url"`
	Type      string    `gorm:"column:type;type:text;not null;" json:"type"`
	Size      int64     `gorm:"column:size;type:integer;not null;" json:"size"`
	Thumbnail string    `gorm:"column:thumbnail;type:text;not null;" json:"thumbnail"`
	BlurHash  string    `gorm:"column:blurHash;type:text;not null;" json:"blurHash"`
}

func (MediaFile) TableName() string {
	return "MediaFile"
}

//---------------------------------------
//---------------------------------------

type UploadMediaReq struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId" minimum:"-1"` // value -1 means its user-to-user message
	ReceiverId int64  `json:"receiverId" minimum:"1"`
	Uuid       string `json:"uuid"`
}

func (m *UploadMediaReq) Validate() string {
	errors := make([]string, 0)

	if m.RoomId < -1 {
		errors = append(errors, "roomId cannot be smaller than -1")
	}
	if m.ReceiverId < 1 {
		errors = append(errors, "receiverId cannot be smaller than 1")
	}

	return strings.Join(errors, ", ")
}
