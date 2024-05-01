package model

import "time"

type Notification struct {
	Id              int64           `gorm:"column:id;type:serial;autoIncrement;primaryKey;"`
	CreatorId       int64           `gorm:"column:creatorId;type:integer;not null;"`
	ReceiverId      int64           `gorm:"column:receiverId;type:integer;not null;uniqueIndex:Notification_receiverId_date_idx;"`
	Date            time.Time       `gorm:"column:date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;uniqueIndex:Notification_receiverId_date_idx;"`
	Status          int             `gorm:"column:status;type:integer;default:0;not null;"` //1: saved, 2: seen
	Message         string          `gorm:"column:message;type:text;not null;"`
	EntityId        string          `gorm:"column:entityId;type:text;not null;"`
	EntityTypeId    int             `gorm:"column:entityTypeId;type:integer;not null;"`
	SubEntityTypeId SubEntityTypeId `gorm:"column:subEntityTypeId;type:integer;not null;"`
}

func (Notification) TableName() string {
	return "Notification"
}

type NotificationEntityType struct {
	EntityTypeId  int            `gorm:"column:entityTypeId;type:integer;not null;primaryKey;"`
	EntityType    string         `gorm:"column:entityType;type:text;not null;uniqueIndex:NotificationEntityType_entityType_key;"`
	Notifications []Notification `gorm:"foreignKey:EntityTypeId;references:EntityTypeId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (NotificationEntityType) TableName() string {
	return "NotificationEntityType"
}

var NotificationEntityTypesAndId = []*NotificationEntityType{
	{
		EntityTypeId: 1,
		EntityType:   "User",
	},
	{
		EntityTypeId: 2,
		EntityType:   "Message",
	},
	{
		EntityTypeId: 3,
		EntityType:   "Movie",
	},
}

const (
	FollowNotificationTypeId     = 1
	NewMessageNotificationTypeId = 2
	MoviesNotificationTypeId     = 3
)

type SubEntityTypeId int

const (
	FinishedListSpinOffSequel SubEntityTypeId = 1
	FollowingMovie            SubEntityTypeId = 2
	FollowMovieBetterQuality  SubEntityTypeId = 3
	FollowMovieSubtitle       SubEntityTypeId = 4
	FutureList                SubEntityTypeId = 5
	FutureListSerialSeasonEnd SubEntityTypeId = 6
	FutureListSubtitle        SubEntityTypeId = 7
)

//-----------------------------------
//-----------------------------------

type NotificationDataModel struct {
	Id              int64           `gorm:"column:id;" json:"id"`
	CreatorId       int64           `gorm:"column:creatorId;" json:"creatorId"`
	ReceiverId      int64           `gorm:"column:receiverId;" json:"receiverId"`
	Date            time.Time       `gorm:"column:date;" json:"date"`
	Status          int             `gorm:"column:status;" json:"status"` //1: saved, 2: seen
	Message         string          `gorm:"column:message;" json:"message"`
	EntityId        string          `gorm:"column:entityId;" json:"entityId"`
	EntityTypeId    int             `gorm:"column:entityTypeId;" json:"entityTypeId"`
	SubEntityTypeId SubEntityTypeId `gorm:"column:subEntityTypeId" json:"subEntityTypeId"`
	CreatorImage    string          `json:"creatorImage"`
}
