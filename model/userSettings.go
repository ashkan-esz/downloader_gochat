package model

import (
	"fmt"
	"slices"
	"strings"

	"github.com/lib/pq"
)

type DownloadLinksSettings struct {
	UserId             int64          `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:DownloadLinksSettings_userId_key;" swaggerignore:"true"`
	IncludeCensored    bool           `gorm:"column:includeCensored;type:boolean;not null;" default:"true"`
	IncludeDubbed      bool           `gorm:"column:includeDubbed;type:boolean;not null;" default:"true"`
	IncludeHardSub     bool           `gorm:"column:includeHardSub;type:boolean;not null;" default:"true"`
	PreferredQualities pq.StringArray `gorm:"column:preferredQualities;type:text[];not null;" swaggertype:"array,string"` //enum: 480p,720p,1080p,2160p
}

func (DownloadLinksSettings) TableName() string {
	return "DownloadLinksSettings"
}

func (s *DownloadLinksSettings) Validate() string {
	errors := make([]string, 0)

	for _, quality := range s.PreferredQualities {
		if !slices.Contains(Qualities, Quality(quality)) {
			errors = append(errors, fmt.Sprintf("preferredQualities has invalid value : (%v)", strings.Join(Qualities_string, ",")))
			break
		}
	}

	return strings.Join(errors, ", ")
}

type Quality string

var Qualities = []Quality{Quality_480, Quality_720, Quality_1080, Quality_2160}
var Qualities_string = []string{string(Quality_480), string(Quality_720), string(Quality_1080), string(Quality_2160)}

const (
	Quality_480  Quality = "480p"
	Quality_720  Quality = "720p"
	Quality_1080 Quality = "1080p"
	Quality_2160 Quality = "2160p"
)

//-----------------------------------------------------
//-----------------------------------------------------

type NotificationSettings struct {
	UserId                    int64 `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:NotificationSettings_userId_key;" swaggerignore:"true"`
	NewFollower               bool  `gorm:"column:newFollower;type:boolean;not null;" default:"true"`
	NewMessage                bool  `gorm:"column:newMessage;type:boolean;not null;" default:"false"`
	FinishedListSpinOffSequel bool  `gorm:"column:finishedList_spinOffSequel;type:boolean;not null;" default:"true"`
	FollowMovie               bool  `gorm:"column:followMovie;type:boolean;not null;" default:"true"`
	FollowMovieBetterQuality  bool  `gorm:"column:followMovie_betterQuality;type:boolean;not null;" default:"false"`
	FollowMovieSubtitle       bool  `gorm:"column:followMovie_subtitle;type:boolean;not null;" default:"false"`
	FutureList                bool  `gorm:"column:futureList;type:boolean;not null;" default:"false"`
	FutureListSerialSeasonEnd bool  `gorm:"column:futureList_serialSeasonEnd;type:boolean;not null;" default:"true"`
	FutureListSubtitle        bool  `gorm:"column:futureList_subtitle;type:boolean;not null;" default:"false"`
}

func (NotificationSettings) TableName() string {
	return "NotificationSettings"
}

func (s *NotificationSettings) Validate() string {
	return ""
}

//-----------------------------------------------------
//-----------------------------------------------------

type MovieSettings struct {
	UserId        int64 `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:MovieSettings_userId_key;" swaggerignore:"true"`
	IncludeAnime  bool  `gorm:"column:includeAnime;type:boolean;not null;" default:"true"`
	IncludeHentai bool  `gorm:"column:includeHentai;type:boolean;not null;" default:"false"`
}

func (MovieSettings) TableName() string {
	return "MovieSettings"
}

func (s *MovieSettings) Validate() string {
	return ""
}

//-----------------------------------------------------
//-----------------------------------------------------

type SettingName string

var SettingNames = []SettingName{AllSettingsName, DownloadSettingsName, NotificationSettingsName, MovieSettingsName}

const (
	AllSettingsName          SettingName = "all"
	DownloadSettingsName     SettingName = "downloadLinks"
	NotificationSettingsName SettingName = "notification"
	MovieSettingsName        SettingName = "movie"
)

//-----------------------------------------------------
//-----------------------------------------------------

type UserSettingsRes struct {
	DownloadLinksSettings *DownloadLinksSettings `gorm:"foreignKey:UserId;references:UserId;" json:"downloadLinksSettings"`
	NotificationSettings  *NotificationSettings  `gorm:"foreignKey:UserId;references:UserId;" json:"notificationSettings"`
	MovieSettings         *MovieSettings         `gorm:"foreignKey:UserId;references:UserId;" json:"movieSettings"`
}
