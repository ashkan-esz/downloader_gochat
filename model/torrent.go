package model

import "time"

type UserTorrent struct {
	UserId         int64     `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:UserTorrent_userId_key;" swaggerignore:"true"`
	TorrentLeachGb int       `gorm:"column:torrentLeachGb;type:integer;not null;"`
	TorrentSearch  int       `gorm:"column:torrentSearch;type:integer;not null;"`
	FirstUseAt     time.Time `gorm:"column:firstUseAt;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
}

func (UserTorrent) TableName() string {
	return "UserTorrent"
}
