package model

type CachedUserData struct {
	UserId               int64                `gorm:"column:userId" json:"userId"`
	Username             string               `gorm:"column:username" json:"username"`
	PublicName           string               `gorm:"column:publicName" json:"publicName"`
	ProfileImages        []ProfileImage       `gorm:"foreignKey:UserId;references:UserId;" json:"profileImages"`
	NotificationSettings NotificationSettings `gorm:"foreignKey:UserId;references:UserId;"`
	NotifTokens          []string             `json:"notifTokens"`
}
