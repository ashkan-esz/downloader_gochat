package model

type DownloadLinksSettings struct {
	UserId             int64    `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:DownloadLinksSettings_userId_key"`
	IncludeCensored    bool     `gorm:"column:includeCensored;type:boolean;default:true;not null;"`
	IncludeDubbed      bool     `gorm:"column:includeDubbed;type:boolean;default:true;not null;"`
	IncludeHardSub     bool     `gorm:"column:includeHardSub;type:boolean;default:true;not null;"`
	PreferredQualities []string `gorm:"column:preferredQualities;type:text[];default:ARRAY ['720p'::text, '1080p'::text, '2160p'::text];not null;"`
}

func (DownloadLinksSettings) TableName() string {
	return "DownloadLinksSettings"
}

//-----------------------------------------------------
//-----------------------------------------------------

type NotificationSettings struct {
	UserId                    int64 `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:NotificationSettings_userId_key"`
	FinishedListSpinOffSequel bool  `gorm:"column:finishedList_spinOffSequel;type:boolean;default:true;not null;"`
	FollowMovie               bool  `gorm:"column:followMovie;type:boolean;default:true;not null;"`
	FollowMovieBetterQuality  bool  `gorm:"column:followMovie_betterQuality;type:boolean;default:true;not null;"`
	FollowMovieSubtitle       bool  `gorm:"column:followMovie_subtitle;type:boolean;default:true;not null;"`
	FutureList                bool  `gorm:"column:futureList;type:boolean;default:true;not null;"`
	FutureListSerialSeasonEnd bool  `gorm:"column:futureList_serialSeasonEnd;type:boolean;default:true;not null;"`
	FutureListSubtitle        bool  `gorm:"column:futureList_subtitle;type:boolean;default:true;not null;"`
}

func (NotificationSettings) TableName() string {
	return "NotificationSettings"
}

//-----------------------------------------------------
//-----------------------------------------------------

type MovieSettings struct {
	UserId        int64 `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:MovieSettings_userId_key"`
	IncludeAnime  bool  `gorm:"column:includeAnime;type:boolean;default:true;not null;"`
	IncludeHentai bool  `gorm:"column:includeHentai;type:boolean;default:false;not null;"`
}

func (MovieSettings) TableName() string {
	return "MovieSettings"
}
