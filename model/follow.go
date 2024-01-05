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
