package model

import "github.com/lib/pq"

type DownloadLinksSettings struct {
	UserId             int64          `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:DownloadLinksSettings_userId_key;"`
	IncludeCensored    bool           `gorm:"column:includeCensored;type:boolean;not null;"`
	IncludeDubbed      bool           `gorm:"column:includeDubbed;type:boolean;not null;"`
	IncludeHardSub     bool           `gorm:"column:includeHardSub;type:boolean;not null;"`
	PreferredQualities pq.StringArray `gorm:"column:preferredQualities;type:text[];not null;"`
}

func (DownloadLinksSettings) TableName() string {
	return "DownloadLinksSettings"
}

//-----------------------------------------------------
//-----------------------------------------------------

type NotificationSettings struct {
	UserId                    int64 `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:NotificationSettings_userId_key;"`
	FinishedListSpinOffSequel bool  `gorm:"column:finishedList_spinOffSequel;type:boolean;not null;"`
	FollowMovie               bool  `gorm:"column:followMovie;type:boolean;not null;"`
	FollowMovieBetterQuality  bool  `gorm:"column:followMovie_betterQuality;type:boolean;not null;"`
	FollowMovieSubtitle       bool  `gorm:"column:followMovie_subtitle;type:boolean;not null;"`
	FutureList                bool  `gorm:"column:futureList;type:boolean;not null;"`
	FutureListSerialSeasonEnd bool  `gorm:"column:futureList_serialSeasonEnd;type:boolean;not null;"`
	FutureListSubtitle        bool  `gorm:"column:futureList_subtitle;type:boolean;not null;"`
}

func (NotificationSettings) TableName() string {
	return "NotificationSettings"
}

//-----------------------------------------------------
//-----------------------------------------------------

type MovieSettings struct {
	UserId        int64 `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:MovieSettings_userId_key;"`
	IncludeAnime  bool  `gorm:"column:includeAnime;type:boolean;not null;"`
	IncludeHentai bool  `gorm:"column:includeHentai;type:boolean;not null;"`
}

func (MovieSettings) TableName() string {
	return "MovieSettings"
}
