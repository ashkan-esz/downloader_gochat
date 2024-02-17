package model

import "time"

type Notification struct {
	Id           int64     `gorm:"column:id;type:serial;autoIncrement;primaryKey;"`
	CreatorId    int64     `gorm:"column:creatorId;type:integer;not null;"`
	ReceiverId   int64     `gorm:"column:receiverId;type:integer;not null;"`
	Date         time.Time `gorm:"column:date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
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

//-----------------------------------
//-----------------------------------

type NotificationDataModel struct {
	//todo : implement
}
