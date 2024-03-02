package model

import "time"

type ProfileImage struct {
	AddDate      time.Time `gorm:"column:addDate;type:timestamp(3);not null;"`
	Url          string    `gorm:"column:url;type:text;uniqueIndex:ProfileImage_url_key;primaryKey"`
	OriginalSize int64     `gorm:"column:originalSize;type:integer;not null;"`
	Size         int64     `gorm:"column:size;type:integer;not null;"`
	Thumbnail    string    `gorm:"column:thumbnail;type:text;not null;"`
	UserId       int64     `gorm:"column:userId;type:integer;not null;"`
}

func (ProfileImage) TableName() string {
	return "ProfileImage"
}

//------------------------------------------
//------------------------------------------

type ProfileImageDataModel struct {
	AddDate   time.Time `gorm:"column:addDate;" json:"addDate"`
	Url       string    `gorm:"column:url;" json:"url"`
	Size      int64     `gorm:"column:size;" json:"size"`
	Thumbnail string    `gorm:"column:thumbnail;" json:"thumbnail"`
	UserId    int64     `gorm:"column:userId;" json:"userId"`
}

type FollowListProfileImageDataModel struct {
	AddDate   time.Time `gorm:"column:addDate;" json:"-"`
	Url       string    `gorm:"column:url;" json:"url"`
	Size      int64     `gorm:"column:size;" json:"size"`
	Thumbnail string    `gorm:"column:thumbnail;" json:"thumbnail"`
	UserId    int64     `gorm:"column:userId;" json:"-"`
}

func (FollowListProfileImageDataModel) TableName() string {
	return "ProfileImage"
}
