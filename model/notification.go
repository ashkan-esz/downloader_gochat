package model

import "time"

type Notification struct {
	Id           int64     `gorm:"column:id;type:serial;autoIncrement;primaryKey;"`
	CreatorId    int64     `gorm:"column:creatorId;type:integer;not null;"`
	ReceiverId   int64     `gorm:"column:receiverId;type:integer;not null;uniqueIndex:Notification_receiverId_date_idx;"`
	Date         time.Time `gorm:"column:date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;uniqueIndex:Notification_receiverId_date_idx;"`
	Status       int       `gorm:"column:status;type:integer;default:0;not null;"`
	EntityId     int64     `gorm:"column:entityId;type:integer;not null;"`
	EntityTypeId int       `gorm:"column:entityTypeId;type:integer;not null;"`
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
}

const (
	FollowNotificationTypeId     = 1
	NewMessageNotificationTypeId = 2
)

//-----------------------------------
//-----------------------------------

type NotificationDataModel struct {
	Id           int64     `gorm:"column:id;" json:"id"`
	CreatorId    int64     `gorm:"column:creatorId;" json:"creatorId"`
	ReceiverId   int64     `gorm:"column:receiverId;" json:"receiverId"`
	Date         time.Time `gorm:"column:date;" json:"date"`
	Status       int       `gorm:"column:status;" json:"status"`
	EntityId     int64     `gorm:"column:entityId;" json:"entityId"`
	EntityTypeId int       `gorm:"column:entityTypeId;" json:"entityTypeId"`
	Message      string    `json:"message"`
}
