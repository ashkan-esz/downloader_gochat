package model

import "time"

type Movie struct {
	MovieId        string `gorm:"column:movieId;type:text;primaryKey;uniqueIndex:Movie_movieId_key;"`
	DislikesCount  int    `gorm:"column:dislikes_count;type:integer;default:0;not null;"`
	DroppedCount   int    `gorm:"column:dropped_count;type:integer;default:0;not null;"`
	FavoriteCount  int    `gorm:"column:favorite_count;type:integer;default:0;not null;"`
	FinishedCount  int    `gorm:"column:finished_count;type:integer;default:0;not null;"`
	FollowCount    int    `gorm:"column:follow_count;type:integer;default:0;not null;"`
	LikesCount     int    `gorm:"column:likes_count;type:integer;default:0;not null;"`
	WatchlistCount int    `gorm:"column:watchlist_count;type:integer;default:0;not null;"`
	ContinueCount  int    `gorm:"column:continue_count;type:integer;default:0;not null;"`
	ViewCount      int    `gorm:"column:view_count;type:integer;default:0;not null;"`
	ViewMonthCount int    `gorm:"column:view_month_count;type:integer;default:0;not null;"`
	//---------------------------------------
	//---------------------------------------
	Credits             []Credit              `gorm:"foreignKey:MovieId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RelatedMovies       []RelatedMovie        `gorm:"foreignKey:MovieId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RelatedMovies2      []RelatedMovie        `gorm:"foreignKey:RelatedId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FollowMovies        []FollowMovie         `gorm:"foreignKey:MovieId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LikeDislikeMovies   []LikeDislikeMovie    `gorm:"foreignKey:MovieId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	WatchedMovies       []WatchedMovie        `gorm:"foreignKey:MovieId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	WatchListMovies     []WatchListMovie      `gorm:"foreignKey:MovieId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserCollectionMovie []UserCollectionMovie `gorm:"foreignKey:MovieId;references:MovieId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Movie) TableName() string {
	return "Movie"
}

//---------------------------------------
//---------------------------------------

type Credit struct {
	Id             int64    `gorm:"column:id;type:serial;autoIncrement;primaryKey;uniqueIndex:Credit_id_key;"`
	MovieId        string   `gorm:"column:movieId;type:integer;not null;uniqueIndex:Credit_movieId_staffId_characterId_actorPositions_key;"`
	StaffId        int64    `gorm:"column:staffId;type:integer;not null;uniqueIndex:Credit_movieId_staffId_characterId_actorPositions_key;"`
	CharacterId    int64    `gorm:"column:characterId;type:integer;not null;uniqueIndex:Credit_movieId_staffId_characterId_actorPositions_key;"`
	ActorPositions []string `gorm:"column:actorPositions;type:text[];uniqueIndex:Credit_movieId_staffId_characterId_actorPositions_key;"`
	CharacterRole  string   `gorm:"column:characterRole;type:text;not null;"`
}

func (Credit) TableName() string {
	return "Credit"
}

//---------------------------------------
//---------------------------------------

type FollowMovie struct {
	MovieId      string    `gorm:"column:movieId;type:text;not null;primaryKey;index:FollowMovie_movieId_userId_idx;"`
	UserId       int64     `gorm:"column:userId;type:integer;not null;primaryKey;index:FollowMovie_movieId_userId_idx"`
	WatchEpisode int       `gorm:"column:watch_episode;type:integer;default:0;not null;"`
	WatchSeason  int       `gorm:"column:watch_season;type:integer;default:0;not null;"`
	Date         time.Time `gorm:"column:date;type:timestamp(3);not null;"`
	Score        float32   `gorm:"column:score;type:double precision;default:0;not null;"`
}

func (FollowMovie) TableName() string {
	return "FollowMovie"
}

//---------------------------------------
//---------------------------------------

type LikeDislikeMovie struct {
	MovieId string      `gorm:"column:movieId;type:text;not null;primaryKey;index:LikeDislikeMovie_movieId_userId_idx;"`
	UserId  int64       `gorm:"column:userId;type:integer;not null;primaryKey;index:LikeDislikeMovie_movieId_userId_idx"`
	Date    time.Time   `gorm:"column:date;type:timestamp(3);not null;"`
	Type    LikeDislike `gorm:"column:type;type:\"likeDislike\";not null;"`
}

func (LikeDislikeMovie) TableName() string {
	return "LikeDislikeMovie"
}

//---------------------------------------
//---------------------------------------

type WatchedMovie struct {
	MovieId      string    `gorm:"column:movieId;type:text;not null;primaryKey;index:WatchedMovie_movieId_userId_idx;"`
	UserId       int64     `gorm:"column:userId;type:integer;not null;primaryKey;index:WatchedMovie_movieId_userId_idx"`
	WatchEpisode int       `gorm:"column:watch_episode;type:integer;default:0;not null;"`
	WatchSeason  int       `gorm:"column:watch_season;type:integer;default:0;not null;"`
	Date         time.Time `gorm:"column:date;type:timestamp(3);not null;"`
	StartDate    time.Time `gorm:"column:startDate;type:timestamp(3);not null;"`
	Score        float32   `gorm:"column:score;type:double precision;default:0;not null;"`
	Dropped      bool      `gorm:"column:dropped;type:boolean;default:false;not null;"`
	Favorite     bool      `gorm:"column:favorite;type:boolean;default:false;not null;"`
}

func (WatchedMovie) TableName() string {
	return "WatchedMovie"
}

//---------------------------------------
//---------------------------------------

type MoviePoster struct {
	Url       string `json:"url" bson:"url"`
	Info      string `json:"info" bson:"info"`
	Size      int64  `json:"size" bson:"size"`
	VpnStatus string `json:"vpnStatus" bson:"vpnStatus"`
	Thumbnail string `json:"thumbnail" bson:"thumbnail"`
	BlurHash  string `json:"blurHash" bson:"blurHash"`
}

type MovieBriefData struct {
	MovieId  string        `bson:"_id" json:"movieId"`
	RawTitle string        `bson:"rawTitle" json:"rawTitle"`
	Type     string        `bson:"type" json:"type"`
	Year     string        `bson:"year" json:"year"`
	Posters  []MoviePoster `bson:"posters" json:"posters"`
}
