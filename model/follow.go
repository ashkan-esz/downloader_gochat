package model

import "time"

type Follow struct {
	// User A follow B
	// followerId follow followingId
	AddDate     time.Time `gorm:"column:addDate;type:timestamp(3);not null;"`
	FollowerId  int64     `gorm:"column:followerId;type:integer;primaryKey"`
	FollowingId int64     `gorm:"column:followingId;type:integer;primaryKey"`
}

func (Follow) TableName() string {
	return "Follow"
}

//------------------------------------------
//------------------------------------------

type FollowUserDataModel struct {
	UserId        int64                             `gorm:"column:userId" json:"userId"`
	Username      string                            `gorm:"column:username" json:"username"`
	RawUsername   string                            `gorm:"column:rawUsername" json:"rawUsername"`
	PublicName    string                            `gorm:"column:publicName" json:"publicName"`
	Bio           string                            `gorm:"column:bio" json:"bio"`
	ProfileImages []FollowListProfileImageDataModel `gorm:"foreignKey:UserId;references:UserId;" json:"profileImages"`
}
