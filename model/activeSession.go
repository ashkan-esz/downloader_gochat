package model

import "time"

type ActiveSession struct {
	UserId       int64     `gorm:"column:userId;type:integer;not null;index:ActiveSession_userId_refreshToken_idx;"`
	AppName      string    `gorm:"column:appName;type:text;not null;"`
	AppVersion   string    `gorm:"column:appVersion;type:text;not null;"`
	DeviceId     string    `gorm:"column:deviceId;type:text;uniqueIndex:ActiveSession_deviceId_key;primaryKey;"`
	DeviceModel  string    `gorm:"column:deviceModel;type:text;not null;"`
	DeviceOs     string    `gorm:"column:deviceOs;type:text;not null;"`
	IpLocation   string    `gorm:"column:ipLocation;type:text;not null;"`
	LastUseDate  time.Time `gorm:"column:lastUseDate;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	LoginDate    time.Time `gorm:"column:loginDate;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	RefreshToken string    `gorm:"column:refreshToken;type:text;not null;uniqueIndex:ActiveSession_refreshToken_key;index:ActiveSession_userId_refreshToken_idx;"`
}

func (ActiveSession) TableName() string {
	return "ActiveSession"
}
