package model

type MediaFile struct {
	Id        int64  `gorm:"column:id;type:serial;autoIncrement;primaryKey;" json:"id"`
	MessageId int64  `gorm:"column:messageId;type:integer;not null;uniqueIndex:MediaFile_messageId_idx;" json:"messageId"`
	Url       string `gorm:"column:url;type:text;not null;" json:"url"`
	Type      string `gorm:"column:type;type:text;not null;" json:"type"`
	Size      int64  `gorm:"column:size;type:integer;not null;" json:"size"`
	Thumbnail string `gorm:"column:thumbnail;type:text;not null;" json:"thumbnail"`
	BlurHash  string `gorm:"column:blurHash;type:text;not null;" json:"blurHash"`
}

func (MediaFile) TableName() string {
	return "MediaFile"
}

//---------------------------------------
//---------------------------------------

type UploadMediaReq struct {
	Content    string `json:"content"`
	RoomId     int64  `json:"roomId"`
	ReceiverId int64  `json:"receiverId"`
	Uuid       string `json:"uuid"`
}
