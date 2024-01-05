package model

import "time"

type WatchListGroup struct {
	GroupName string    `gorm:"column:group_name;type:text;default:\"default\";not null;primaryKey;"`
	UserId    int64     `gorm:"column:userId;type:integer;not null;primaryKey;"`
	Date      time.Time `gorm:"column:date;type:timestamp(3);not null;"`
	//---------------------------------------
	WatchListMovies []WatchListMovie `gorm:"foreignKey:UserId,GroupName;references:UserId,GroupName;constraint:OnUpdate:CASCADE,OnDelete:SET default;"`
}

func (WatchListGroup) TableName() string {
	return "WatchListGroup"
}

//---------------------------------------
//---------------------------------------

type WatchListMovie struct {
	MovieId   string    `gorm:"column:movieId;type:text;not null;primaryKey;index:WatchListMovie_movieId_userId_idx;"`
	UserId    int64     `gorm:"column:userId;type:integer;not null;primaryKey;index:WatchListMovie_movieId_userId_idx;"`
	GroupName string    `gorm:"column:group_name;type:text;default:\"default\";not null;"`
	Date      time.Time `gorm:"column:date;type:timestamp(3);not null;"`
	Score     float32   `gorm:"column:score;type:double precision;default:0;not null;"`
}

func (WatchListMovie) TableName() string {
	return "WatchListMovie"
}

//---------------------------------------
//---------------------------------------

type UserCollection struct {
	UserId         int64     `gorm:"column:userId;type:integer;not null;primaryKey;"`
	CollectionName string    `gorm:"column:collection_name;type:text;not null;primaryKey;"`
	Date           time.Time `gorm:"column:date;type:timestamp(3);not null;"`
	Public         bool      `gorm:"column:public;type:boolean;default:true;not null;"`
	Description    string    `gorm:"column:description;type:text;default:\"\";not null;"`
	//---------------------------------------
	UserCollectionMovie []UserCollectionMovie `gorm:"foreignKey:UserId,CollectionName;references:UserId,CollectionName;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (UserCollection) TableName() string {
	return "UserCollection"
}

//---------------------------------------
//---------------------------------------

type UserCollectionMovie struct {
	MovieId        string    `gorm:"column:movieId;type:text;not null;primaryKey;index:UserCollectionMovie_movieId_userId_collection_name_idx;"`
	UserId         int64     `gorm:"column:userId;type:integer;not null;primaryKey;index:UserCollectionMovie_movieId_userId_collection_name_idx;"`
	CollectionName string    `gorm:"column:collection_name;type:text;not null;primaryKey;index:UserCollectionMovie_movieId_userId_collection_name_idx;"`
	Date           time.Time `gorm:"column:date;type:timestamp(3);not null;"`
}

func (UserCollectionMovie) TableName() string {
	return "UserCollectionMovie"
}

//---------------------------------------
//---------------------------------------
