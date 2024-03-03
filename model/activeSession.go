package model

import "time"

type ActiveSession struct {
	UserId       int64     `gorm:"column:userId;type:integer;not null;index:ActiveSession_userId_refreshToken_idx;primaryKey;"`
	AppName      string    `gorm:"column:appName;type:text;not null;"`
	AppVersion   string    `gorm:"column:appVersion;type:text;not null;"`
	DeviceId     string    `gorm:"column:deviceId;type:text;primaryKey;"`
	DeviceModel  string    `gorm:"column:deviceModel;type:text;not null;"`
	DeviceOs     string    `gorm:"column:deviceOs;type:text;not null;"`
	NotifToken   string    `gorm:"column:notifToken;type:text;not null;default ''::text;"`
	IpLocation   string    `gorm:"column:ipLocation;type:text;not null;"`
	LastUseDate  time.Time `gorm:"column:lastUseDate;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	LoginDate    time.Time `gorm:"column:loginDate;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	RefreshToken string    `gorm:"column:refreshToken;type:text;not null;uniqueIndex:ActiveSession_refreshToken_key;index:ActiveSession_userId_refreshToken_idx;"`
}

func (ActiveSession) TableName() string {
	return "ActiveSession"
}

//------------------------------------------
//------------------------------------------

type ActiveSessionDataModel struct {
	UserId      int64  `gorm:"column:userId;" json:"-"`
	AppName     string `gorm:"column:appName;"`
	AppVersion  string `gorm:"column:appVersion;"`
	DeviceId    string `gorm:"column:deviceId;"`
	DeviceModel string `gorm:"column:deviceModel;"`
	DeviceOs    string `gorm:"column:deviceOs;"`
	//NotifToken   string    `gorm:"column:notifToken;" json:"-"`
	IpLocation   string    `gorm:"column:ipLocation;"`
	LastUseDate  time.Time `gorm:"column:lastUseDate;"`
	LoginDate    time.Time `gorm:"column:loginDate;"`
	RefreshToken string    `gorm:"column:refreshToken;" json:"-"`
}

func (ActiveSessionDataModel) TableName() string {
	return "ActiveSession"
}

type ActiveSessionRes struct {
	ThisDevice     *ActiveSessionDataModel   `json:"thisDevice"`
	ActiveSessions *[]ActiveSessionDataModel `json:"activeSessions"`
}
